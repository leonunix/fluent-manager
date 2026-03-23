package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func newFluentTestRouter(handler *FluentHandler, allowedClusters []uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if allowedClusters != nil {
			c.Set("allowed_clusters", allowedClusters)
		}
		c.Next()
	})
	router.POST("/aggregation-groups", handler.CreateAggregationGroup)
	router.GET("/aggregation-groups", handler.ListAggregationGroups)
	router.GET("/aggregation-groups/:id", handler.GetAggregationGroup)
	router.PUT("/aggregation-groups/:id", handler.UpdateAggregationGroup)
	router.DELETE("/aggregation-groups/:id", handler.DeleteAggregationGroup)
	router.POST("/aggregation-groups/:id/restore", handler.RestoreAggregationGroup)
	router.GET("/nodes/:id/fluent-profile", handler.GetNodeProfile)
	router.PUT("/nodes/:id/fluent-profile", handler.UpdateNodeProfile)
	return router
}

func seedHandlerCluster(t *testing.T, db *gorm.DB, suffix string) models.Cluster {
	t.Helper()

	dc := models.DataCenter{Name: "handler-dc-" + suffix}
	if err := db.Create(&dc).Error; err != nil {
		t.Fatalf("create datacenter: %v", err)
	}

	region := models.Region{Name: "handler-region-" + suffix, DataCenterID: dc.ID}
	if err := db.Create(&region).Error; err != nil {
		t.Fatalf("create region: %v", err)
	}

	cluster := models.Cluster{Name: "handler-cluster-" + suffix, RegionID: region.ID}
	if err := db.Create(&cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	return cluster
}

func TestCreateAggregationGroupReturnsBadRequestForInvalidCluster(t *testing.T) {
	db := testutil.NewTestDB()
	models.DB = db
	handler := &FluentHandler{Svc: services.NewFluentService(db, "handler-test-secret")}
	router := newFluentTestRouter(handler, nil)

	body := map[string]interface{}{
		"name":       "agg-invalid-cluster",
		"cluster_id": 999999,
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/aggregation-groups", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestCreateAggregationGroupReturnsConflictForDuplicateName(t *testing.T) {
	db := testutil.NewTestDB()
	models.DB = db
	cluster := seedHandlerCluster(t, db, "dup")
	handler := &FluentHandler{Svc: services.NewFluentService(db, "handler-test-secret")}
	router := newFluentTestRouter(handler, nil)

	first := map[string]interface{}{
		"name":       "agg-dup",
		"cluster_id": cluster.ID,
	}
	raw, _ := json.Marshal(first)
	req := httptest.NewRequest(http.MethodPost, "/aggregation-groups", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected first create 201, got %d body=%s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/aggregation-groups", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestScopedAggregationGroupEndpointsHideOutOfScopeGroups(t *testing.T) {
	db := testutil.NewTestDB()
	models.DB = db
	clusterA := seedHandlerCluster(t, db, "scope-a")
	clusterB := seedHandlerCluster(t, db, "scope-b")

	fluentSvc := services.NewFluentService(db, "handler-test-secret")
	groupA, err := fluentSvc.CreateAggregationGroup(&services.AggregationGroupInput{
		Name:      "agg-scope-a",
		ClusterID: &clusterA.ID,
	})
	if err != nil {
		t.Fatalf("create group a: %v", err)
	}
	groupB, err := fluentSvc.CreateAggregationGroup(&services.AggregationGroupInput{
		Name:      "agg-scope-b",
		ClusterID: &clusterB.ID,
	})
	if err != nil {
		t.Fatalf("create group b: %v", err)
	}

	handler := &FluentHandler{Svc: fluentSvc}
	router := newFluentTestRouter(handler, []uint{clusterA.ID})

	listResp := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/aggregation-groups", nil)
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	if !bytes.Contains(listResp.Body.Bytes(), []byte(groupA.Name)) {
		t.Fatalf("expected in-scope group in response, got %s", listResp.Body.String())
	}
	if bytes.Contains(listResp.Body.Bytes(), []byte(groupB.Name)) {
		t.Fatalf("expected out-of-scope group to be filtered out, got %s", listResp.Body.String())
	}

	getResp := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/aggregation-groups/"+strconv.FormatUint(uint64(groupB.ID), 10), nil)
	router.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for out-of-scope group, got %d body=%s", getResp.Code, getResp.Body.String())
	}
}

func TestAggregationGroupEndpointsRejectInvalidID(t *testing.T) {
	db := testutil.NewTestDB()
	models.DB = db
	handler := &FluentHandler{
		Svc:     services.NewFluentService(db, "handler-test-secret"),
		NodeSvc: services.NewNodeService(db),
	}
	router := newFluentTestRouter(handler, nil)

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/aggregation-groups/abc"},
		{method: http.MethodPut, path: "/aggregation-groups/abc", body: `{"name":"agg"}`},
		{method: http.MethodDelete, path: "/aggregation-groups/abc"},
		{method: http.MethodPost, path: "/aggregation-groups/abc/restore"},
		{method: http.MethodGet, path: "/nodes/abc/fluent-profile"},
		{method: http.MethodPut, path: "/nodes/abc/fluent-profile", body: `{}`},
	}

	for _, tc := range requests {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("%s %s expected 400, got %d body=%s", tc.method, tc.path, resp.Code, resp.Body.String())
		}
	}
}
