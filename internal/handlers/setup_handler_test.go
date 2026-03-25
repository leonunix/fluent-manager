package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

func newSetupTestRouter(handler *SetupHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/setup", handler.Initialize)
	return router
}

func writeSetupTestConfig(t *testing.T, dir, secret string) string {
	t.Helper()

	path := filepath.Join(dir, "config.yaml")
	content := []byte("server:\n  host: \"0.0.0.0\"\n  port: 8080\nauth:\n  jwt_secret: \"" + secret + "\"\n  token_expire_hours: 24\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestSetupInitialize_DefaultSQLiteSecretChangeTriggersRestartAndReturnsValidToken(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "setup.db")
	db, err := models.InitDBWithConn("sqlite", dbPath)
	if err != nil {
		t.Fatalf("init test db: %v", err)
	}
	models.DB = db

	configPath := writeSetupTestConfig(t, t.TempDir(), "change-me-in-production")
	jwtSvc := auth.NewJWTService("change-me-in-production", 24)
	handler := &SetupHandler{
		Svc:       services.NewSetupService(db),
		JWT:       jwtSvc,
		CfgPath:   configPath,
		RestartCh: make(chan struct{}, 1),
	}
	router := newSetupTestRouter(handler)

	body := map[string]string{
		"username": "admin",
		"password": "admin1234",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/setup", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Restart bool   `json:"restart"`
		Token   string `json:"token"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Restart {
		t.Fatalf("expected restart=true when jwt secret changes")
	}
	if payload.Token == "" {
		t.Fatalf("expected token in setup response")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load updated config: %v", err)
	}
	if cfg.Auth.JWTSecret == "" || cfg.Auth.JWTSecret == "change-me-in-production" {
		t.Fatalf("expected setup to persist a non-placeholder jwt secret")
	}

	if _, err := auth.NewJWTService("change-me-in-production", 24).ParseToken(payload.Token); err == nil {
		t.Fatalf("expected token to be invalid for the old jwt secret")
	}
	claims, err := auth.NewJWTService(cfg.Auth.JWTSecret, 24).ParseToken(payload.Token)
	if err != nil {
		t.Fatalf("expected token to be valid for updated jwt secret: %v", err)
	}
	if claims.Username != "admin" {
		t.Fatalf("expected token for admin, got %q", claims.Username)
	}

	select {
	case <-handler.RestartCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected restart signal after setup secret rotation")
	}
}
