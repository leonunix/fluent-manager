package executor

import (
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
)

func TestRewriteFluentBitConfigForManagedParsers(t *testing.T) {
	parserPath := "/etc/fluent-bit/conf.d/fluent-manager-parsers.conf"
	defaultParserPath := "/etc/fluent-bit/parsers.conf"
	content := `[SERVICE]
    Flush 1

[INPUT]
    Name tail
    Path /var/log/app.log
    Parser nginx_json

[PARSER]
    Name nginx_json
    Format json`

	mainConfig, parserConfig := rewriteFluentBitConfigForManagedParsers(content, parserPath, []string{defaultParserPath})

	if strings.Contains(mainConfig, "[PARSER]") {
		t.Fatalf("expected main config to exclude parser blocks, got %s", mainConfig)
	}
	if !strings.Contains(mainConfig, "Parsers_File "+parserPath) {
		t.Fatalf("expected main config to reference managed parser file, got %s", mainConfig)
	}
	if !strings.Contains(mainConfig, "Parsers_File "+defaultParserPath) {
		t.Fatalf("expected main config to reference default parser file, got %s", mainConfig)
	}
	if !strings.Contains(parserConfig, "[PARSER]") || !strings.Contains(parserConfig, "Name nginx_json") {
		t.Fatalf("expected parser config to contain parser block, got %s", parserConfig)
	}
}

func TestApplyFluentBitConfigExtractsParsersAndPersistsSourceHash(t *testing.T) {
	exec, parserPath := newApplyTestExecutor(t)
	startNamedRuntimeProcess(t, "fluent-bit")

	content := `[SERVICE]
    Flush 1

[INPUT]
    Name tail
    Path /var/log/app.log
    Parser nginx_json

[PARSER]
    Name nginx_json
    Format json`

	success, msg := exec.Apply(content, 42)
	if !success {
		t.Fatalf("expected apply success, got failure: %s", msg)
	}

	mainData, err := os.ReadFile(exec.cfg.FluentConfigPath)
	if err != nil {
		t.Fatalf("read main config: %v", err)
	}
	mainConfig := string(mainData)
	if strings.Contains(mainConfig, "[PARSER]") {
		t.Fatalf("expected parser blocks to be removed from main config, got %s", mainConfig)
	}
	if !strings.Contains(mainConfig, "Parsers_File "+parserPath) {
		t.Fatalf("expected managed parser reference in main config, got %s", mainConfig)
	}

	parserData, err := os.ReadFile(parserPath)
	if err != nil {
		t.Fatalf("read parser file: %v", err)
	}
	parserConfig := string(parserData)
	if !strings.Contains(parserConfig, "[PARSER]") || !strings.Contains(parserConfig, "Name nginx_json") {
		t.Fatalf("expected managed parser file to contain parser definitions, got %s", parserConfig)
	}

	if got := exec.CurrentConfigHash(); got != hashConfigContent(content) {
		t.Fatalf("expected current hash %s, got %s", hashConfigContent(content), got)
	}
}

func TestApplyFluentBitConfigRemovesManagedParserArtifactsWhenNoParserBlocksRemain(t *testing.T) {
	exec, parserPath := newApplyTestExecutor(t)
	startNamedRuntimeProcess(t, "fluent-bit")

	first := `[SERVICE]
    Flush 1

[INPUT]
    Name tail
    Path /var/log/app.log
    Parser nginx_json

[PARSER]
    Name nginx_json
    Format json`
	if success, msg := exec.Apply(first, 1); !success {
		t.Fatalf("seed apply failed: %s", msg)
	}

	second := `[SERVICE]
    Flush 5

[INPUT]
    Name cpu`
	if success, msg := exec.Apply(second, 2); !success {
		t.Fatalf("second apply failed: %s", msg)
	}

	mainData, err := os.ReadFile(exec.cfg.FluentConfigPath)
	if err != nil {
		t.Fatalf("read main config after cleanup apply: %v", err)
	}
	mainConfig := string(mainData)
	if strings.Contains(mainConfig, "Parsers_File "+parserPath) {
		t.Fatalf("expected managed parser reference to be removed, got %s", mainConfig)
	}
	if _, err := os.Stat(parserPath); !os.IsNotExist(err) {
		t.Fatalf("expected managed parser file to be removed, stat err=%v", err)
	}
	if got := exec.CurrentConfigHash(); got != hashConfigContent(second) {
		t.Fatalf("expected current hash %s, got %s", hashConfigContent(second), got)
	}
}

func TestApplyFluentBitConfigAddsDefaultParsersFileWhenOnlyParserReferenceExists(t *testing.T) {
	exec, parserPath := newApplyTestExecutor(t)
	startNamedRuntimeProcess(t, "fluent-bit")

	defaultParserPath := filepath.Join(filepath.Dir(exec.cfg.FluentConfigPath), "parsers.conf")
	if err := os.WriteFile(defaultParserPath, []byte("[PARSER]\n    Name docker\n    Format json\n"), 0o644); err != nil {
		t.Fatalf("write default parser file: %v", err)
	}

	content := `[INPUT]
    Name tail
    Path /var/log/app.log
    Parser docker`

	success, msg := exec.Apply(content, 99)
	if !success {
		t.Fatalf("expected apply success, got failure: %s", msg)
	}

	mainData, err := os.ReadFile(exec.cfg.FluentConfigPath)
	if err != nil {
		t.Fatalf("read main config: %v", err)
	}
	mainConfig := string(mainData)
	if !strings.Contains(mainConfig, "Parsers_File "+defaultParserPath) {
		t.Fatalf("expected main config to include default parsers file, got %s", mainConfig)
	}
	if _, err := os.Stat(parserPath); !os.IsNotExist(err) {
		t.Fatalf("expected no managed parser file for parser reference only, stat err=%v", err)
	}
}

func TestInjectFluentBitHTTPServerNoServiceBlock(t *testing.T) {
	content := `[INPUT]
    Name cpu`
	result := injectFluentBitHTTPServer(content, "2020")
	if !strings.Contains(result, "[SERVICE]") {
		t.Fatalf("expected [SERVICE] block to be injected, got:\n%s", result)
	}
	if !strings.Contains(result, "HTTP_Server On") {
		t.Fatalf("expected HTTP_Server On, got:\n%s", result)
	}
	if !strings.Contains(result, "HTTP_Port   2020") {
		t.Fatalf("expected HTTP_Port 2020, got:\n%s", result)
	}
	// INPUT block must still be present
	if !strings.Contains(result, "[INPUT]") {
		t.Fatalf("expected [INPUT] block to be preserved, got:\n%s", result)
	}
}

func TestInjectFluentBitHTTPServerExistingServiceNoHTTP(t *testing.T) {
	content := `[SERVICE]
    Flush 1

[INPUT]
    Name cpu`
	result := injectFluentBitHTTPServer(content, "2020")
	if !strings.Contains(result, "HTTP_Server On") {
		t.Fatalf("expected HTTP_Server On injected into SERVICE block, got:\n%s", result)
	}
	if !strings.Contains(result, "HTTP_Port   2020") {
		t.Fatalf("expected HTTP_Port injected, got:\n%s", result)
	}
	if strings.Count(result, "[SERVICE]") != 1 {
		t.Fatalf("expected exactly one [SERVICE] block, got:\n%s", result)
	}
}

func TestInjectFluentBitHTTPServerAlreadyEnabled(t *testing.T) {
	content := `[SERVICE]
    Flush 1
    HTTP_Server On
    HTTP_Port   2020

[INPUT]
    Name cpu`
	result := injectFluentBitHTTPServer(content, "2020")
	if result != content {
		t.Fatalf("expected content unchanged when HTTP_Server already On, got:\n%s", result)
	}
}

func TestInjectFluentBitHTTPServerOffReplacedWithOn(t *testing.T) {
	content := `[SERVICE]
    Flush 1
    HTTP_Server Off

[INPUT]
    Name cpu`
	result := injectFluentBitHTTPServer(content, "2020")
	if strings.Contains(result, "HTTP_Server Off") {
		t.Fatalf("expected HTTP_Server Off to be replaced, got:\n%s", result)
	}
	if !strings.Contains(result, "HTTP_Server On") {
		t.Fatalf("expected HTTP_Server On after replacement, got:\n%s", result)
	}
}

func TestInjectFluentBitHTTPServerPreservesExistingPort(t *testing.T) {
	content := `[SERVICE]
    Flush 1
    HTTP_Port   9999

[INPUT]
    Name cpu`
	result := injectFluentBitHTTPServer(content, "2020")
	if strings.Contains(result, "HTTP_Port   2020") {
		t.Fatalf("expected existing HTTP_Port 9999 to be preserved, got:\n%s", result)
	}
	if !strings.Contains(result, "HTTP_Port   9999") {
		t.Fatalf("expected HTTP_Port 9999 preserved, got:\n%s", result)
	}
}

func TestInjectFluentdMonitorAgentNotPresent(t *testing.T) {
	content := `<source>
  @type tail
  path /var/log/app.log
</source>`
	result := injectFluentdMonitorAgent(content, "24220")
	if !strings.Contains(result, "monitor_agent") {
		t.Fatalf("expected monitor_agent to be injected, got:\n%s", result)
	}
	if !strings.Contains(result, "port 24220") {
		t.Fatalf("expected port 24220 in injected block, got:\n%s", result)
	}
}

func TestInjectFluentdMonitorAgentAlreadyPresent(t *testing.T) {
	content := `<source>
  @type monitor_agent
  bind 0.0.0.0
  port 24220
</source>`
	result := injectFluentdMonitorAgent(content, "24220")
	if result != content {
		t.Fatalf("expected content unchanged when monitor_agent already present, got:\n%s", result)
	}
}

func TestMetricsPortFromURL(t *testing.T) {
	cases := []struct{ rawURL, fallback, want string }{
		{"http://127.0.0.1:2020/api/v1/metrics/prometheus", "2020", "2020"},
		{"http://127.0.0.1:24220/api/plugins.json", "24220", "24220"},
		{"http://127.0.0.1:9999/metrics", "2020", "9999"},
		{"", "2020", "2020"},
		{"not-a-url", "2020", "2020"},
	}
	for _, c := range cases {
		got := metricsPortFromURL(c.rawURL, c.fallback)
		if got != c.want {
			t.Errorf("metricsPortFromURL(%q, %q) = %q, want %q", c.rawURL, c.fallback, got, c.want)
		}
	}
}

func TestPrepareManagedConfigInjectsHTTPServer(t *testing.T) {
	exec, _ := newApplyTestExecutor(t)
	exec.cfg.FluentMetricsURL = "http://127.0.0.1:2020/api/v1/metrics/prometheus"

	content := `[SERVICE]
    Flush 1

[INPUT]
    Name cpu`

	prepared := exec.prepareManagedConfig(content)
	if !strings.Contains(prepared.mainContent, "HTTP_Server On") {
		t.Fatalf("expected HTTP_Server On in prepared config, got:\n%s", prepared.mainContent)
	}
	if prepared.sourceHash != hashConfigContent(content) {
		t.Fatalf("sourceHash should reflect original content, not injected content")
	}
}

func TestPrepareManagedConfigInjectsFluentdMonitorAgent(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/fluentd.conf"
	if err := os.WriteFile(configPath, []byte("<source>\n  @type tail\n</source>\n"), 0o644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	cfg := &config.Config{
		FluentType:       "fluentd",
		FluentConfigPath: configPath,
		FluentMetricsURL: "http://127.0.0.1:24220/api/plugins.json",
	}
	exec := New(cfg)

	content := "<source>\n  @type tail\n  path /var/log/app.log\n</source>"
	prepared := exec.prepareManagedConfig(content)
	if !strings.Contains(prepared.mainContent, "monitor_agent") {
		t.Fatalf("expected monitor_agent injected for fluentd, got:\n%s", prepared.mainContent)
	}
}

func newApplyTestExecutor(t *testing.T) (*Executor, string) {
	t.Helper()

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf.d")
	backupDir := filepath.Join(tempDir, "backups")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("mkdir backup dir: %v", err)
	}

	configPath := filepath.Join(tempDir, "fluent-bit.conf")
	if err := os.WriteFile(configPath, []byte("[SERVICE]\n    Flush 1\n"), 0o644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	cfg := &config.Config{
		FluentType:       "fluentbit",
		FluentConfigPath: configPath,
		FluentConfigDir:  configDir,
		FluentReloadCmd:  "true",
		FluentRestartCmd: "true",
		BackupDir:        backupDir,
		MaxBackups:       5,
		StateDir:         tempDir,
		RuntimeProfile: config.RuntimeProfile{
			SupportsHotReload: true,
		},
	}

	return New(cfg), filepath.Join(configDir, managedFluentBitParserFilename)
}

func startNamedRuntimeProcess(t *testing.T, processName string) {
	t.Helper()

	binDir := t.TempDir()
	runtimeBinary := filepath.Join(binDir, processName)
	if err := os.Symlink("/bin/sleep", runtimeBinary); err != nil {
		t.Fatalf("create runtime symlink: %v", err)
	}

	cmd := osexec.Command(runtimeBinary, "120")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start runtime process: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
}
