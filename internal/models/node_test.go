package models

import "testing"

func TestEffectiveConfigID_NodeLevel(t *testing.T) {
	nodeConfigID := uint(10)
	clusterConfigID := uint(20)
	node := Node{
		ConfigID: &nodeConfigID,
		Cluster:  &Cluster{ConfigID: &clusterConfigID},
	}
	result := node.EffectiveConfigID()
	if result == nil || *result != 10 {
		t.Error("node-level config should take precedence")
	}
}

func TestEffectiveConfigID_ClusterLevel(t *testing.T) {
	clusterConfigID := uint(20)
	node := Node{
		ConfigID: nil,
		Cluster:  &Cluster{ConfigID: &clusterConfigID},
	}
	result := node.EffectiveConfigID()
	if result == nil || *result != 20 {
		t.Error("should fall back to cluster config")
	}
}

func TestEffectiveConfigID_NoConfig(t *testing.T) {
	node := Node{ConfigID: nil, Cluster: &Cluster{ConfigID: nil}}
	if node.EffectiveConfigID() != nil {
		t.Error("should return nil when no config at any level")
	}
}

func TestEffectiveConfigID_NoCluster(t *testing.T) {
	node := Node{ConfigID: nil, Cluster: nil}
	if node.EffectiveConfigID() != nil {
		t.Error("should return nil when no cluster")
	}
}

func TestEffectiveEnvironmentID_NodeLevel(t *testing.T) {
	nodeEnvID := uint(1)
	clusterEnvID := uint(2)
	node := Node{
		EnvironmentID: &nodeEnvID,
		Cluster:       &Cluster{EnvironmentID: &clusterEnvID},
	}
	result := node.EffectiveEnvironmentID()
	if result == nil || *result != 1 {
		t.Error("node-level env should take precedence")
	}
}

func TestEffectiveEnvironmentID_ClusterLevel(t *testing.T) {
	clusterEnvID := uint(2)
	node := Node{
		EnvironmentID: nil,
		Cluster:       &Cluster{EnvironmentID: &clusterEnvID},
	}
	result := node.EffectiveEnvironmentID()
	if result == nil || *result != 2 {
		t.Error("should fall back to cluster env")
	}
}

func TestEffectiveEnvironmentID_None(t *testing.T) {
	node := Node{EnvironmentID: nil, Cluster: nil}
	if node.EffectiveEnvironmentID() != nil {
		t.Error("should return nil when no env")
	}
}
