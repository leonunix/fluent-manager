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
	}, 1); err != nil {
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
	}, 1); err != nil {
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
	}, 1); err != nil {
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
	}, 1); err != nil {
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
	}, 1); err != nil {
		t.Fatalf("create first global policy: %v", err)
	}
	if _, err := svc.CreatePolicy(&AgentPolicyInput{
		Name:      "global-b",
		ScopeType: models.AgentPolicyScopeGlobal,
		IsEnabled: enabled,
	}, 1); err == nil {
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
	}, 1); err != nil {
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

