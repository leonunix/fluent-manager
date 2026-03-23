package collector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
)

var prometheusMetricLine = regexp.MustCompile(`^([a-zA-Z_:][a-zA-Z0-9_:]*)(?:\{[^}]*\})?\s+([-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?)$`)

// Metrics holds a snapshot of system and fluent process metrics.
type Metrics struct {
	// System
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`
	MemTotalMB       uint64  `json:"mem_total_mb"`
	MemUsedMB        uint64  `json:"mem_used_mb"`
	MemUsagePercent  float64 `json:"mem_usage_percent"`
	DiskTotalGB      uint64  `json:"disk_total_gb"`
	DiskUsedGB       uint64  `json:"disk_used_gb"`
	DiskUsagePercent float64 `json:"disk_usage_percent"`
	LoadAvg1         float64 `json:"load_avg_1"`
	LoadAvg5         float64 `json:"load_avg_5"`
	LoadAvg15        float64 `json:"load_avg_15"`
	NumCPUs          int     `json:"num_cpus"`

	// Fluent process
	FluentRunning    bool    `json:"fluent_running"`
	FluentPID        int     `json:"fluent_pid"`
	FluentCPUPercent float64 `json:"fluent_cpu_percent"`
	FluentMemMB      float64 `json:"fluent_mem_mb"`
	FluentOpenFDs    int     `json:"fluent_open_fds"`

	// Runtime signals
	QueueDepth     int    `json:"queue_depth"`
	RetryCount     int    `json:"retry_count"`
	FlushLatencyMS int    `json:"flush_latency_ms"`
	InputStatus    string `json:"input_status"`
	OutputStatus   string `json:"output_status"`

	// Metadata
	CollectedAt time.Time `json:"collected_at"`
}

// Collector periodically collects system and fluent metrics.
type Collector struct {
	cfg        *config.Config
	httpClient *http.Client
	mu         sync.RWMutex
	latest     Metrics
	stopCh     chan struct{}
}

func New(cfg *config.Config) *Collector {
	return &Collector{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

func (c *Collector) Start() {
	// Collect once immediately
	c.collect()

	go func() {
		for {
			wait := time.Duration(c.cfg.Snapshot().MetricsInterval) * time.Second
			if wait <= 0 {
				wait = 60 * time.Second
			}
			select {
			case <-time.After(wait):
				c.collect()
			case <-c.stopCh:
				return
			}
		}
	}()
}

func (c *Collector) Stop() {
	close(c.stopCh)
}

// Snapshot returns the most recent metrics.
func (c *Collector) Snapshot() Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latest
}

func (c *Collector) collect() {
	cfg := c.cfg.Snapshot()
	m := Metrics{
		CollectedAt:  time.Now(),
		NumCPUs:      runtime.NumCPU(),
		InputStatus:  "unknown",
		OutputStatus: "unknown",
	}

	c.collectCPU(&m)
	c.collectMemory(&m)
	c.collectDisk(&m)
	c.collectLoadAvg(&m)
	c.collectFluentProcess(cfg, &m)
	c.collectRuntimeSignals(cfg, &m)

	c.mu.Lock()
	c.latest = m
	c.mu.Unlock()
}

// collectCPU reads /proc/stat for CPU usage.
func (c *Collector) collectCPU(m *Metrics) {
	if runtime.GOOS != "linux" {
		return
	}

	read := func() (idle, total uint64) {
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return
		}
		lines := strings.Split(string(data), "\n")
		if len(lines) == 0 {
			return
		}
		fields := strings.Fields(lines[0])
		if len(fields) < 5 || fields[0] != "cpu" {
			return
		}
		var values []uint64
		for _, f := range fields[1:] {
			v, _ := strconv.ParseUint(f, 10, 64)
			values = append(values, v)
		}
		for _, v := range values {
			total += v
		}
		if len(values) >= 4 {
			idle = values[3]
		}
		return
	}

	idle1, total1 := read()
	time.Sleep(500 * time.Millisecond)
	idle2, total2 := read()

	totalDelta := float64(total2 - total1)
	idleDelta := float64(idle2 - idle1)
	if totalDelta > 0 {
		m.CPUUsagePercent = (1.0 - idleDelta/totalDelta) * 100
	}
}

// collectMemory reads /proc/meminfo.
func (c *Collector) collectMemory(m *Metrics) {
	if runtime.GOOS != "linux" {
		return
	}

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}

	info := map[string]uint64{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.ParseUint(parts[1], 10, 64)
			info[key] = val
		}
	}

	m.MemTotalMB = info["MemTotal"] / 1024
	available := info["MemAvailable"]
	if available == 0 {
		available = info["MemFree"] + info["Buffers"] + info["Cached"]
	}
	m.MemUsedMB = (info["MemTotal"] - available) / 1024
	if info["MemTotal"] > 0 {
		m.MemUsagePercent = float64(info["MemTotal"]-available) / float64(info["MemTotal"]) * 100
	}
}

// collectDisk gets disk usage for the root filesystem.
func (c *Collector) collectDisk(m *Metrics) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return
	}
	m.DiskTotalGB = stat.Blocks * uint64(stat.Bsize) / (1024 * 1024 * 1024)
	m.DiskUsedGB = (stat.Blocks - stat.Bfree) * uint64(stat.Bsize) / (1024 * 1024 * 1024)
	if stat.Blocks > 0 {
		m.DiskUsagePercent = float64(stat.Blocks-stat.Bfree) / float64(stat.Blocks) * 100
	}
}

// collectLoadAvg reads /proc/loadavg.
func (c *Collector) collectLoadAvg(m *Metrics) {
	if runtime.GOOS != "linux" {
		return
	}
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}
	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		m.LoadAvg1, _ = strconv.ParseFloat(fields[0], 64)
		m.LoadAvg5, _ = strconv.ParseFloat(fields[1], 64)
		m.LoadAvg15, _ = strconv.ParseFloat(fields[2], 64)
	}
}

// collectFluentProcess finds the fluent process and reads its resource usage.
func (c *Collector) collectFluentProcess(cfg config.Snapshot, m *Metrics) {
	procName := "fluent-bit"
	if cfg.FluentType == "fluentd" {
		procName = "fluentd"
	}

	out, err := exec.Command("pgrep", "-x", procName).Output()
	if err != nil {
		out, err = exec.Command("pgrep", "-f", cfg.FluentBinary).Output()
		if err != nil {
			m.FluentRunning = false
			return
		}
	}

	pids := strings.Fields(strings.TrimSpace(string(out)))
	if len(pids) == 0 {
		m.FluentRunning = false
		return
	}

	m.FluentRunning = true
	m.FluentPID, _ = strconv.Atoi(pids[0])

	if m.FluentPID > 0 && runtime.GOOS == "linux" {
		c.readProcStat(m.FluentPID, m)
		c.readProcStatus(m.FluentPID, m)
		c.countFDs(m.FluentPID, m)
	}
}

func (c *Collector) collectRuntimeSignals(cfg config.Snapshot, m *Metrics) {
	if !m.FluentRunning {
		m.InputStatus = "unhealthy"
		m.OutputStatus = "unhealthy"
		return
	}

	m.InputStatus = "healthy"
	m.OutputStatus = "healthy"

	if !cfg.RuntimeProfile.SupportsMetricsAPI || strings.TrimSpace(cfg.FluentMetricsURL) == "" {
		return
	}

	payload, err := c.fetchRuntimeMetrics(cfg)
	if err != nil {
		log.Printf("[collector] runtime metrics scrape failed: %v", err)
		return
	}

	switch normalizeMetricsFormat(cfg.FluentMetricsFormat, cfg.FluentType) {
	case "fluentd_monitor_agent":
		c.populateFromFluentdMonitorAgent(payload, m)
	default:
		c.populateFromPrometheus(payload, m)
	}
}

func (c *Collector) fetchRuntimeMetrics(cfg config.Snapshot) ([]byte, error) {
	resp, err := c.httpClient.Get(cfg.FluentMetricsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("metrics endpoint returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Collector) populateFromPrometheus(payload []byte, m *Metrics) {
	sums := map[string]float64{}
	scanner := bufio.NewScanner(strings.NewReader(string(payload)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		match := prometheusMetricLine.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		value, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}
		sums[match[1]] += value
	}

	m.QueueDepth = int(firstMetricValue(sums,
		"fluentbit_storage_backlog_chunks",
		"fluentbit_output_queue_chunks",
		"fluentbit_output_queue_records",
		"fluentbit_output_queue_bytes",
	))
	m.RetryCount = int(firstMetricValue(sums,
		"fluentbit_output_retries_total",
		"fluentbit_output_retried_records_total",
		"fluentbit_output_retries",
	))

	latencyMS := firstMetricValue(sums,
		"fluentbit_output_latency_ms",
		"fluentbit_output_proc_latency_ms",
	)
	if latencyMS == 0 {
		latencySeconds := firstMetricValue(sums,
			"fluentbit_output_latency_seconds",
			"fluentbit_output_proc_latency_seconds",
		)
		latencyMS = latencySeconds * 1000
	}
	m.FlushLatencyMS = int(latencyMS)

	inputHardErrors := metricSum(sums,
		"fluentbit_input_errors_total",
		"fluentbit_input_dropped_records_total",
	)
	outputHardErrors := metricSum(sums,
		"fluentbit_output_errors_total",
		"fluentbit_output_retries_failed_total",
		"fluentbit_output_dropped_records_total",
	)

	m.InputStatus = deriveStatus(m.FluentRunning, inputHardErrors, 0, 0)
	m.OutputStatus = deriveStatus(m.FluentRunning, outputHardErrors, float64(m.RetryCount), float64(m.QueueDepth))
}

func (c *Collector) populateFromFluentdMonitorAgent(payload []byte, m *Metrics) {
	plugins, err := decodeMonitorAgentPlugins(payload)
	if err != nil {
		log.Printf("[collector] fluentd monitor_agent decode failed: %v", err)
		return
	}

	var (
		totalQueue    float64
		totalRetries  float64
		totalFlushMS  float64
		inputErrors   float64
		outputErrors  float64
		outputSignals float64
	)

	for _, plugin := range plugins {
		category := strings.ToLower(stringField(plugin, "plugin_category"))
		if category == "" {
			category = strings.ToLower(stringField(plugin, "type"))
		}
		queueLen := numberField(plugin, "buffer_queue_length")
		retries := numberField(plugin, "retry_count")
		errors := numberField(plugin, "num_errors")
		slowFlushes := numberField(plugin, "slow_flush_count")
		flushTimeCount := numberField(plugin, "flush_time_count")

		totalQueue += queueLen
		totalRetries += retries
		totalFlushMS += slowFlushes*1000 + flushTimeCount*1000

		if category == "input" || strings.EqualFold(stringField(plugin, "plugin_type"), "input") {
			inputErrors += errors
			continue
		}

		if category == "output" || boolField(plugin, "output_plugin") {
			outputErrors += errors
			outputSignals += queueLen + retries
		}
	}

	m.QueueDepth = int(totalQueue)
	m.RetryCount = int(totalRetries)
	m.FlushLatencyMS = int(totalFlushMS)
	m.InputStatus = deriveStatus(m.FluentRunning, inputErrors, 0, 0)
	m.OutputStatus = deriveStatus(m.FluentRunning, outputErrors, totalRetries, outputSignals)
}

func decodeMonitorAgentPlugins(payload []byte) ([]map[string]interface{}, error) {
	var list []map[string]interface{}
	if err := json.Unmarshal(payload, &list); err == nil {
		return list, nil
	}

	var wrapped struct {
		Plugins []map[string]interface{} `json:"plugins"`
	}
	if err := json.Unmarshal(payload, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Plugins, nil
}

func normalizeMetricsFormat(format, runtimeType string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "prometheus":
		if runtimeType == "fluentd" {
			return "fluentd_monitor_agent"
		}
		return "prometheus"
	case "fluentd", "monitor_agent", "fluentd_monitor_agent":
		return "fluentd_monitor_agent"
	default:
		return strings.ToLower(strings.TrimSpace(format))
	}
}

func metricSum(sums map[string]float64, names ...string) float64 {
	var total float64
	for _, name := range names {
		total += sums[name]
	}
	return total
}

func firstMetricValue(sums map[string]float64, names ...string) float64 {
	for _, name := range names {
		if value := sums[name]; value > 0 {
			return value
		}
	}
	return 0
}

func deriveStatus(running bool, hardErrors, retries, queue float64) string {
	if !running {
		return "unhealthy"
	}
	if hardErrors > 0 {
		return "degraded"
	}
	if retries > 0 || queue > 0 {
		return "degraded"
	}
	return "healthy"
}

func stringField(values map[string]interface{}, key string) string {
	raw, ok := values[key]
	if !ok || raw == nil {
		return ""
	}
	if str, ok := raw.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", raw)
}

func numberField(values map[string]interface{}, key string) float64 {
	raw, ok := values[key]
	if !ok || raw == nil {
		return 0
	}

	switch value := raw.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case json.Number:
		parsed, _ := value.Float64()
		return parsed
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed
	default:
		return 0
	}
}

func boolField(values map[string]interface{}, key string) bool {
	raw, ok := values[key]
	if !ok || raw == nil {
		return false
	}
	if value, ok := raw.(bool); ok {
		return value
	}
	if value, ok := raw.(string); ok {
		parsed, _ := strconv.ParseBool(strings.TrimSpace(value))
		return parsed
	}
	return false
}

func (c *Collector) readProcStat(pid int, m *Metrics) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return
	}
	closeParen := strings.LastIndex(string(data), ")")
	if closeParen < 0 {
		return
	}
	fields := strings.Fields(string(data)[closeParen+2:])
	if len(fields) < 13 {
		return
	}
	utime, _ := strconv.ParseFloat(fields[11], 64)
	stime, _ := strconv.ParseFloat(fields[12], 64)
	total := utime + stime

	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return
	}
	uptimeFields := strings.Fields(string(uptimeData))
	if len(uptimeFields) < 1 {
		return
	}
	systemUptime, _ := strconv.ParseFloat(uptimeFields[0], 64)

	if len(fields) > 19 {
		starttime, _ := strconv.ParseFloat(fields[19], 64)
		hertz := 100.0
		elapsed := systemUptime - (starttime / hertz)
		if elapsed > 0 {
			m.FluentCPUPercent = (total / hertz / elapsed) * 100
		}
	}
}

func (c *Collector) readProcStatus(pid int, m *Metrics) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseFloat(fields[1], 64)
				m.FluentMemMB = kb / 1024
			}
			break
		}
	}
}

func (c *Collector) countFDs(pid int, m *Metrics) {
	entries, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	if err != nil {
		return
	}
	m.FluentOpenFDs = len(entries)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
