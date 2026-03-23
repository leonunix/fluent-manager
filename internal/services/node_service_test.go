package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupNodeTest(t *testing.T) (*gorm.DB, NodeService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewNodeService(db)
	return db, svc
}

func seedTopology(t *testing.T, db *gorm.DB) (dc *models.DataCenter, region *models.Region, cluster *models.Cluster) {
	t.Helper()
	dc = &models.DataCenter{Name: "dc1", Provider: "aws"}
	db.Create(dc)
	region = &models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(region)
	cluster = &models.Cluster{Name: "c1", RegionID: region.ID}
	db.Create(cluster)
	return
}

func TestNodeList_Global(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	db.Create(&models.Node{NodeUID: "n1", Hostname: "web-01", ClusterID: &c.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "web-02", ClusterID: &c.ID, Status: "offline"})

	nodes, total, err := svc.List(NodeListFilters{}, nil, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got total=%d len=%d", total, len(nodes))
	}
}

func TestNodeList_ScopeFiltered(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c1 := seedTopology(t, db)

	r2 := models.Region{Name: "r2", DataCenterID: 1}
	db.Create(&r2)
	c2 := models.Cluster{Name: "c2", RegionID: r2.ID}
	db.Create(&c2)

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &c1.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &c2.ID, Status: "online"})

	// Scoped to c1 only
	nodes, total, _ := svc.List(NodeListFilters{}, []uint{c1.ID}, 1, 10)
	if total != 1 || nodes[0].Hostname != "h1" {
		t.Errorf("scope filter should only return nodes in c1, got total=%d", total)
	}
}

func TestNodeList_StatusFilter(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &c.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &c.ID, Status: "offline"})

	nodes, total, _ := svc.List(NodeListFilters{Status: "online"}, nil, 1, 10)
	if total != 1 || nodes[0].Status != "online" {
		t.Error("status filter not working")
	}
}

func TestNodeList_SearchFilter(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	db.Create(&models.Node{NodeUID: "uid-abc", Hostname: "web-01", IPAddress: "10.0.0.1", ClusterID: &c.ID})
	db.Create(&models.Node{NodeUID: "uid-def", Hostname: "db-01", IPAddress: "10.0.0.2", ClusterID: &c.ID})

	nodes, total, _ := svc.List(NodeListFilters{Search: "web"}, nil, 1, 10)
	if total != 1 || nodes[0].Hostname != "web-01" {
		t.Error("search filter not working")
	}
}

func TestNodeList_Pagination(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	for i := 0; i < 15; i++ {
		db.Create(&models.Node{NodeUID: "n" + string(rune('a'+i)), Hostname: "h", ClusterID: &c.ID})
	}

	nodes, total, _ := svc.List(NodeListFilters{}, nil, 2, 10)
	if total != 15 {
		t.Errorf("expected total=15, got %d", total)
	}
	if len(nodes) != 5 {
		t.Errorf("expected 5 nodes on page 2, got %d", len(nodes))
	}
}

func TestNodeGet(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	n := models.Node{NodeUID: "n1", Hostname: "web-01", ClusterID: &c.ID}
	db.Create(&n)

	node, err := svc.Get(n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Hostname != "web-01" {
		t.Error("wrong node returned")
	}
}

func TestNodeGet_NotFound(t *testing.T) {
	_, svc := setupNodeTest(t)
	_, err := svc.Get(999)
	if err == nil {
		t.Error("expected error for non-existent node")
	}
}

func TestNodeUpdate(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	n := models.Node{NodeUID: "n1", Hostname: "web-01", ClusterID: &c.ID}
	db.Create(&n)

	updated, err := svc.Update(n.ID, map[string]interface{}{"hostname": "web-01-renamed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Hostname != "web-01-renamed" {
		t.Error("node not updated")
	}
}

func TestNodeDelete(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	n := models.Node{NodeUID: "n1", Hostname: "web-01", ClusterID: &c.ID}
	db.Create(&n)

	if err := svc.Delete(n.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(n.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestNodeBatchMoveCluster(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c1 := seedTopology(t, db)

	c2 := models.Cluster{Name: "c2", RegionID: 1}
	db.Create(&c2)

	n1 := models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &c1.ID}
	n2 := models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &c1.ID}
	db.Create(&n1)
	db.Create(&n2)

	err := svc.BatchMoveCluster([]uint{n1.ID, n2.ID}, c2.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var movedNode models.Node
	db.First(&movedNode, n1.ID)
	if movedNode.ClusterID == nil || *movedNode.ClusterID != c2.ID {
		t.Error("node not moved to c2")
	}
}

func TestNodeStats_Global(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c := seedTopology(t, db)

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &c.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &c.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n3", Hostname: "h3", ClusterID: &c.ID, Status: "offline"})

	counts, total, err := svc.Stats(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}

	statusMap := map[string]int64{}
	for _, c := range counts {
		statusMap[c.Status] = c.Count
	}
	if statusMap["online"] != 2 || statusMap["offline"] != 1 {
		t.Errorf("wrong status counts: %v", statusMap)
	}
}

func TestNodeStats_Scoped(t *testing.T) {
	db, svc := setupNodeTest(t)
	_, _, c1 := seedTopology(t, db)

	c2 := models.Cluster{Name: "c2", RegionID: 1}
	db.Create(&c2)

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &c1.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &c2.ID, Status: "online"})

	_, total, _ := svc.Stats([]uint{c1.ID})
	if total != 1 {
		t.Errorf("expected 1 node for scoped stats, got %d", total)
	}
}
