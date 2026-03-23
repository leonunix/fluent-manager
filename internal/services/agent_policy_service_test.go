package services

import (
	"errors"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
)

func TestAgentPolicyResolveAppliesScopeOverrides(t *testing.T) {
	db := testutil.NewTestDB()
	svc := NewAgentPolicyService(db, AgentSettings{
		HeartbeatInterval: 30,
		MetricsInterval:   60,
		FluentType:        "fluentbit",
	})

	env := models.Environment{Name: "prod"}
	if err := db.Create(&env).Error; err != nil {
		t.Fatalf("create env: %v", err)
	}
	dc := models.DataCenter{Name: "dc1"}
	db.Create(&dc)
	region := models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&region)
	cluster := models.Cluster{Name: "c1", RegionID: region.ID, EnvironmentID: &env.ID}
	db.Create(&cluster)
	node := models.Node{
		NodeUID:   "node-1",
		Hostname:  "n1",
		ClusterID: &cluster.ID,
		Labels:    `{"env":"prod","team":"infra"}`,
	}
	db.Create(&node)

	enabled := true
	heartbeat50 := 50
	metrics90 := 90
	logPath := "/var/log/custom.log"
	files := []string{"/etc/fluent/custom-extra.conf"}

	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "global",
		ScopeType: models.AgentPolicyScopeGlobal,
		IsEnabled: enabled,
		Settings: AgentSettingsPatch{
			HeartbeatInterval: &heartbeat50,
		},
	}, 1, nil); err != nil {
		t.Fatalf("create global policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:          "env-prod",
		ScopeType:     models.AgentPolicyScopeEnvironment,
		EnvironmentID: &env.ID,
		IsEnabled:     enabled,
		Priority:      20,
		Settings: AgentSettingsPatch{
			MetricsInterval: &metrics90,
		},
	}, 1, nil); err != nil {
		t.Fatalf("create env policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "cluster",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &cluster.ID,
		IsEnabled: enabled,
		Priority:  30,
		Settings: AgentSettingsPatch{
			FluentLogPath: &logPath,
		},
	}, 1, nil); err != nil {
		t.Fatalf("create cluster policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:          "labels",
		ScopeType:     models.AgentPolicyScopeLabelSelector,
		LabelSelector: `{"team":"infra"}`,
		IsEnabled:     enabled,
		Priority:      40,
		Settings: AgentSettingsPatch{
			FluentExtraFiles: &files,
		},
	}, 1, nil); err != nil {
		t.Fatalf("create label policy: %v", err)
	}

	resolved, err := svc.ResolveForNodeID(node.ID)
	if err != nil {
		t.Fatalf("resolve policy: %v", err)
	}
	if resolved.Settings.HeartbeatInterval != 50 {
		t.Fatalf("expected global heartbeat override, got %d", resolved.Settings.HeartbeatInterval)
	}
	if resolved.Settings.MetricsInterval != 90 {
		t.Fatalf("expected env metrics override, got %d", resolved.Settings.MetricsInterval)
	}
	if resolved.Settings.FluentLogPath != "/var/log/custom.log" {
		t.Fatalf("expected cluster fluent log path override, got %q", resolved.Settings.FluentLogPath)
	}
	if len(resolved.Settings.FluentExtraFiles) != 1 || resolved.Settings.FluentExtraFiles[0] != files[0] {
		t.Fatalf("expected label selector extra files override, got %#v", resolved.Settings.FluentExtraFiles)
	}
	if len(resolved.MatchedPolicies) != 4 {
		t.Fatalf("expected 4 matched policies, got %d", len(resolved.MatchedPolicies))
	}
}

func TestAgentPolicyCreateRejectsDuplicateGlobalPolicy(t *testing.T) {
	db := testutil.NewTestDB()
	svc := NewAgentPolicyService(db, AgentSettings{})

	enabled := true
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "global-a",
		ScopeType: models.AgentPolicyScopeGlobal,
		IsEnabled: enabled,
	}, 1, nil); err != nil {
		t.Fatalf("create first global policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "global-b",
		ScopeType: models.AgentPolicyScopeGlobal,
		IsEnabled: enabled,
	}, 1, nil); err == nil {
		t.Fatal("expected duplicate global policy to fail")
	} else if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestAgentPolicyResolveUsesClusterEnvironmentWhenNodeEnvironmentUnset(t *testing.T) {
	db := testutil.NewTestDB()
	svc := NewAgentPolicyService(db, AgentSettings{MetricsInterval: 60})

	env := models.Environment{Name: "prod"}
	db.Create(&env)
	dc := models.DataCenter{Name: "dc2"}
	db.Create(&dc)
	region := models.Region{Name: "r2", DataCenterID: dc.ID}
	db.Create(&region)
	cluster := models.Cluster{Name: "c2", RegionID: region.ID, EnvironmentID: &env.ID}
	db.Create(&cluster)
	node := models.Node{NodeUID: "uid2", Hostname: "n2", ClusterID: &cluster.ID}
	db.Create(&node)

	enabled := true
	metrics120 := 120
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:          "env",
		ScopeType:     models.AgentPolicyScopeEnvironment,
		EnvironmentID: &env.ID,
		IsEnabled:     enabled,
		Settings: AgentSettingsPatch{
			MetricsInterval: &metrics120,
		},
	}, 1, nil); err != nil {
		t.Fatalf("create env policy: %v", err)
	}

	resolved, err := svc.ResolveForNodeID(node.ID)
	if err != nil {
		t.Fatalf("resolve env policy: %v", err)
	}
	if resolved.Settings.MetricsInterval != 120 {
		t.Fatalf("expected env override to use cluster environment, got %d", resolved.Settings.MetricsInterval)
	}
}

func TestAgentPolicyListPoliciesFiltersReadableScope(t *testing.T) {
	db := testutil.NewTestDB()
	svc := NewAgentPolicyService(db, AgentSettings{})

	envA := models.Environment{Name: "prod-a"}
	envB := models.Environment{Name: "prod-b"}
	db.Create(&envA)
	db.Create(&envB)

	dc := models.DataCenter{Name: "dc-scope"}
	db.Create(&dc)
	region := models.Region{Name: "r-scope", DataCenterID: dc.ID}
	db.Create(&region)

	clusterA := models.Cluster{Name: "cluster-a", RegionID: region.ID, EnvironmentID: &envA.ID}
	clusterB := models.Cluster{Name: "cluster-b", RegionID: region.ID, EnvironmentID: &envB.ID}
	db.Create(&clusterA)
	db.Create(&clusterB)

	db.Create(&models.Node{
		NodeUID:   "node-a",
		Hostname:  "node-a",
		ClusterID: &clusterA.ID,
		Labels:    `{"team":"infra"}`,
	})

	enabled := true
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "global",
		ScopeType: models.AgentPolicyScopeGlobal,
		IsEnabled: enabled,
	}, 1, nil); err != nil {
		t.Fatalf("create global: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:          "env-a",
		ScopeType:     models.AgentPolicyScopeEnvironment,
		EnvironmentID: &envA.ID,
		IsEnabled:     enabled,
		Priority:      10,
	}, 1, nil); err != nil {
		t.Fatalf("create env policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "cluster-a",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &clusterA.ID,
		IsEnabled: enabled,
		Priority:  20,
	}, 1, nil); err != nil {
		t.Fatalf("create cluster-a policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "cluster-b",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &clusterB.ID,
		IsEnabled: enabled,
		Priority:  30,
	}, 1, nil); err != nil {
		t.Fatalf("create cluster-b policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:          "labels",
		ScopeType:     models.AgentPolicyScopeLabelSelector,
		LabelSelector: `{"team":"infra"}`,
		IsEnabled:     enabled,
		Priority:      40,
	}, 1, nil); err != nil {
		t.Fatalf("create labels policy: %v", err)
	}

	policies, err := svc.ListPolicies([]uint{clusterA.ID})
	if err != nil {
		t.Fatalf("list policies: %v", err)
	}
	if len(policies) != 4 {
		t.Fatalf("expected 4 visible policies, got %d", len(policies))
	}
	for _, policy := range policies {
		if policy.Name == "cluster-b" {
			t.Fatalf("out-of-scope cluster policy should not be visible: %+v", policy)
		}
	}
}

func TestAgentPolicyScopedUsersCanOnlyManageClusterScopedPolicies(t *testing.T) {
	db := testutil.NewTestDB()
	svc := NewAgentPolicyService(db, AgentSettings{})

	dc := models.DataCenter{Name: "dc-write"}
	db.Create(&dc)
	region := models.Region{Name: "r-write", DataCenterID: dc.ID}
	db.Create(&region)
	clusterA := models.Cluster{Name: "cluster-a", RegionID: region.ID}
	clusterB := models.Cluster{Name: "cluster-b", RegionID: region.ID}
	db.Create(&clusterA)
	db.Create(&clusterB)

	enabled := true
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "cluster-a",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &clusterA.ID,
		IsEnabled: enabled,
	}, 1, []uint{clusterA.ID}); err != nil {
		t.Fatalf("scoped cluster policy should be allowed: %v", err)
	}

	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "global",
		ScopeType: models.AgentPolicyScopeGlobal,
		IsEnabled: enabled,
	}, 1, []uint{clusterA.ID}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for global scoped write, got %v", err)
	}

	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "cluster-b",
		ScopeType: models.AgentPolicyScopeCluster,
		ClusterID: &clusterB.ID,
		IsEnabled: enabled,
	}, 1, []uint{clusterA.ID}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for out-of-scope cluster write, got %v", err)
	}
}
