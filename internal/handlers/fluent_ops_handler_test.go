package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"github.com/gin-gonic/gin"
)

func newFluentOpsTestRouter(handler *FluentOpsHandler, allowedClusters []uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if allowedClusters != nil {
			c.Set("allowed_clusters", allowedClusters)
		}
		c.Next()
	})
	router.GET("/pipelines/:id", handler.GetPipeline)
	router.PUT("/pipelines/:id", handler.UpdatePipeline)
	router.DELETE("/pipelines/:id", handler.DeletePipeline)
	router.GET("/analysis/:id", handler.GetAnalysisResult)
	router.GET("/aggregation-groups/:id/metrics", handler.AggregationGroupMetrics)
	return router
}

func TestFluentOpsEndpointsRejectInvalidID(t *testing.T) {
	db := testutil.NewTestDB()
	handler := &FluentOpsHandler{Svc: services.NewFluentOpsService(db)}
	router := newFluentOpsTestRouter(handler, nil)

	requests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/pipelines/abc"},
		{method: http.MethodPut, path: "/pipelines/abc"},
		{method: http.MethodDelete, path: "/pipelines/abc"},
		{method: http.MethodGet, path: "/analysis/abc"},
		{method: http.MethodGet, path: "/aggregation-groups/abc/metrics"},
	}

	for _, tc := range requests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("%s %s expected 400, got %d body=%s", tc.method, tc.path, resp.Code, resp.Body.String())
		}
	}
}
