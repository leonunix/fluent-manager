package services

import (
	"errors"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupAgentAccessKeyTest(t *testing.T) (*gorm.DB, AgentAccessKeyService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewAgentAccessKeyService(db, "test-secret")
	return db, svc
}

func TestAgentAccessKeyCreateAndAuthenticate(t *testing.T) {
	_, svc := setupAgentAccessKeyTest(t)

	result, err := svc.Create(AgentAccessKeyInput{Name: "prod rollout"}, 1, nil)
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	if result.PlaintextKey == "" {
		t.Fatal("expected plaintext key to be returned once")
	}
	if result.Key == nil || result.Key.KeyPreview == "" {
		t.Fatalf("expected key metadata with preview, got %#v", result.Key)
	}

	authenticated, err := svc.Authenticate(result.PlaintextKey, "")
	if err != nil {
		t.Fatalf("authenticate key: %v", err)
	}
	if authenticated == nil || authenticated.Name != "prod rollout" {
		t.Fatalf("unexpected authenticated key: %#v", authenticated)
	}
}

func TestAgentAccessKeyScopedCreateRequiresCluster(t *testing.T) {
	db, svc := setupAgentAccessKeyTest(t)

	dc := models.DataCenter{Name: "dc-scope"}
	db.Create(&dc)
	region := models.Region{Name: "region-scope", DataCenterID: dc.ID}
	db.Create(&region)
	cluster := models.Cluster{Name: "cluster-scope", RegionID: region.ID}
	db.Create(&cluster)

	_, err := svc.Create(AgentAccessKeyInput{Name: "scoped without cluster"}, 1, []uint{cluster.ID})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
