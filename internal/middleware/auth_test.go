package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var mwDBCounter int64

func setupMiddlewareDB(t *testing.T) *gorm.DB {
	t.Helper()
	mwDBCounter++
	dsn := fmt.Sprintf("file:mwtest%d?mode=memory&cache=shared", mwDBCounter)
	db, _ := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	db.AutoMigrate(
		&models.User{}, &models.Role{}, &models.Permission{}, &models.UserScope{},
		&models.DataCenter{}, &models.Region{}, &models.Cluster{}, &models.AgentAccessKey{},
	)
	models.DB = db
	return db
}

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------- ScopeFilter ----------

func TestScopeFilter_NoUserID(t *testing.T) {
	setupMiddlewareDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := ScopeFilter()
	handler(c)

	if c.IsAborted() {
		t.Error("should not abort when no user_id")
	}
}

func TestScopeFilter_GlobalAccess(t *testing.T) {
	db := setupMiddlewareDB(t)

	user := models.User{Username: "admin", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	handler := ScopeFilter()
	handler(c)

	result := GetAllowedClusters(c)
	if result != nil {
		t.Error("admin with no scopes should have nil (global access)")
	}
}

func TestScopeFilter_ScopedAccess(t *testing.T) {
	db := setupMiddlewareDB(t)

	dc := models.DataCenter{Name: "dc1"}
	db.Create(&dc)
	r := models.Region{Name: "r1", DataCenterID: dc.ID}
	db.Create(&r)
	c1 := models.Cluster{Name: "c1", RegionID: r.ID}
	c2 := models.Cluster{Name: "c2", RegionID: r.ID}
	db.Create(&c1)
	db.Create(&c2)

	user := models.User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	db.Create(&models.UserScope{UserID: user.ID, ScopeType: "cluster", ScopeID: c1.ID})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	handler := ScopeFilter()
	handler(c)

	result := GetAllowedClusters(c)
	if result == nil {
		t.Fatal("expected non-nil allowed clusters for scoped user")
	}
	if len(result) != 1 || result[0] != c1.ID {
		t.Errorf("expected only cluster %d, got %v", c1.ID, result)
	}
}

// ---------- GetAllowedClusters ----------

func TestGetAllowedClusters_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	result := GetAllowedClusters(c)
	if result != nil {
		t.Error("should return nil when not set")
	}
}

func TestGetAllowedClusters_Set(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("allowed_clusters", []uint{1, 2, 3})

	result := GetAllowedClusters(c)
	if len(result) != 3 {
		t.Errorf("expected 3 clusters, got %d", len(result))
	}
}

// ---------- RequirePermission ----------

func TestRequirePermission_NoAuth(t *testing.T) {
	setupMiddlewareDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := RequirePermission("nodes", "read")
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequirePermission_HasPermission(t *testing.T) {
	db := setupMiddlewareDB(t)

	perm := models.Permission{Name: "nodes:read", Resource: "nodes", Action: "read"}
	db.Create(&perm)
	role := models.Role{Name: "viewer"}
	db.Create(&role)
	db.Model(&role).Association("Permissions").Append(&perm)

	user := models.User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	db.Model(&user).Association("Roles").Append(&role)

	w := httptest.NewRecorder()
	c, r2 := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("user_id", user.ID)

	r2.GET("/test", RequirePermission("nodes", "read"), func(c *gin.Context) {
		c.Status(200)
	})
	r2.ServeHTTP(w, c.Request)

	// Also test the middleware directly
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/", nil)
	c2.Set("user_id", user.ID)

	handler := RequirePermission("nodes", "read")
	handler(c2)

	if c2.IsAborted() {
		t.Error("user with permission should not be aborted")
	}
}

func TestRequirePermission_LacksPermission(t *testing.T) {
	db := setupMiddlewareDB(t)

	role := models.Role{Name: "empty"}
	db.Create(&role)

	user := models.User{Username: "user1", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	db.Model(&user).Association("Roles").Append(&role)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("user_id", user.ID)

	handler := RequirePermission("nodes", "delete")
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// ---------- AgentAuth ----------

func TestAgentAuth_ValidKey(t *testing.T) {
	db := setupMiddlewareDB(t)
	svc := services.NewAgentAccessKeyService(db)
	result, err := svc.Create(services.AgentAccessKeyInput{Name: "test-key"}, 0, nil)
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", nil)
	c.Request.Header.Set("X-Agent-Key", result.PlaintextKey)

	handler := AgentAuth(svc, "")
	handler(c)

	if c.IsAborted() {
		t.Error("valid agent key should not be aborted")
	}
	authenticatedKey := GetAuthenticatedAgentKey(c)
	if authenticatedKey == nil || authenticatedKey.Name != "test-key" {
		t.Fatalf("expected authenticated key context, got %#v", authenticatedKey)
	}
}

func TestAgentAuth_InvalidKey(t *testing.T) {
	db := setupMiddlewareDB(t)
	svc := services.NewAgentAccessKeyService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", nil)
	c.Request.Header.Set("X-Agent-Key", "wrong-key")

	handler := AgentAuth(svc, "")
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAgentAuth_MissingKey(t *testing.T) {
	db := setupMiddlewareDB(t)
	svc := services.NewAgentAccessKeyService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", nil)

	handler := AgentAuth(svc, "")
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAgentAuth_LegacyKeyFallback(t *testing.T) {
	db := setupMiddlewareDB(t)
	svc := services.NewAgentAccessKeyService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", nil)
	c.Request.Header.Set("X-Agent-Key", "legacy-api-key")

	handler := AgentAuth(svc, "legacy-api-key")
	handler(c)

	if c.IsAborted() {
		t.Error("legacy config key should still be accepted")
	}
	authenticatedKey := GetAuthenticatedAgentKey(c)
	if authenticatedKey == nil || !authenticatedKey.Legacy {
		t.Fatalf("expected legacy authenticated key context, got %#v", authenticatedKey)
	}
}
