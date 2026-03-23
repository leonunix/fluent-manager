package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupDeployTest(t *testing.T) (*gorm.DB, DeployService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewDeployService(db)
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

	task, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, nil)
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

	task, err := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil)
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

	task, err := svc.Create(cv.ID, nil, nil, &region.ID, nil, nil, 1, nil)
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

	task, err := svc.Create(cv.ID, nil, nil, nil, &dc.ID, nil, 1, nil)
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

	_, err := svc.Create(cv.ID, nil, nil, nil, nil, nil, 1, nil)
	if err == nil {
		t.Error("expected error when no target nodes")
	}
}

func TestDeployCreate_InvalidConfigVersion(t *testing.T) {
	_, svc := setupDeployTest(t)

	_, err := svc.Create(999, []uint{1}, nil, nil, nil, nil, 1, nil)
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
	_, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, []uint{c2.ID})
	if err == nil {
		t.Error("expected scope violation error")
	}
}

func TestDeployCreate_ScopeCheckPasses(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, nodes := seedDeployData(t, db)

	// User scoped to c1 where nodes are
	task, err := svc.Create(cv.ID, []uint{nodes[0].ID}, nil, nil, nil, nil, 1, []uint{cluster.ID})
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
	task, err := svc.Create(cv.ID, []uint{nodes[0].ID, nodes[1].ID}, &cluster.ID, nil, nil, nil, 1, nil)
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

	svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil)

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

	svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil)

	// User scoped to non-existent cluster
	tasks, total, _ := svc.List(1, 10, []uint{999})
	if total != 0 || len(tasks) != 0 {
		t.Error("scoped user should see no tasks outside their scope")
	}
}

func TestDeployGet(t *testing.T) {
	db, svc := setupDeployTest(t)
	cv, cluster, _ := seedDeployData(t, db)

	task, _ := svc.Create(cv.ID, nil, &cluster.ID, nil, nil, nil, 1, nil)

	gotTask, records, err := svc.Get(task.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTask.ID != task.ID {
		t.Error("wrong task returned")
	}
	if len(records) != 2 {
		t.Errorf("expected 2 deploy records, got %d", len(records))
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
