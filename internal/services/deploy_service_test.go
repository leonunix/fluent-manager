package services

import (
	"errors"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupDeployTest(t *testing.T) (*gorm.DB, DeployService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewDeployService(db, NewFluentOpsService(db))
	return db, svc
}

func seedDeployData(t *testing.T, db *gorm.DB) (cv *models.ConfigVersion, cluster *models.Cluster, nodes []models.Node) {
	t.Helper()
	dc := models.DataCenter{Name: "dc1", Provider: "aws"}
	db.Create(&dc)
	r := models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c := models.Cluster{Name: "c1", RegionID: r.ID}
	db.Create(&c)

	tpl := models.ConfigTemplate{Name: "tpl1", FluentType: "fluentbit", Content: "config"}
	db.Create(&tpl)
	v := models.ConfigVersion{
		TemplateID: tpl.ID, Version: 1, Content: "config v1",
		Hash: models.HashConfig("config v1"),
	}
	db.Create(&v)

	n1 := models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &c.ID, Status: "online"}
	n2 := models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &c.ID, Status: "online"}
	db.Create(&n1)
	db.Create(&n2)

	return &v, &c, []models.Node{n1, n2}
}

func TestDeployCreate_ByNodeIDs(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, _, nodes := seedDeployData(t, db)

	task, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.TotalNodes != 1 || task.Scope != "node" {
		t.Errorf("expected 1 node with scope=node, got %d scope=%s", task.TotalNodes, task.Scope)
	}
}

func TestDeployCreate_ByCluster(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	task, err := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.TotalNodes != 2 || task.Scope != "cluster" {
		t.Errorf("expected 2 nodes with scope=cluster, got %d scope=%s", task.TotalNodes, task.Scope)
	}
}

func TestDeployCreate_ByRegion(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, _, _ := seedDeployData(t, db)

	var region models.Region
	db.First(&region)

	task, err := svc.Create(cv.ID, nil, nil, &region.ID, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Scope != "region" || task.TotalNodes != 2 {
		t.Errorf("expected scope=region with 2 nodes, got scope=%s nodes=%d", task.Scope, task.TotalNodes)
	}
}

func TestDeployCreate_ByDatacenter(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, _, _ := seedDeployData(t, db)

	var dc models.DataCenter
	db.First(&dc)

	task, err := svc.Create(cv.ID, nil, nil, nil, &dc.ID, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Scope != "datacenter" || task.TotalNodes != 2 {
		t.Errorf("expected scope=datacenter with 2 nodes, got scope=%s nodes=%d", task.Scope, task.TotalNodes)
	}
}

func TestDeployCreate_NoTargets(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, _, _ := seedDeployData(t, db)
	_ = db

	_, err := svc.Create(cv.ID, nil, nil, nil, nil, nil, 1, nil, false)
	if err == nil {
		t.Error("expected error when no target nodes")
	}
}

func TestDeployCreate_InvalidConfigVersion(t *testing.T) {
	_, svc := setupDeployTest(t)

	_, err := svc.Create(999, []uint{1}, nil, nil, nil, nil, 1, nil, false)
	if err == nil {
		t.Error("expected error for non-existent config version")
	}
}

func TestDeployCreate_ScopeCheck(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, _, nodes := seedDeployData(t, db)

	// Create second cluster and scope user to it
	c2 := models.Cluster{Name: "c2", RegionID: 1}
	db.Create(&c2)

	// Nodes are in c1 but user is scoped to c2 only
	_, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, []uint{c2.ID}, false)
	if err == nil {
		t.Error("expected scope violation error")
	}
}

func TestDeployCreate_ScopeCheckPasses(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, nodes := seedDeployData(t, db)

	// User scoped to c1 where nodes are
	task, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, []uint{cluster.ID}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.TotalNodes != 1 {
		t.Errorf("expected 1 node, got %d", task.TotalNodes)
	}
}

func TestDeployCreate_Dedup(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, nodes := seedDeployData(t, db)

	// Both cluster scope and explicit node IDs target same nodes
	task, err := svc.Create(cv.ID, []uint{nodes[0].ID, nodes[1].ID}, &cluster.ID, nil, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.TotalNodes != 2 {
		t.Errorf("expected 2 after dedup, got %d", task.TotalNodes)
	}
}

func TestDeployList_Global(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)

	tasks, total, err := svc.List(1, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(tasks) != 1 {
		t.Errorf("expected 1 task, got total=%d len=%d", total, len(tasks))
	}
}

func TestDeployList_Scoped(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)

	// User scoped to non-existent cluster
	tasks, total, _ := svc.List(1, 10, []uint{999})
	if total != 0 || len(tasks) != 0 {
		t.Error("scoped user should see no tasks outside their scope")
	}
}

func TestDeployGet(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	task, _ := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)

	gotTask, records, total, err := svc.Get(task.ID, 1, 20, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTask.ID != task.ID {
		t.Error("wrong task returned")
	}
	if total != 2 {
		t.Errorf("expected total=2 deploy records, got %d", total)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 deploy records, got %d", len(records))
	}
}

func TestDeployGet_PaginatesRecords(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	task, _ := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)

	_, firstPage, total, err := svc.Get(task.ID, 1, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error loading first page: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2 deploy records, got %d", total)
	}
	if len(firstPage) != 1 {
		t.Fatalf("expected 1 deploy record on first page, got %d", len(firstPage))
	}

	_, secondPage, total, err := svc.Get(task.ID, 2, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error loading second page: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2 deploy records on second page, got %d", total)
	}
	if len(secondPage) != 1 {
		t.Fatalf("expected 1 deploy record on second page, got %d", len(secondPage))
	}
	if firstPage[0].ID == secondPage[0].ID {
		t.Error("expected different deploy records on each page")
	}
}

func TestGetAuditLogs(t *testing.T) {
	db, svc := setupDeployTest(t)

	db.Create(&models.AuditLog{UserID: 1, Username: "admin", Action: "create", Resource: "node"})
	db.Create(&models.AuditLog{UserID: 1, Username: "admin", Action: "delete", Resource: "node"})

	logs, total, err := svc.GetAuditLogs(1, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(logs) != 2 {
		t.Errorf("expected 2 audit logs, got total=%d len=%d", total, len(logs))
	}
}

func TestDeployCreate_ConflictDetected(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	// First deploy succeeds
	_, err := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("first deploy unexpected error: %v", err)
	}

	// Second deploy to same cluster without force should return DeployConflictError
	_, err = svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	var conflictErr *DeployConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected DeployConflictError, got %T: %v", err, err)
	}
	if conflictErr.Count != 1 {
		t.Errorf("expected conflict count=1, got %d", conflictErr.Count)
	}
}

func TestDeployCreate_ConflictForced(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	_, err := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("first deploy unexpected error: %v", err)
	}

	// force=true should bypass the conflict check
	task, err := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil, true)
	if err != nil {
		t.Fatalf("forced deploy unexpected error: %v", err)
	}
	if task.TotalNodes != 2 {
		t.Errorf("expected 2 nodes, got %d", task.TotalNodes)
	}
}

func TestDeployCreate_ConflictNodeScope(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, _, nodes := seedDeployData(t, db)

	_, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, nil, false)
	if err != nil {
		t.Fatalf("first deploy unexpected error: %v", err)
	}

	// Second deploy targeting overlapping node without force should conflict
	_, err = svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, nil, false)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	var conflictErr *DeployConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected DeployConflictError, got %T: %v", err, err)
	}
}

func TestGetAuditLogsFiltersAgentPolicyByClusterScope(t *testing.T) {
	db, svc := setupDeployTest(t)
	_, cluster, _ := seedDeployData(t, db)

	otherDC := models.DataCenter{Name: "dc-other"}
	db.Create(&otherDC)
	otherRegion := models.Region{Name: "r-other", DataCenterID: otherDC.ID}
	db.Create(&otherRegion)
	otherCluster := models.Cluster{Name: "c-other", RegionID: otherRegion.ID}
	db.Create(&otherCluster)

	inScopePolicy := models.AgentPolicy{
		Name:      "policy-in",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &cluster.ID,
		IsEnabled: true,
		Settings:  `{}`,
	}
	outOfScopePolicy := models.AgentPolicy{
		Name:      "policy-out",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &otherCluster.ID,
		IsEnabled: true,
		Settings:  `{}`,
	}
	db.Create(&inScopePolicy)
	db.Create(&outOfScopePolicy)

	db.Create(&models.AuditLog{UserID: 1, Username: "admin", Action: "POST", ResourceType: "agent_policy", ResourceID: inScopePolicy.ID, Resource: "/api/v1/agent-policies"})
	db.Create(&models.AuditLog{UserID: 1, Username: "admin", Action: "PUT", ResourceType: "agent_policy", ResourceID: outOfScopePolicy.ID, Resource: "/api/v1/agent-policies/2"})

	logs, total, err := svc.GetAuditLogs(1, 10, []uint{cluster.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected 1 in-scope agent policy audit log, got total=%d len=%d", total, len(logs))
	}
	if logs[0].ResourceID != inScopePolicy.ID {
		t.Fatalf("expected in-scope policy audit log, got resource_id=%d", logs[0].ResourceID)
	}
}
