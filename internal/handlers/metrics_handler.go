package handlers

import (
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	Svc services.MetricsService
}

func (h *MetricsHandler) Overview(c *gin.Context) {
	resp, err := h.Svc.Overview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MetricsHandler) TopNodes(c *gin.Context) {
	resp, err := h.Svc.TopNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *MetricsHandler) ByDatacenter(c *gin.Context) {
	resp, err := h.Svc.ByDatacenter()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
