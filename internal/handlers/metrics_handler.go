package handlers

import (
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	Svc     services.MetricsService
	TopoSvc services.TopologyService // for DC ID resolution
}

func (h *MetricsHandler) Overview(c *gin.Context) {
	allowed := middleware.GetAllowedClusters(c)
	resp, err := h.Svc.Overview(allowed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MetricsHandler) TopNodes(c *gin.Context) {
	allowed := middleware.GetAllowedClusters(c)
	resp, err := h.Svc.TopNodes(allowed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *MetricsHandler) Throughput(c *gin.Context) {
	allowed := middleware.GetAllowedClusters(c)
	resp, err := h.Svc.Throughput(allowed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MetricsHandler) ByDatacenter(c *gin.Context) {
	allowed := middleware.GetAllowedClusters(c)
	var allowedDCIDs []uint
	if allowed != nil && h.TopoSvc != nil {
		allowedDCIDs = h.TopoSvc.AllowedDCIDs(allowed)
	}
	resp, err := h.Svc.ByDatacenter(allowedDCIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
