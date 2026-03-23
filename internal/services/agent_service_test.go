package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupAgentTest(t *testing.T) (*gorm.DB, AgentService) {
	t.Helper()
	db := testutil.NewTestDB()
	policySvc := NewAgentPolicyService(db, AgentSettings{})
	svc := NewAgentService(db, policySvc)
	return db, svc
}

func setupAgentTestWithSettings(t *testing.T, settings AgentSettings) (*gorm.DB, AgentService) {
	t.Helper()
	db := testutil.NewTestDB()
	policySvc := NewAgentPolicyService(db, settings)
	svc := NewAgentService(db, policySvc)
	return db, svc
}

func TestRegister_NewNode(t *testing.T) {
	db, svc := setupAgentTest(t)

	// Create default cluster for auto-assign
	dc := models.DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c := models.Cluster{Name: "default", RegionID: r.ID, IsDefault: true}
	db.Create(&c)

	nodeID, err := svc.Register("uid-001", "web-01", "10.0.0.1", "linux", "1.0.0", "fluentbit", "2.0.0", `{"env":"prod"}`, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nodeID == 0 {
		t.Error("expected non-zero node ID")
	}

	var node models.Node
	db.First(&node, nodeID)
	if node.Hostname != "web-01" || node.Status != "online" {
		t.Errorf("node not registered correctly: hostname=%s status=%s", node.Hostname, node.Status)
	}
	if node.ClusterID == nil || *node.ClusterID != c.ID {
		t.Error("node should be auto-assigned to default cluster")
	}
}

func TestRegister_ExistingNode(t *testing.T) {
	db, svc := setupAgentTest(t)

	db.Create(&models.Node{NodeUID: "uid-001", Hostname: "old-host", Status: "offline"})

	nodeID, err := svc.Register("uid-001", "new-host", "10.0.0.2", "linux", "2.0.0", "fluentbit", "3.0.0", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var node models.Node
	db.First(&node, nodeID)
	if node.Hostname != "new-host" || node.Status != "online" {
		t.Errorf("existing node not updated: hostname=%s status=%s", node.Hostname, node.Status)
	}
}

func TestRegister_PersistsFluentProfile(t *testing.T) {
	db, svc := setupAgentTest(t)

	dc := models.DataCenter{Name: "dc-profile"}
	db.Create(&dc)
	r := models.Region{Name: "r-profile", DataCenterID: dc.ID}
	db.Create(&r)
	c := models.Cluster{Name: "default-profile", RegionID: r.ID, IsDefault: true}
	db.Create(&c)

	profile := &AgentFluentProfileReport{
		LoadedPlugins:        "forward,tls",
		SupportsHotReload:    true,
		SupportsMultiline:    true,
		SupportsStorageLayer: true,
		SupportsForwardTLS:   true,
		SupportsMetricsAPI:   true,
		Metadata:             `{"runtime_type":"fluentbit"}`,
	}

	nodeID, err := svc.Register("uid-profile", "profile-node", "10.0.0.10", "linux", "1.0.0", "fluentbit", "3.0.0", "", profile)
	if err != nil {
		t.Fatalf("register with profile: %v", err)
	}

	var stored models.NodeFluentProfile
	if err := db.Where("node_id = ?", nodeID).First(&stored).Error; err != nil {
		t.Fatalf("load fluent profile: %v", err)
	}
	if stored.LoadedPlugins != "forward,tls" {
		t.Fatalf("expected loaded plugins to be persisted, got %q", stored.LoadedPlugins)
	}
	if !stored.SupportsMetricsAPI || !stored.SupportsForwardTLS {
		t.Fatalf("expected profile capabilities to be persisted: %#v", stored)
	}
	if stored.LastReportedAt == nil {
		t.Fatal("expected profile report timestamp to be set")
	}
}

func TestGetSettingsForNodeReturnsConfiguredPolicy(t *testing.T) {
	db, svc := setupAgentTestWithSettings(t, AgentSettings{
		HeartbeatInterval: 45,
		MetricsInterval:   90,
		FluentType:        "fluentd",
		FluentConfigPath:  "/etc/fluent/fluentd.conf",
	})

	node := models.Node{NodeUID: "uid-settings", Hostname: "settings-node"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	settings, err := svc.GetSettingsForNodeID(node.ID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.HeartbeatInterval != 45 || settings.MetricsInterval != 90 {
		t.Fatalf("unexpected interval settings: %#v", settings)
	}
	if settings.FluentType != "fluentd" || settings.FluentConfigPath != "/etc/fluent/fluentd.conf" {
		t.Fatalf("unexpected fluent settings: %#v", settings)
	}
}

func TestRegister_AutoAssignByMatchRule(t *testing.T) {
	db, svc := setupAgentTest(t)

	dc := models.DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	webCluster := models.Cluster{Name: "web-cluster", RegionID: r.ID}
	db.Create(&webCluster)

	db.Create(&models.ClusterMatchRule{
		ClusterID:       webCluster.ID,
		Name:            "web-nodes",
		Priority:        1,
		HostnamePattern: "web-*",
		IsActive:        true,
	})

	nodeID, _ := svc.Register("uid-002", "web-01", "10.0.0.1", "linux", "1.0.0", "fluentbit", "2.0.0", "", nil)

	var node models.Node
	db.First(&node, nodeID)
	if node.ClusterID == nil || *node.ClusterID != webCluster.ID {
		t.Error("node should be assigned to web-cluster by match rule")
	}
}

func TestHeartbeat_UpdatesStatus(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1", Status: "offline"}
	db.Create(&node)

	resp, err := svc.Heartbeat("uid-001", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}

	var updated models.Node
	db.First(&updated, node.ID)
	if updated.Status != "online" {
		t.Error("heartbeat should set status to online")
	}
}

func TestHeartbeat_ConfigUpdate(t *testing.T) {
	db, svc := setupAgentTest(t)

	cv := models.ConfigVersion{
		TemplateID: 0, Version: 1, Content: "new config",
		Hash: models.HashConfig("new config"),
	}
	db.Create(&cv)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1", ConfigID: &cv.ID}
	db.Create(&node)

	resp, _ := svc.Heartbeat("uid-001", "old-hash", nil, nil)
	if resp.Status != "update_config" {
		t.Errorf("expected update_config, got %s", resp.Status)
	}
	if resp.ConfigContent != "new config" {
		t.Error("should include new config content")
	}
}

func TestHeartbeat_ConfigUpToDate(t *testing.T) {
	db, svc := setupAgentTest(t)

	content := "current config"
	cv := models.ConfigVersion{
		TemplateID: 0, Version: 1, Content: content,
		Hash: models.HashConfig(content),
	}
	db.Create(&cv)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1", ConfigID: &cv.ID}
	db.Create(&node)

	resp, _ := svc.Heartbeat("uid-001", models.HashConfig(content), nil, nil)
	if resp.Status != "ok" {
		t.Errorf("expected ok (config up to date), got %s", resp.Status)
	}
}

func TestHeartbeat_ClusterConfigInheritance(t *testing.T) {
	db, svc := setupAgentTest(t)

	cv := models.ConfigVersion{
		TemplateID: 0, Version: 1, Content: "cluster config",
		Hash: models.HashConfig("cluster config"),
	}
	db.Create(&cv)

	dc := models.DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	cluster := models.Cluster{Name: "c1", RegionID: r.ID, ConfigID: &cv.ID}
	db.Create(&cluster)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1", ClusterID: &cluster.ID}
	db.Create(&node)

	resp, _ := svc.Heartbeat("uid-001", "old-hash", nil, nil)
	if resp.Status != "update_config" {
		t.Errorf("expected update_config from cluster inheritance, got %s", resp.Status)
	}
}

func TestHeartbeat_PendingCommands(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	db.Create(&models.RemoteCommand{NodeID: node.ID, Action: "restart", Status: "pending"})
	db.Create(&models.RemoteCommand{NodeID: node.ID, Action: "status", Status: "pending"})

	resp, _ := svc.Heartbeat("uid-001", "", nil, nil)
	if len(resp.Commands) != 2 {
		t.Fatalf("expected 2 pending commands, got %d", len(resp.Commands))
	}

	// Verify commands marked as delivered
	var cmd models.RemoteCommand
	db.First(&cmd, resp.Commands[0].ID)
	if cmd.Status != "delivered" {
		t.Errorf("command should be marked as delivered, got %s", cmd.Status)
	}
}

func TestHeartbeat_StoresMetrics(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	metrics := map[string]interface{}{
		"cpu_usage_percent": 45.5,
		"mem_total_mb":      float64(8192),
		"mem_used_mb":       float64(4096),
		"mem_usage_percent": 50.0,
		"fluent_running":    true,
		"fluent_pid":        float64(1234),
	}

	// Test storeMetrics directly via Heartbeat
	_, err := svc.Heartbeat("uid-001", "", metrics, nil)
	if err != nil {
		t.Fatalf("heartbeat error: %v", err)
	}

	// Verify metrics stored - use raw SQL to bypass any GORM caching
	var cpuVal float64
	var fluentVal bool
	row := db.Raw("SELECT cpu_usage_percent, fluent_running FROM node_metrics WHERE node_id = ?", node.ID).Row()
	if err := row.Scan(&cpuVal, &fluentVal); err != nil {
		t.Fatalf("raw query error: %v", err)
	}
	if cpuVal != 45.5 {
		t.Errorf("expected CPU 45.5, got %f", cpuVal)
	}
	if !fluentVal {
		t.Error("expected fluent_running = true")
	}
}

func TestHeartbeat_UpdatesRuntimeStateHashes(t *testing.T) {
	db, svc := setupAgentTest(t)

	cv := models.ConfigVersion{
		TemplateID: 0,
		Version:    1,
		Content:    "desired config",
		Hash:       "desired-hash",
	}
	db.Create(&cv)

	dc := models.DataCenter{Name: "dc-hash"}
	db.Create(&dc)
	region := models.Region{Name: "region-hash", DataCenterID: dc.ID}
	db.Create(&region)
	cluster := models.Cluster{Name: "cluster-hash", RegionID: region.ID, ConfigID: &cv.ID}
	db.Create(&cluster)

	node := models.Node{NodeUID: "uid-runtime", Hostname: "runtime-node", ClusterID: &cluster.ID}
	db.Create(&node)

	_, err := svc.Heartbeat("uid-runtime", "effective-hash", map[string]interface{}{
		"queue_depth":      float64(25),
		"retry_count":      float64(3),
		"flush_latency_ms": float64(120),
		"input_status":     "healthy",
		"output_status":    "healthy",
	}, nil)
	if err != nil {
		t.Fatalf("heartbeat error: %v", err)
	}

	var state models.NodeRuntimeState
	if err := db.Where("node_id = ?", node.ID).First(&state).Error; err != nil {
		t.Fatalf("load runtime state: %v", err)
	}
	if state.DesiredConfigHash != "desired-hash" {
		t.Fatalf("expected desired hash desired-hash, got %s", state.DesiredConfigHash)
	}
	if state.EffectiveConfigHash != "effective-hash" {
		t.Fatalf("expected effective hash effective-hash, got %s", state.EffectiveConfigHash)
	}
	if state.QueueDepth != 25 || state.RetryCount != 3 || state.FlushLatencyMS != 120 {
		t.Fatalf("unexpected runtime metrics: %#v", state)
	}
	if state.LastSyncAt == nil {
		t.Fatal("expected last sync timestamp to be set")
	}
}

func TestHeartbeat_UpdatesFluentProfile(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-heartbeat-profile", Hostname: "profile-heartbeat"}
	db.Create(&node)

	initial := models.NodeFluentProfile{
		NodeID:             node.ID,
		LoadedPlugins:      "old",
		SupportsHotReload:  false,
		SupportsMetricsAPI: false,
	}
	if err := db.Create(&initial).Error; err != nil {
		t.Fatalf("create initial profile: %v", err)
	}

	profile := &AgentFluentProfileReport{
		LoadedPlugins:        "tail,forward",
		SupportsHotReload:    true,
		SupportsMultiline:    true,
		SupportsStorageLayer: true,
		SupportsForwardTLS:   true,
		SupportsMetricsAPI:   true,
		Metadata:             `{"runtime_type":"fluentd"}`,
	}
	if _, err := svc.Heartbeat("uid-heartbeat-profile", "hash-1", map[string]interface{}{
		"queue_depth":   float64(2),
		"input_status":  "healthy",
		"output_status": "degraded",
	}, profile); err != nil {
		t.Fatalf("heartbeat with profile: %v", err)
	}

	var stored models.NodeFluentProfile
	if err := db.Where("node_id = ?", node.ID).First(&stored).Error; err != nil {
		t.Fatalf("load updated profile: %v", err)
	}
	if stored.LoadedPlugins != "tail,forward" {
		t.Fatalf("expected loaded plugins to update, got %q", stored.LoadedPlugins)
	}
	if !stored.SupportsHotReload || !stored.SupportsMetricsAPI {
		t.Fatalf("expected updated fluent profile capabilities, got %#v", stored)
	}
	if stored.Metadata != `{"runtime_type":"fluentd"}` {
		t.Fatalf("expected metadata to update, got %q", stored.Metadata)
	}
	if stored.LastReportedAt == nil {
		t.Fatal("expected heartbeat profile timestamp to be set")
	}
}

func TestHeartbeat_IncludesAgentSettings(t *testing.T) {
	db, svc := setupAgentTestWithSettings(t, AgentSettings{
		HeartbeatInterval: 15,
		LogUploadInterval: 180,
		FluentLogPath:     "/var/log/custom.log",
	})

	node := models.Node{NodeUID: "uid-heartbeat-settings", Hostname: "settings-node"}
	db.Create(&node)

	resp, err := svc.Heartbeat("uid-heartbeat-settings", "", nil, nil)
	if err != nil {
		t.Fatalf("heartbeat error: %v", err)
	}
	if resp.AgentSettings == nil {
		t.Fatal("expected heartbeat to include agent settings")
	}
	if resp.AgentSettings.HeartbeatInterval != 15 || resp.AgentSettings.LogUploadInterval != 180 {
		t.Fatalf("unexpected heartbeat settings: %#v", resp.AgentSettings)
	}
	if resp.AgentSettings.FluentLogPath != "/var/log/custom.log" {
		t.Fatalf("expected fluent log path in settings, got %#v", resp.AgentSettings)
	}
}

func TestReportCommandResult(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)
	cmd := models.RemoteCommand{NodeID: node.ID, Action: "restart", Status: "delivered"}
	db.Create(&cmd)

	err := svc.ReportCommandResult(cmd.ID, "completed", "service restarted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var updated models.RemoteCommand
	db.First(&updated, cmd.ID)
	if updated.Status != "completed" || updated.Output != "service restarted" {
		t.Error("command result not stored correctly")
	}
}

func TestReportStatusUpdatesRuntimeState(t *testing.T) {
	db, svc := setupAgentTest(t)

	version := models.ConfigVersion{
		TemplateID: 0,
		Version:    1,
		Content:    "deploy config",
		Hash:       "deploy-hash",
	}
	db.Create(&version)

	node := models.Node{NodeUID: "uid-report", Hostname: "report-node", Status: "online"}
	db.Create(&node)

	if err := svc.ReportStatus("uid-report", version.ID, false, "syntax error"); err != nil {
		t.Fatalf("report status failure path: %v", err)
	}

	var state models.NodeRuntimeState
	if err := db.Where("node_id = ?", node.ID).First(&state).Error; err != nil {
		t.Fatalf("load runtime state after failure: %v", err)
	}
	if state.DesiredConfigHash != "deploy-hash" {
		t.Fatalf("expected desired hash deploy-hash, got %s", state.DesiredConfigHash)
	}
	if state.LastError != "syntax error" {
		t.Fatalf("expected last error syntax error, got %s", state.LastError)
	}
	if state.LastReloadAt == nil {
		t.Fatal("expected last reload time to be set after failure")
	}

	if err := svc.ReportStatus("uid-report", version.ID, true, "ok"); err != nil {
		t.Fatalf("report status success path: %v", err)
	}
	if err := db.Where("node_id = ?", node.ID).First(&state).Error; err != nil {
		t.Fatalf("reload runtime state after success: %v", err)
	}
	if state.LastError != "" {
		t.Fatalf("expected last error to be cleared after success, got %s", state.LastError)
	}
}

func TestUploadLogs(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	err := svc.UploadLogs("uid-001", []string{"line1", "line2", "line3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var log models.NodeLog
	db.Where("node_id = ?", node.ID).First(&log)
	if log.LineCount != 3 {
		t.Errorf("expected 3 lines, got %d", log.LineCount)
	}
}

func TestSendCommand(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	user := models.User{Username: "admin", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)

	cmd, err := svc.SendCommand("1", user.ID, "restart", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Action != "restart" || cmd.Status != "pending" {
		t.Error("command not created correctly")
	}
}

func TestGetNodeMetrics(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	db.Create(&models.NodeMetrics{NodeID: node.ID, CPUUsagePercent: 55.0})

	m, err := svc.GetNodeMetrics("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.CPUUsagePercent != 55.0 {
		t.Errorf("expected CPU 55.0, got %f", m.CPUUsagePercent)
	}
}

func TestGetNodeLogs(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	for i := 0; i < 25; i++ {
		db.Create(&models.NodeLog{NodeID: node.ID, Lines: "log line", LineCount: 1})
	}

	logs, err := svc.GetNodeLogs("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 20 {
		t.Errorf("expected 20 logs (limit), got %d", len(logs))
	}
}

func TestListNodeCommands(t *testing.T) {
	db, svc := setupAgentTest(t)

	node := models.Node{NodeUID: "uid-001", Hostname: "h1"}
	db.Create(&node)

	db.Create(&models.RemoteCommand{NodeID: node.ID, Action: "restart", Status: "completed"})
	db.Create(&models.RemoteCommand{NodeID: node.ID, Action: "status", Status: "pending"})

	cmds, err := svc.ListNodeCommands("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cmds))
	}
}
