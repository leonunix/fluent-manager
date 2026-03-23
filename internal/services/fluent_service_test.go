package services

import (
	"errors"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupFluentTest(t *testing.T) (*gorm.DB, FluentService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewFluentService(db, "test-shared-key-secret")
	return db, svc
}

func strPtr(v string) *string {
	return &v
}

func seedCluster(t *testing.T, db *gorm.DB, suffix string) models.Cluster {
	t.Helper()

	dc := models.DataCenter{Name: "dc-" + suffix}
	if err := db.Create(&dc).Error; err != nil {
		t.Fatalf("create datacenter: %v", err)
	}

	region := models.Region{Name: "region-" + suffix, DataCenterID: dc.ID}
	if err := db.Create(&region).Error; err != nil {
		t.Fatalf("create region: %v", err)
	}

	cluster := models.Cluster{Name: "cluster-" + suffix, RegionID: region.ID}
	if err := db.Create(&cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	return cluster
}

func TestAggregationGroupCRUD(t *testing.T) {
	db, svc := setupFluentTest(t)

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name:         "agg-east",
		Alias:        "Aggregator East",
		Description:  "Regional Fluentd aggregation group",
		FluentType:   "fluentd",
		Mode:         "forward",
		EndpointHost: "fluentd.internal",
		EndpointPort: 24224,
		SharedKey:    strPtr("super-secret"),
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if !group.HasSharedKey {
		t.Fatal("expected has_shared_key to be true")
	}
	if group.SharedKey != "" {
		t.Fatal("expected shared_key to be omitted from API model")
	}

	var stored models.AggregationGroup
	if err := db.Unscoped().First(&stored, group.ID).Error; err != nil {
		t.Fatalf("load stored group: %v", err)
	}
	if stored.SharedKey == "" || stored.SharedKey == "super-secret" {
		t.Fatalf("expected encrypted shared key in db, got %q", stored.SharedKey)
	}
	if !isEncryptedSharedKey(stored.SharedKey) {
		t.Fatalf("expected encrypted prefix, got %q", stored.SharedKey)
	}

	updated, err := svc.UpdateAggregationGroup(group.ID, &AggregationGroupInput{
		Name:         "agg-east",
		Alias:        "Aggregator East 1",
		Description:  "Regional Fluentd aggregation group",
		FluentType:   "fluentd",
		Mode:         "forward",
		EndpointHost: "fluentd.internal",
		EndpointPort: 24224,
	})
	if err != nil {
		t.Fatalf("update group: %v", err)
	}
	if updated.Alias != "Aggregator East 1" {
		t.Fatalf("expected updated alias, got %q", updated.Alias)
	}
	if !updated.HasSharedKey {
		t.Fatal("expected shared key to remain configured after update")
	}

	groups, err := svc.ListAggregationGroups(nil)
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	if err := svc.DeleteAggregationGroup(group.ID); err != nil {
		t.Fatalf("delete group: %v", err)
	}
}

func TestListAggregationGroupsScoped(t *testing.T) {
	db, svc := setupFluentTest(t)
	clusterA := seedCluster(t, db, "a")
	clusterB := seedCluster(t, db, "b")

	if _, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name:      "agg-a",
		ClusterID: &clusterA.ID,
	}); err != nil {
		t.Fatalf("create group a: %v", err)
	}
	if _, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name:      "agg-b",
		ClusterID: &clusterB.ID,
	}); err != nil {
		t.Fatalf("create group b: %v", err)
	}
	if _, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name: "agg-global",
	}); err != nil {
		t.Fatalf("create group global: %v", err)
	}

	scoped, err := svc.ListAggregationGroups([]uint{clusterA.ID})
	if err != nil {
		t.Fatalf("list scoped groups: %v", err)
	}
	if len(scoped) != 2 {
		t.Fatalf("expected cluster A group and global group, got %+v", scoped)
	}
	foundCluster := false
	foundGlobal := false
	for _, group := range scoped {
		if group.ClusterID == nil && group.Name == "agg-global" {
			foundGlobal = true
		}
		if group.ClusterID != nil && *group.ClusterID == clusterA.ID && group.Name == "agg-a" {
			foundCluster = true
		}
	}
	if !foundCluster || !foundGlobal {
		t.Fatalf("expected scoped result to include cluster and global groups, got %+v", scoped)
	}
}

func TestGetAggregationGroupScopeDenied(t *testing.T) {
	db, svc := setupFluentTest(t)
	clusterA := seedCluster(t, db, "scope")

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name:      "agg-scope",
		ClusterID: &clusterA.ID,
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	if _, err := svc.GetAggregationGroup(group.ID, []uint{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateAggregationGroupRejectsClusterMismatch(t *testing.T) {
	db, svc := setupFluentTest(t)
	clusterA := seedCluster(t, db, "move-a")
	clusterB := seedCluster(t, db, "move-b")

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name:      "agg-move",
		ClusterID: &clusterA.ID,
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	node := models.Node{
		NodeUID:            "node-1",
		Hostname:           "host-1",
		NodeRole:           models.NodeRoleEdgeCollector,
		AggregationGroupID: &group.ID,
		ClusterID:          &clusterA.ID,
		FluentType:         "fluentbit",
		FluentVersion:      "3.0.0",
		AgentVersion:       "1.0.0",
		Status:             "online",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	_, err = svc.UpdateAggregationGroup(group.ID, &AggregationGroupInput{
		Name:      "agg-move",
		ClusterID: &clusterB.ID,
	})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDeleteAggregationGroupBlockedWhenNodesAssigned(t *testing.T) {
	db, svc := setupFluentTest(t)

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{Name: "agg-east"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	node := models.Node{
		NodeUID:            "node-1",
		Hostname:           "host-1",
		NodeRole:           models.NodeRoleEdgeCollector,
		AggregationGroupID: &group.ID,
		FluentType:         "fluentbit",
		FluentVersion:      "3.0.0",
		AgentVersion:       "1.0.0",
		Status:             "online",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	if err := svc.DeleteAggregationGroup(group.ID); err != ErrHasChildren {
		t.Fatalf("expected ErrHasChildren, got %v", err)
	}
}

func TestRestoreDeletedAggregationGroup(t *testing.T) {
	db, svc := setupFluentTest(t)
	cluster := seedCluster(t, db, "restore")

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{
		Name:      "agg-restore",
		ClusterID: &cluster.ID,
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	if err := svc.DeleteAggregationGroup(group.ID); err != nil {
		t.Fatalf("delete group: %v", err)
	}

	deleted, err := svc.ListDeletedAggregationGroups([]uint{cluster.ID})
	if err != nil {
		t.Fatalf("list deleted groups: %v", err)
	}
	if len(deleted) != 1 {
		t.Fatalf("expected 1 deleted group, got %d", len(deleted))
	}

	restored, err := svc.RestoreAggregationGroup(group.ID, []uint{cluster.ID})
	if err != nil {
		t.Fatalf("restore group: %v", err)
	}
	if restored.ID != group.ID {
		t.Fatalf("expected restored group %d, got %d", group.ID, restored.ID)
	}

	active, err := svc.ListAggregationGroups([]uint{cluster.ID})
	if err != nil {
		t.Fatalf("list active groups: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active group after restore, got %d", len(active))
	}
}

func TestLegacySharedKeysMigratedOnServiceInit(t *testing.T) {
	db := testutil.NewTestDB()
	legacy := models.AggregationGroup{
		Name:      "legacy-secret",
		SharedKey: "legacy-plaintext",
	}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("create legacy group: %v", err)
	}

	_ = NewFluentService(db, "migration-secret")

	var stored models.AggregationGroup
	if err := db.Unscoped().First(&stored, legacy.ID).Error; err != nil {
		t.Fatalf("load migrated group: %v", err)
	}
	if stored.SharedKey == "legacy-plaintext" {
		t.Fatal("expected legacy shared key to be migrated to encrypted storage")
	}
	if !isEncryptedSharedKey(stored.SharedKey) {
		t.Fatalf("expected encrypted prefix after migration, got %q", stored.SharedKey)
	}
}

func TestCreateAggregationGroupAllowsNameReuseAfterSoftDelete(t *testing.T) {
	_, svc := setupFluentTest(t)

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{Name: "agg-reusable"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := svc.DeleteAggregationGroup(group.ID); err != nil {
		t.Fatalf("delete group: %v", err)
	}
	recreated, err := svc.CreateAggregationGroup(&AggregationGroupInput{Name: "agg-reusable"})
	if err != nil {
		t.Fatalf("recreate group with soft-deleted name: %v", err)
	}
	if recreated.ID == group.ID {
		t.Fatalf("expected recreated group to have a new id, got %d", recreated.ID)
	}
}

func TestUpsertNodeProfile(t *testing.T) {
	db, svc := setupFluentTest(t)

	group, err := svc.CreateAggregationGroup(&AggregationGroupInput{Name: "agg-east"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	node := models.Node{
		NodeUID:       "node-1",
		Hostname:      "host-1",
		NodeRole:      models.NodeRoleStandalone,
		FluentType:    "fluentbit",
		FluentVersion: "3.0.0",
		AgentVersion:  "1.0.0",
		Status:        "online",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	profile, err := svc.UpsertNodeProfile(node.ID, &NodeFluentProfileInput{
		NodeRole:             models.NodeRoleEdgeCollector,
		AggregationGroupID:   &group.ID,
		LoadedPlugins:        "forward,tail,kubernetes",
		SupportsHotReload:    true,
		SupportsMultiline:    true,
		SupportsStorageLayer: true,
		SupportsForwardTLS:   true,
		SupportsMetricsAPI:   true,
		Metadata:             `{"distribution":"custom"}`,
	})
	if err != nil {
		t.Fatalf("upsert profile: %v", err)
	}
	if profile.NodeID != node.ID {
		t.Fatalf("expected profile for node %d, got %d", node.ID, profile.NodeID)
	}
	if !profile.SupportsForwardTLS {
		t.Fatalf("expected forward TLS support to be true")
	}
	if profile.Node == nil || profile.Node.AggregationGroup == nil {
		t.Fatalf("expected aggregation group to be preloaded on node profile response")
	}
	if profile.Node.AggregationGroup.HasSharedKey {
		t.Fatalf("expected has_shared_key to be false for group without shared key")
	}
	if profile.Node.AggregationGroup.SharedKey != "" {
		t.Fatalf("expected shared_key to stay hidden on node profile response")
	}

	var updated models.Node
	if err := db.First(&updated, node.ID).Error; err != nil {
		t.Fatalf("reload node: %v", err)
	}
	if updated.NodeRole != models.NodeRoleEdgeCollector {
		t.Fatalf("expected node role %q, got %q", models.NodeRoleEdgeCollector, updated.NodeRole)
	}
	if updated.AggregationGroupID == nil || *updated.AggregationGroupID != group.ID {
		t.Fatalf("expected aggregation group %d, got %v", group.ID, updated.AggregationGroupID)
	}
}
