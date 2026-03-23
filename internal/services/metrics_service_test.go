package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupMetricsTest(t *testing.T) (*gorm.DB, MetricsService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewMetricsService(db)
	return db, svc
}

func seedMetricsData(t *testing.T, db *gorm.DB) (dc *models.DataCenter, cluster *models.Cluster) {
	t.Helper()
	dc1 := models.DataCenter{Name: "dc1", Alias: "DC One"}
	db.Create(&dc1)
	r := models.Region{Name: "r1", DataCenterID: dc1.ID}
	db.Create(&r)
	c := models.Cluster{Name: "c1", RegionID: r.ID}
	db.Create(&c)

	n1 := models.Node{NodeUID: "n1", Hostname: "web-01", IPAddress: "10.0.0.1", ClusterID: &c.ID, Status: "online"}
	n2 := models.Node{NodeUID: "n2", Hostname: "web-02", IPAddress: "10.0.0.2", ClusterID: &c.ID, Status: "online"}
	db.Create(&n1)
	db.Create(&n2)

	db.Create(&models.NodeMetrics{NodeID: n1.ID, CPUUsagePercent: 40, MemUsagePercent: 60, DiskUsagePercent: 30, FluentRunning: true})
	db.Create(&models.NodeMetrics{NodeID: n2.ID, CPUUsagePercent: 60, MemUsagePercent: 80, DiskUsagePercent: 50, FluentRunning: false})

	return &dc1, &c
}

func TestMetricsOverview_Global(t *testing.T) {
	db, svc := setupMetricsTest(t)
	seedMetricsData(t, db)

	result, err := svc.Overview(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalNodes != 2 {
		t.Errorf("expected 2 nodes, got %d", result.TotalNodes)
	}
	if result.AvgCPU != 50 {
		t.Errorf("expected avg CPU 50, got %f", result.AvgCPU)
	}
	if result.FluentRunning != 1 {
		t.Errorf("expected 1 fluent running, got %d", result.FluentRunning)
	}
	if result.FluentRunRate != 50 {
		t.Errorf("expected 50%% run rate, got %f", result.FluentRunRate)
	}
}

func TestMetricsOverview_Scoped(t *testing.T) {
	db, svc := setupMetricsTest(t)
	_, c := seedMetricsData(t, db)

	// Scoped to existing cluster
	result, _ := svc.Overview([]uint{c.ID})
	if result.TotalNodes != 2 {
		t.Errorf("expected 2, got %d", result.TotalNodes)
	}

	// Scoped to non-existent cluster
	result, _ = svc.Overview([]uint{999})
	if result.TotalNodes != 0 {
		t.Errorf("expected 0, got %d", result.TotalNodes)
	}
}

func TestMetricsOverview_Empty(t *testing.T) {
	_, svc := setupMetricsTest(t)

	result, err := svc.Overview(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalNodes != 0 {
		t.Errorf("expected 0 nodes, got %d", result.TotalNodes)
	}
}

func TestTopNodes_Global(t *testing.T) {
	db, svc := setupMetricsTest(t)
	seedMetricsData(t, db)

	nodes, err := svc.TopNodes(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2 top nodes, got %d", len(nodes))
	}
	// Should be sorted by CPU DESC
	if nodes[0].CPU < nodes[1].CPU {
		t.Error("top nodes should be sorted by CPU descending")
	}
}

func TestTopNodes_Scoped(t *testing.T) {
	db, svc := setupMetricsTest(t)
	seedMetricsData(t, db)
	_ = db

	nodes, _ := svc.TopNodes([]uint{999})
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes for out-of-scope cluster, got %d", len(nodes))
	}
}

func TestByDatacenter_Global(t *testing.T) {
	db, svc := setupMetricsTest(t)
	dc, _ := seedMetricsData(t, db)

	results, err := svc.ByDatacenter(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 DC result, got %d", len(results))
	}
	if results[0].DCID != dc.ID {
		t.Error("wrong DC in result")
	}
	if results[0].NodeCount != 2 {
		t.Errorf("expected 2 nodes, got %d", results[0].NodeCount)
	}
}

func TestByDatacenter_Scoped(t *testing.T) {
	db, svc := setupMetricsTest(t)
	seedMetricsData(t, db)

	results, _ := svc.ByDatacenter([]uint{999})
	if len(results) != 0 {
		t.Error("expected 0 results for out-of-scope DC")
	}
}

func TestByDatacenter_Empty(t *testing.T) {
	_, svc := setupMetricsTest(t)

	results, err := svc.ByDatacenter(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Error("expected empty results")
	}
}
