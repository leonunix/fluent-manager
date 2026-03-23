package services

import (
	"fmt"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupTopoTest(t *testing.T) (*gorm.DB, TopologyService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewTopologyService(db)
	return db, svc
}

// ---------- DataCenter ----------

func TestCreateDataCenter(t *testing.T) {
	_, svc := setupTopoTest(t)
	dc, err := svc.CreateDataCenter("aws-us", "AWS US", "aws", "Virginia", "Main DC", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dc.ID == 0 || dc.Name != "aws-us" {
		t.Error("datacenter not created correctly")
	}
}

func TestListDataCenters_Global(t *testing.T) {
	_, svc := setupTopoTest(t)
	svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	svc.CreateDataCenter("dc2", "", "azure", "", "", "")

	dcs, err := svc.ListDataCenters(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dcs) != 2 {
		t.Errorf("expected 2 DCs, got %d", len(dcs))
	}
}

func TestListDataCenters_Scoped(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc1, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	svc.CreateDataCenter("dc2", "", "azure", "", "", "")

	r := models.Region{Name: "r1", DataCenterID: dc1.ID}
	db.Create(&r)

	dcs, err := svc.ListDataCenters([]uint{dc1.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dcs) != 1 || dcs[0].Name != "dc1" {
		t.Errorf("expected only dc1, got %d results", len(dcs))
	}
}

func TestGetDataCenter(t *testing.T) {
	_, svc := setupTopoTest(t)
	created, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")

	dc, err := svc.GetDataCenter(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dc.Name != "dc1" {
		t.Error("wrong datacenter returned")
	}
}

func TestGetDataCenter_NotFound(t *testing.T) {
	_, svc := setupTopoTest(t)
	_, err := svc.GetDataCenter(999)
	if err == nil {
		t.Error("expected error for non-existent DC")
	}
}

func TestUpdateDataCenter(t *testing.T) {
	_, svc := setupTopoTest(t)
	created, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")

	updated, err := svc.UpdateDataCenter(created.ID, "dc1-renamed", "Alias", "azure", "Tokyo", "Updated", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "dc1-renamed" || updated.Provider != "azure" {
		t.Error("datacenter not updated correctly")
	}
}

func TestDeleteDataCenter(t *testing.T) {
	_, svc := setupTopoTest(t)
	created, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")

	if err := svc.DeleteDataCenter(created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetDataCenter(created.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestDeleteDataCenter_HasChildren(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")

	db.Create(&models.Region{Name: "r1", DataCenterID: dc.ID})

	err := svc.DeleteDataCenter(dc.ID)
	if err != ErrHasChildren {
		t.Errorf("expected ErrHasChildren, got %v", err)
	}
}

// ---------- Region ----------

func TestCreateRegion(t *testing.T) {
	_, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")

	region, err := svc.CreateRegion("us-west-2", "US West", dc.ID, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if region.Name != "us-west-2" || region.DataCenterID != dc.ID {
		t.Error("region not created correctly")
	}
}

func TestCreateRegion_InvalidDC(t *testing.T) {
	_, svc := setupTopoTest(t)
	_, err := svc.CreateRegion("r1", "", 999, "", "")
	if err == nil {
		t.Error("expected error for non-existent DC")
	}
}

func TestListRegions_Filtered(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc1, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	dc2, _ := svc.CreateDataCenter("dc2", "", "azure", "", "", "")

	db.Create(&models.Region{Name: "r1", DataCenterID: dc1.ID})
	db.Create(&models.Region{Name: "r2", DataCenterID: dc2.ID})

	regions, err := svc.ListRegions([]uint{dc1.ID}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(regions) != 1 || regions[0].Name != "r1" {
		t.Errorf("expected only r1 for dc1 scope")
	}
}

func TestDeleteRegion_HasChildren(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")

	db.Create(&models.Cluster{Name: "c1", RegionID: region.ID})

	err := svc.DeleteRegion(region.ID)
	if err != ErrHasChildren {
		t.Errorf("expected ErrHasChildren, got %v", err)
	}
}

// ---------- Cluster ----------

func TestCreateCluster(t *testing.T) {
	_, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")

	cluster, err := svc.CreateCluster("c1", "Cluster 1", region.ID, nil, false, nil, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cluster.Name != "c1" || cluster.RegionID != region.ID {
		t.Error("cluster not created correctly")
	}
}

func TestCreateCluster_Default(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")

	// Create first as default
	svc.CreateCluster("c1", "", region.ID, nil, true, nil, "", "")
	// Create second as default - should unset first
	svc.CreateCluster("c2", "", region.ID, nil, true, nil, "", "")

	var c1 models.Cluster
	db.Where("name = ?", "c1").First(&c1)
	if c1.IsDefault {
		t.Error("c1 should no longer be default")
	}
}

func TestListClusters_ScopeFiltered(t *testing.T) {
	_, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")

	c1, _ := svc.CreateCluster("c1", "", region.ID, nil, false, nil, "", "")
	svc.CreateCluster("c2", "", region.ID, nil, false, nil, "", "")

	clusters, err := svc.ListClusters([]uint{c1.ID}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "c1" {
		t.Error("scope filter should only return c1")
	}
}

func TestDeleteCluster_HasChildren(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")
	cluster, _ := svc.CreateCluster("c1", "", region.ID, nil, false, nil, "", "")

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &cluster.ID})

	err := svc.DeleteCluster(cluster.ID)
	if err != ErrHasChildren {
		t.Errorf("expected ErrHasChildren, got %v", err)
	}
}

// ---------- Match Rules ----------

func TestMatchRuleCRUD(t *testing.T) {
	_, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")
	cluster, _ := svc.CreateCluster("c1", "", region.ID, nil, false, nil, "", "")

	// Create
	rule, err := svc.CreateMatchRule(cluster.ID, &models.ClusterMatchRule{
		Name:            "web-rule",
		Priority:        1,
		HostnamePattern: "web-*",
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ClusterID != cluster.ID {
		t.Error("rule should be associated with cluster")
	}

	// List
	rules, err := svc.ListMatchRules(fmt.Sprintf("%d", cluster.ID))
	if err != nil {
		t.Fatalf("unexpected error listing rules: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}

	// Update
	updated, err := svc.UpdateMatchRule(cluster.ID, rule.ID, &models.ClusterMatchRule{
		Name:            "web-rule-updated",
		Priority:        2,
		HostnamePattern: "web-prod-*",
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "web-rule-updated" {
		t.Error("rule not updated")
	}

	// Delete
	if err := svc.DeleteMatchRule(cluster.ID, rule.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- User Scopes ----------

func TestSetUserScopes(t *testing.T) {
	db, svc := setupTopoTest(t)

	dc, _ := svc.CreateDataCenter("dc1", "DC Alias", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "Region Alias", dc.ID, "", "")
	cluster, _ := svc.CreateCluster("c1", "Cluster Alias", region.ID, nil, false, nil, "", "")

	user := models.User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)

	err := svc.SetUserScopes(user.ID, []ScopeInput{
		{ScopeType: "datacenter", ScopeID: dc.ID},
		{ScopeType: "cluster", ScopeID: cluster.ID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var scopes []models.UserScope
	db.Where("user_id = ?", user.ID).Find(&scopes)
	if len(scopes) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(scopes))
	}

	// Verify scope names are resolved
	for _, s := range scopes {
		if s.ScopeName == "" {
			t.Errorf("scope name should be resolved for %s", s.ScopeType)
		}
	}
}

func TestSetUserScopes_Replaces(t *testing.T) {
	db, svc := setupTopoTest(t)

	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")

	user := models.User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)

	svc.SetUserScopes(user.ID, []ScopeInput{{ScopeType: "datacenter", ScopeID: dc.ID}})

	// Replace with empty
	svc.SetUserScopes(user.ID, []ScopeInput{})

	var scopes []models.UserScope
	db.Where("user_id = ?", user.ID).Find(&scopes)
	if len(scopes) != 0 {
		t.Errorf("expected 0 scopes after replace, got %d", len(scopes))
	}
}

// ---------- Environment ----------

func TestEnvironmentCRUD(t *testing.T) {
	_, svc := setupTopoTest(t)

	env, err := svc.CreateEnvironment(&models.Environment{
		Name: "staging", Color: "#ffc107", SortOrder: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envs, _ := svc.ListEnvironments()
	if len(envs) != 1 {
		t.Errorf("expected 1 env, got %d", len(envs))
	}

	updated, err := svc.UpdateEnvironment(env.ID, &models.Environment{
		Name: "pre-prod", Color: "#ff5733", SortOrder: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "pre-prod" {
		t.Error("env not updated")
	}

	if err := svc.DeleteEnvironment(env.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- Tree ----------

func TestGetTree_Global(t *testing.T) {
	db, svc := setupTopoTest(t)

	dc, _ := svc.CreateDataCenter("dc1", "DC1", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")
	cluster, _ := svc.CreateCluster("c1", "", region.ID, nil, false, nil, "", "")

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &cluster.ID, Status: "online"})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &cluster.ID, Status: "offline"})

	tree, err := svc.GetTree(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("expected 1 DC in tree, got %d", len(tree))
	}
	if len(tree[0].Regions) != 1 {
		t.Fatalf("expected 1 region, got %d", len(tree[0].Regions))
	}
	if len(tree[0].Regions[0].Clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(tree[0].Regions[0].Clusters))
	}
	cl := tree[0].Regions[0].Clusters[0]
	if cl.NodeCount != 2 || cl.OnlineCount != 1 {
		t.Errorf("expected 2 nodes (1 online), got %d (%d online)", cl.NodeCount, cl.OnlineCount)
	}
}

func TestGetTree_Scoped(t *testing.T) {
	_, svc := setupTopoTest(t)

	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")
	c1, _ := svc.CreateCluster("c1", "", region.ID, nil, false, nil, "", "")
	svc.CreateCluster("c2", "", region.ID, nil, false, nil, "", "")

	tree, _ := svc.GetTree([]uint{c1.ID}, []uint{dc.ID})
	if len(tree) != 1 {
		t.Fatalf("expected 1 DC, got %d", len(tree))
	}
	if len(tree[0].Regions[0].Clusters) != 1 {
		t.Errorf("expected 1 cluster (c1 only), got %d", len(tree[0].Regions[0].Clusters))
	}
}

// ---------- AllowedDCIDs ----------

func TestAllowedDCIDs_Nil(t *testing.T) {
	_, svc := setupTopoTest(t)
	result := svc.AllowedDCIDs(nil)
	if result != nil {
		t.Error("nil clusters should return nil DC IDs")
	}
}

func TestAllowedDCIDs_Resolved(t *testing.T) {
	_, svc := setupTopoTest(t)
	dc1, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	dc2, _ := svc.CreateDataCenter("dc2", "", "azure", "", "", "")
	r1, _ := svc.CreateRegion("r1", "", dc1.ID, "", "")
	r2, _ := svc.CreateRegion("r2", "", dc2.ID, "", "")
	c1, _ := svc.CreateCluster("c1", "", r1.ID, nil, false, nil, "", "")
	svc.CreateCluster("c2", "", r2.ID, nil, false, nil, "", "")

	dcIDs := svc.AllowedDCIDs([]uint{c1.ID})
	if len(dcIDs) != 1 || dcIDs[0] != dc1.ID {
		t.Errorf("expected only dc1, got %v", dcIDs)
	}
}

// ---------- DataCenter with counts ----------

func TestListDataCenters_Counts(t *testing.T) {
	db, svc := setupTopoTest(t)
	dc, _ := svc.CreateDataCenter("dc1", "", "aws", "", "", "")
	region, _ := svc.CreateRegion("r1", "", dc.ID, "", "")
	cluster, _ := svc.CreateCluster("c1", "", region.ID, nil, false, nil, "", "")

	db.Create(&models.Node{NodeUID: "n1", Hostname: "h1", ClusterID: &cluster.ID})
	db.Create(&models.Node{NodeUID: "n2", Hostname: "h2", ClusterID: &cluster.ID})

	dcs, _ := svc.ListDataCenters(nil)
	if dcs[0].RegionCount != 1 || dcs[0].ClusterCount != 1 || dcs[0].NodeCount != 2 {
		t.Errorf("wrong counts: region=%d, cluster=%d, node=%d",
			dcs[0].RegionCount, dcs[0].ClusterCount, dcs[0].NodeCount)
	}
}
