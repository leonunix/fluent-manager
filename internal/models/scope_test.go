package models

import (
	"fmt"
	"sync/atomic"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDBCounter atomic.Int64

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	n := testDBCounter.Add(1)
	dsn := fmt.Sprintf("file:testdb%d?mode=memory&cache=shared", n)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&User{}, &Role{}, &Permission{}, &UserScope{},
		&Environment{}, &DataCenter{}, &Region{}, &Cluster{},
		&ClusterMatchRule{}, &Node{}, &NodeMetrics{},
		&RemoteCommand{}, &NodeLog{},
		&ConfigTemplate{}, &ConfigVersion{},
		&DeployTask{}, &DeployRecord{}, &BootstrapHost{}, &BootstrapTask{}, &BootstrapRecord{}, &AuditLog{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	DB = db
	return db
}

// ---------- MatchNode tests ----------

func TestMatchNode_InactiveRule(t *testing.T) {
	rule := ClusterMatchRule{IsActive: false, HostnamePattern: "web-*"}
	if rule.MatchNode("web-01", "10.0.0.1", "fluentbit", "linux", "") {
		t.Error("inactive rule should not match")
	}
}

func TestMatchNode_HostnameGlob(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, HostnamePattern: "web-*"}

	if !rule.MatchNode("web-01", "", "", "", "") {
		t.Error("should match web-01")
	}
	if rule.MatchNode("db-01", "", "", "", "") {
		t.Error("should not match db-01")
	}
}

func TestMatchNode_HostnameExact(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, HostnamePattern: "app-server"}
	if !rule.MatchNode("app-server", "", "", "", "") {
		t.Error("should match exact hostname")
	}
	if rule.MatchNode("app-server-2", "", "", "", "") {
		t.Error("should not match different hostname")
	}
}

func TestMatchNode_IPPatternCIDR(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, IPPattern: "10.0.0.0/24"}

	if !rule.MatchNode("", "10.0.0.5", "", "", "") {
		t.Error("should match IP in CIDR range")
	}
	if rule.MatchNode("", "10.0.1.5", "", "", "") {
		t.Error("should not match IP outside CIDR range")
	}
}

func TestMatchNode_IPPatternCIDR16(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, IPPattern: "192.168.0.0/16"}

	if !rule.MatchNode("", "192.168.1.100", "", "", "") {
		t.Error("should match IP in /16 range")
	}
	if rule.MatchNode("", "192.169.0.1", "", "", "") {
		t.Error("should not match IP outside /16 range")
	}
}

func TestMatchNode_IPPatternGlob(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, IPPattern: "10.0.1.*"}

	if !rule.MatchNode("", "10.0.1.50", "", "", "") {
		t.Error("should match IP glob")
	}
	if rule.MatchNode("", "10.0.2.50", "", "", "") {
		t.Error("should not match IP outside glob")
	}
}

func TestMatchNode_IPPatternInvalidCIDR(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, IPPattern: "invalid/99"}
	if rule.MatchNode("", "10.0.0.1", "", "", "") {
		t.Error("invalid CIDR should not match")
	}
}

func TestMatchNode_FluentType(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, FluentType: "fluentbit"}

	if !rule.MatchNode("", "", "fluentbit", "", "") {
		t.Error("should match fluentbit")
	}
	if rule.MatchNode("", "", "fluentd", "", "") {
		t.Error("should not match fluentd")
	}
}

func TestMatchNode_FluentTypeEmpty(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, FluentType: ""}
	if !rule.MatchNode("", "", "fluentbit", "", "") {
		t.Error("empty fluent type should match any")
	}
}

func TestMatchNode_OSPattern(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, OSPattern: "linux*"}

	if !rule.MatchNode("", "", "", "Linux", "") {
		t.Error("should match linux (case insensitive)")
	}
	if !rule.MatchNode("", "", "", "linux-amd64", "") {
		t.Error("should match linux-amd64")
	}
	if rule.MatchNode("", "", "", "Windows", "") {
		t.Error("should not match windows")
	}
}

func TestMatchNode_LabelSelector(t *testing.T) {
	rule := ClusterMatchRule{
		IsActive:      true,
		LabelSelector: `{"env":"prod","role":"web"}`,
	}

	if !rule.MatchNode("", "", "", "", `{"env":"prod","role":"web","team":"infra"}`) {
		t.Error("should match when node has all required labels (superset)")
	}
	if rule.MatchNode("", "", "", "", `{"env":"prod"}`) {
		t.Error("should not match when node is missing labels")
	}
	if rule.MatchNode("", "", "", "", `{"env":"staging","role":"web"}`) {
		t.Error("should not match when label value differs")
	}
}

func TestMatchNode_LabelSelector_EmptyNodeLabels(t *testing.T) {
	rule := ClusterMatchRule{
		IsActive:      true,
		LabelSelector: `{"env":"prod"}`,
	}
	if rule.MatchNode("", "", "", "", "") {
		t.Error("empty node labels should not match")
	}
}

func TestMatchNode_LabelSelector_InvalidJSON(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, LabelSelector: `{invalid}`}
	if rule.MatchNode("", "", "", "", `{"env":"prod"}`) {
		t.Error("invalid selector JSON should not match")
	}
}

func TestMatchNode_LabelSelector_InvalidNodeJSON(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true, LabelSelector: `{"env":"prod"}`}
	if rule.MatchNode("", "", "", "", `{invalid}`) {
		t.Error("invalid node labels JSON should not match")
	}
}

func TestMatchNode_MultipleConditions(t *testing.T) {
	rule := ClusterMatchRule{
		IsActive:        true,
		HostnamePattern: "web-*",
		IPPattern:       "10.0.0.0/24",
		FluentType:      "fluentbit",
		OSPattern:       "linux",
	}

	if !rule.MatchNode("web-01", "10.0.0.5", "fluentbit", "linux", "") {
		t.Error("should match when all conditions met")
	}
	if rule.MatchNode("web-01", "10.0.0.5", "fluentd", "linux", "") {
		t.Error("should not match when one condition fails (fluent_type)")
	}
	if rule.MatchNode("db-01", "10.0.0.5", "fluentbit", "linux", "") {
		t.Error("should not match when hostname fails")
	}
}

func TestMatchNode_NoConditions(t *testing.T) {
	rule := ClusterMatchRule{IsActive: true}
	if !rule.MatchNode("anything", "1.2.3.4", "fluentbit", "linux", "") {
		t.Error("rule with no conditions should match everything")
	}
}

// ---------- AllowedClusterIDs tests ----------

func TestAllowedClusterIDs_NoScopes(t *testing.T) {
	db := setupTestDB(t)
	user := User{Username: "admin", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)

	result := AllowedClusterIDs(user.ID)
	if result != nil {
		t.Error("user with no scopes should have global access (nil)")
	}
}

func TestAllowedClusterIDs_ClusterScope(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	region := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&region)
	c1 := Cluster{Name: "c1", RegionID: region.ID}
	c2 := Cluster{Name: "c2", RegionID: region.ID}
	c3 := Cluster{Name: "c3", RegionID: region.ID}
	db.Create(&c1)
	db.Create(&c2)
	db.Create(&c3)

	user := User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)

	db.Create(&UserScope{UserID: user.ID, ScopeType: "cluster", ScopeID: c1.ID})
	db.Create(&UserScope{UserID: user.ID, ScopeType: "cluster", ScopeID: c2.ID})

	result := AllowedClusterIDs(user.ID)
	if result == nil {
		t.Fatal("expected non-nil result for scoped user")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(result))
	}

	allowed := map[uint]bool{}
	for _, id := range result {
		allowed[id] = true
	}
	if !allowed[c1.ID] || !allowed[c2.ID] {
		t.Error("should contain c1 and c2")
	}
	if allowed[c3.ID] {
		t.Error("should not contain c3")
	}
}

func TestAllowedClusterIDs_RegionScope(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r1 := Region{Name: "r1", DataCenterID: dc.ID}
	r2 := Region{Name: "r2", DataCenterID: dc.ID}
	db.Create(&r1)
	db.Create(&r2)
	c1 := Cluster{Name: "c1", RegionID: r1.ID}
	c2 := Cluster{Name: "c2", RegionID: r1.ID}
	c3 := Cluster{Name: "c3", RegionID: r2.ID}
	db.Create(&c1)
	db.Create(&c2)
	db.Create(&c3)

	user := User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	db.Create(&UserScope{UserID: user.ID, ScopeType: "region", ScopeID: r1.ID})

	result := AllowedClusterIDs(user.ID)
	if len(result) != 2 {
		t.Fatalf("expected 2 clusters for region scope, got %d", len(result))
	}

	allowed := map[uint]bool{}
	for _, id := range result {
		allowed[id] = true
	}
	if !allowed[c1.ID] || !allowed[c2.ID] {
		t.Error("should contain clusters from r1")
	}
	if allowed[c3.ID] {
		t.Error("should not contain clusters from r2")
	}
}

func TestAllowedClusterIDs_DatacenterScope(t *testing.T) {
	db := setupTestDB(t)

	dc1 := DataCenter{Name: "dc1"}
	dc2 := DataCenter{Name: "dc2"}
	db.Create(&dc1)
	db.Create(&dc2)
	r1 := Region{Name: "r1", DataCenterID: dc1.ID}
	r2 := Region{Name: "r2", DataCenterID: dc2.ID}
	db.Create(&r1)
	db.Create(&r2)
	c1 := Cluster{Name: "c1", RegionID: r1.ID}
	c2 := Cluster{Name: "c2", RegionID: r1.ID}
	c3 := Cluster{Name: "c3", RegionID: r2.ID}
	db.Create(&c1)
	db.Create(&c2)
	db.Create(&c3)

	user := User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	db.Create(&UserScope{UserID: user.ID, ScopeType: "datacenter", ScopeID: dc1.ID})

	result := AllowedClusterIDs(user.ID)
	if len(result) != 2 {
		t.Fatalf("expected 2 clusters for datacenter scope, got %d", len(result))
	}

	allowed := map[uint]bool{}
	for _, id := range result {
		allowed[id] = true
	}
	if !allowed[c1.ID] || !allowed[c2.ID] {
		t.Error("should contain clusters from dc1")
	}
	if allowed[c3.ID] {
		t.Error("should not contain clusters from dc2")
	}
}

func TestAllowedClusterIDs_MixedScopes(t *testing.T) {
	db := setupTestDB(t)

	dc1 := DataCenter{Name: "dc1"}
	dc2 := DataCenter{Name: "dc2"}
	db.Create(&dc1)
	db.Create(&dc2)
	r1 := Region{Name: "r1", DataCenterID: dc1.ID}
	r2 := Region{Name: "r2", DataCenterID: dc2.ID}
	db.Create(&r1)
	db.Create(&r2)
	c1 := Cluster{Name: "c1", RegionID: r1.ID}
	c2 := Cluster{Name: "c2", RegionID: r2.ID}
	c3 := Cluster{Name: "c3", RegionID: r2.ID}
	db.Create(&c1)
	db.Create(&c2)
	db.Create(&c3)

	user := User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	// region scope for r1 (gives c1) + cluster scope for c2
	db.Create(&UserScope{UserID: user.ID, ScopeType: "region", ScopeID: r1.ID})
	db.Create(&UserScope{UserID: user.ID, ScopeType: "cluster", ScopeID: c2.ID})

	result := AllowedClusterIDs(user.ID)
	if len(result) != 2 {
		t.Fatalf("expected 2 clusters for mixed scopes, got %d", len(result))
	}

	allowed := map[uint]bool{}
	for _, id := range result {
		allowed[id] = true
	}
	if !allowed[c1.ID] || !allowed[c2.ID] {
		t.Error("should contain c1 (from region) and c2 (direct)")
	}
	if allowed[c3.ID] {
		t.Error("should not contain c3")
	}
}

func TestAllowedClusterIDs_DuplicateDedup(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c1 := Cluster{Name: "c1", RegionID: r.ID}
	db.Create(&c1)

	user := User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	// Both region scope and direct cluster scope point to c1
	db.Create(&UserScope{UserID: user.ID, ScopeType: "region", ScopeID: r.ID})
	db.Create(&UserScope{UserID: user.ID, ScopeType: "cluster", ScopeID: c1.ID})

	result := AllowedClusterIDs(user.ID)
	if len(result) != 1 {
		t.Fatalf("expected 1 cluster after dedup, got %d", len(result))
	}
}

// ---------- AutoAssignCluster tests ----------

func TestAutoAssignCluster_MatchesByPriority(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c1 := Cluster{Name: "low-priority", RegionID: r.ID}
	c2 := Cluster{Name: "high-priority", RegionID: r.ID}
	db.Create(&c1)
	db.Create(&c2)

	// Both rules match web-*, but c2 has higher priority (lower number)
	db.Create(&ClusterMatchRule{ClusterID: c1.ID, Name: "r1", Priority: 10, HostnamePattern: "web-*", IsActive: true})
	db.Create(&ClusterMatchRule{ClusterID: c2.ID, Name: "r2", Priority: 1, HostnamePattern: "web-*", IsActive: true})

	result := AutoAssignCluster("web-01", "", "", "", "")
	if result == nil {
		t.Fatal("expected a cluster match")
	}
	if *result != c2.ID {
		t.Errorf("expected cluster %d (high priority), got %d", c2.ID, *result)
	}
}

func TestAutoAssignCluster_FallbackToDefault(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	defaultCluster := Cluster{Name: "default", RegionID: r.ID, IsDefault: true}
	db.Create(&defaultCluster)

	// No rules match "unknown-host"
	result := AutoAssignCluster("unknown-host", "1.2.3.4", "fluentbit", "linux", "")
	if result == nil {
		t.Fatal("expected fallback to default cluster")
	}
	if *result != defaultCluster.ID {
		t.Errorf("expected default cluster %d, got %d", defaultCluster.ID, *result)
	}
}

func TestAutoAssignCluster_NoMatchNoDefault(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c := Cluster{Name: "not-default", RegionID: r.ID, IsDefault: false}
	db.Create(&c)

	result := AutoAssignCluster("unknown-host", "", "", "", "")
	if result != nil {
		t.Error("expected nil when no match and no default cluster")
	}
}

func TestAutoAssignCluster_InactiveRulesSkipped(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c1 := Cluster{Name: "c1", RegionID: r.ID, IsDefault: false}
	db.Create(&c1)

	rule := ClusterMatchRule{ClusterID: c1.ID, Name: "inactive", Priority: 1, HostnamePattern: "web-*", IsActive: true}
	db.Create(&rule)
	// Set IsActive to false explicitly (GORM skips zero-value bools on Create due to default:true)
	db.Model(&rule).Update("is_active", false)

	// Verify no default cluster exists
	var defaultCount int64
	db.Model(&Cluster{}).Where("is_default = ?", true).Count(&defaultCount)
	if defaultCount != 0 {
		t.Fatalf("expected 0 default clusters, got %d", defaultCount)
	}

	// Verify no active rules
	var activeRules int64
	db.Model(&ClusterMatchRule{}).Where("is_active = ?", true).Count(&activeRules)
	if activeRules != 0 {
		t.Fatalf("expected 0 active rules, got %d", activeRules)
	}

	result := AutoAssignCluster("web-01", "", "", "", "")
	if result != nil {
		t.Errorf("inactive rules should be skipped, got cluster ID %d", *result)
	}
}

func TestAutoAssignCluster_IPMatching(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c1 := Cluster{Name: "private-net", RegionID: r.ID}
	db.Create(&c1)

	db.Create(&ClusterMatchRule{ClusterID: c1.ID, Name: "private", Priority: 1, IPPattern: "192.168.0.0/16", IsActive: true})

	result := AutoAssignCluster("any", "192.168.1.100", "", "", "")
	if result == nil {
		t.Fatal("expected match for IP in CIDR range")
	}
	if *result != c1.ID {
		t.Errorf("expected cluster %d, got %d", c1.ID, *result)
	}
}

func TestAutoAssignCluster_LabelMatching(t *testing.T) {
	db := setupTestDB(t)

	dc := DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c1 := Cluster{Name: "prod-web", RegionID: r.ID}
	db.Create(&c1)

	db.Create(&ClusterMatchRule{
		ClusterID:     c1.ID,
		Name:          "prod-web",
		Priority:      1,
		LabelSelector: `{"env":"prod","role":"web"}`,
		IsActive:      true,
	})

	result := AutoAssignCluster("any", "", "", "", `{"env":"prod","role":"web","team":"infra"}`)
	if result == nil {
		t.Fatal("expected match for labels")
	}
	if *result != c1.ID {
		t.Errorf("expected cluster %d, got %d", c1.ID, *result)
	}
}

// ---------- HashConfig tests ----------

func TestHashConfig(t *testing.T) {
	hash1 := HashConfig("test content")
	hash2 := HashConfig("test content")
	hash3 := HashConfig("different content")

	if hash1 != hash2 {
		t.Error("same content should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different content should produce different hash")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char hex SHA256 hash, got %d chars", len(hash1))
	}
}
