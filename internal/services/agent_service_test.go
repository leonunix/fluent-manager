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
	svc := NewAgentService(db)
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

	nodeID, err := svc.Register("uid-001", "web-01", "10.0.0.1", "linux", "1.0.0", "fluentbit", "2.0.0", `{"env":"prod"}`)
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

	nodeID, err := svc.Register("uid-001", "new-host", "10.0.0.2", "linux", "2.0.0", "fluentbit", "3.0.0", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var node models.Node
	db.First(&node, nodeID)
	if node.Hostname != "new-host" || node.Status != "online" {
		t.Errorf("existing node not updated: hostname=%s status=%s", node.Hostname, node.Status)
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

	nodeID, _ := svc.Register("uid-002", "web-01", "10.0.0.1", "linux", "1.0.0", "fluentbit", "2.0.0", "")

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

	resp, err := svc.Heartbeat("uid-001", "", nil)
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

	resp, _ := svc.Heartbeat("uid-001", "old-hash", nil)
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

	resp, _ := svc.Heartbeat("uid-001", models.HashConfig(content), nil)
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

	resp, _ := svc.Heartbeat("uid-001", "old-hash", nil)
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

	resp, _ := svc.Heartbeat("uid-001", "", nil)
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
	_, err := svc.Heartbeat("uid-001", "", metrics)
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
