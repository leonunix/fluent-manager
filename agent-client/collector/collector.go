package collector

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
)

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

	// Metadata
	CollectedAt time.Time `json:"collected_at"`
}

// Collector periodically collects system and fluent metrics.
type Collector struct {
	cfg    *config.Config
	mu     sync.RWMutex
	latest Metrics
	stopCh chan struct{}
}

func New(cfg *config.Config) *Collector {
	return &Collector{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

func (c *Collector) Start() {
	// Collect once immediately
	c.collect()

	go func() {
		ticker := time.NewTicker(time.Duration(c.cfg.MetricsInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
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
	m := Metrics{
		CollectedAt: time.Now(),
		NumCPUs:     runtime.NumCPU(),
	}

	c.collectCPU(&m)
	c.collectMemory(&m)
	c.collectDisk(&m)
	c.collectLoadAvg(&m)
	c.collectFluentProcess(&m)

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
			idle = values[3] // idle is the 4th field
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
func (c *Collector) collectFluentProcess(m *Metrics) {
	// Find PID via pgrep
	procName := "fluent-bit"
	if c.cfg.FluentType == "fluentd" {
		procName = "fluentd"
	}

	out, err := exec.Command("pgrep", "-x", procName).Output()
	if err != nil {
		// Also try with the full binary name
		out, err = exec.Command("pgrep", "-f", c.cfg.FluentBinary).Output()
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
		// Read /proc/PID/stat for CPU
		c.readProcStat(m.FluentPID, m)
		// Read /proc/PID/status for memory
		c.readProcStatus(m.FluentPID, m)
		// Count open file descriptors
		c.countFDs(m.FluentPID, m)
	}
}

func (c *Collector) readProcStat(pid int, m *Metrics) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return
	}
	// Fields after the last ")" are space-separated; fields 14 and 15 are utime and stime
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

	// Read uptime to calculate CPU percent
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return
	}
	uptimeFields := strings.Fields(string(uptimeData))
	if len(uptimeFields) < 1 {
		return
	}
	systemUptime, _ := strconv.ParseFloat(uptimeFields[0], 64)

	// starttime is field 19 (index 19 from the post-) section)
	if len(fields) > 19 {
		starttime, _ := strconv.ParseFloat(fields[19], 64)
		// All times are in clock ticks (usually 100 per second)
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
