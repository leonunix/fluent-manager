package handlers

import (
	"errors"
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FluentOpsHandler struct {
	Svc services.FluentOpsService
}

func (h *FluentOpsHandler) ListOutputTargets(c *gin.Context) {
	targets, err := h.Svc.ListOutputTargets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": targets})
}

func (h *FluentOpsHandler) GetOutputTarget(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	target, err := h.Svc.GetOutputTarget(id)
	if err != nil {
		writeFluentOpsError(c, err, "output target")
		return
	}
	c.JSON(http.StatusOK, target)
}

func (h *FluentOpsHandler) CreateOutputTarget(c *gin.Context) {
	var req services.OutputTargetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	target, err := h.Svc.CreateOutputTarget(&req, c.GetUint("user_id"))
	if err != nil {
		writeFluentOpsError(c, err, "output target")
		return
	}
	c.JSON(http.StatusCreated, target)
}

func (h *FluentOpsHandler) UpdateOutputTarget(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req services.OutputTargetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	target, err := h.Svc.UpdateOutputTarget(id, &req)
	if err != nil {
		writeFluentOpsError(c, err, "output target")
		return
	}
	c.JSON(http.StatusOK, target)
}

func (h *FluentOpsHandler) DeleteOutputTarget(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Svc.DeleteOutputTarget(id); err != nil {
		writeFluentOpsError(c, err, "output target")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "output target deleted"})
}

func (h *FluentOpsHandler) ListPipelines(c *gin.Context) {
	pipelines, err := h.Svc.ListPipelines(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pipelines})
}

func (h *FluentOpsHandler) GetPipeline(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	pipeline, err := h.Svc.GetPipeline(id, middleware.GetAllowedClusters(c))
	if err != nil {
		writeFluentOpsError(c, err, "pipeline")
		return
	}
	c.JSON(http.StatusOK, pipeline)
}

func (h *FluentOpsHandler) CreatePipeline(c *gin.Context) {
	var req services.LogPipelineInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pipeline, err := h.Svc.CreatePipeline(&req, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeFluentOpsError(c, err, "pipeline")
		return
	}
	c.JSON(http.StatusCreated, pipeline)
}

func (h *FluentOpsHandler) UpdatePipeline(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req services.LogPipelineInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pipeline, err := h.Svc.UpdatePipeline(id, &req, middleware.GetAllowedClusters(c))
	if err != nil {
		writeFluentOpsError(c, err, "pipeline")
		return
	}
	c.JSON(http.StatusOK, pipeline)
}

func (h *FluentOpsHandler) DeletePipeline(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Svc.DeletePipeline(id, middleware.GetAllowedClusters(c)); err != nil {
		writeFluentOpsError(c, err, "pipeline")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pipeline deleted"})
}

func (h *FluentOpsHandler) PipelineGraph(c *gin.Context) {
	graph, err := h.Svc.PipelineGraph(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, graph)
}

func (h *FluentOpsHandler) RuntimeHealthGraph(c *gin.Context) {
	graph, err := h.Svc.RuntimeHealthGraph(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, graph)
}

func (h *FluentOpsHandler) LintConfig(c *gin.Context) {
	var req services.ConfigLintInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Svc.LintConfig(&req, c.GetUint("user_id"))
	if err != nil {
		writeFluentOpsError(c, err, "analysis")
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *FluentOpsHandler) ReplayConfig(c *gin.Context) {
	var req services.ConfigReplayInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Svc.ReplayConfig(&req)
	if err != nil {
		writeFluentOpsError(c, err, "analysis")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *FluentOpsHandler) SemanticDiff(c *gin.Context) {
	var req services.ConfigSemanticDiffInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Svc.SemanticDiff(&req)
	if err != nil {
		writeFluentOpsError(c, err, "analysis")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *FluentOpsHandler) CheckCompatibility(c *gin.Context) {
	var req services.CompatibilityCheckInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Svc.CheckCompatibility(&req, middleware.GetAllowedClusters(c))
	if err != nil {
		writeFluentOpsError(c, err, "analysis")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *FluentOpsHandler) GetAnalysisResult(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.Svc.GetAnalysisResult(id)
	if err != nil {
		writeFluentOpsError(c, err, "analysis")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *FluentOpsHandler) RuntimeDrift(c *gin.Context) {
	items, err := h.Svc.ListRuntimeDrift(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *FluentOpsHandler) RuntimeRecommendations(c *gin.Context) {
	items, err := h.Svc.RuntimeRecommendations(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *FluentOpsHandler) AggregationGroupMetrics(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	metric, err := h.Svc.AggregationGroupMetrics(id, middleware.GetAllowedClusters(c))
	if err != nil {
		writeFluentOpsError(c, err, "aggregation group")
		return
	}
	c.JSON(http.StatusOK, metric)
}

func writeFluentOpsError(c *gin.Context, err error, resource string) {
	switch {
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": resource + " not in your scope"})
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": resource + " not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
