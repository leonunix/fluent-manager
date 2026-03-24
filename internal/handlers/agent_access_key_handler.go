package handlers

import (
	"errors"
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgentAccessKeyHandler struct {
	Svc services.AgentAccessKeyService
}

func (h *AgentAccessKeyHandler) List(c *gin.Context) {
	keys, err := h.Svc.List(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

func (h *AgentAccessKeyHandler) Create(c *gin.Context) {
	var req services.AgentAccessKeyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.Svc.Create(req, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentAccessKeyError(c, err)
		return
	}
	middleware.SetAuditResource(c, "agent_access_key", result.Key.ID)
	c.JSON(http.StatusCreated, result)
}

func (h *AgentAccessKeyHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req services.AgentAccessKeyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.Svc.Update(id, req, middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentAccessKeyError(c, err)
		return
	}
	middleware.SetAuditResource(c, "agent_access_key", key.ID)
	c.JSON(http.StatusOK, key)
}

func (h *AgentAccessKeyHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Svc.Delete(id, middleware.GetAllowedClusters(c)); err != nil {
		writeAgentAccessKeyError(c, err)
		return
	}
	middleware.SetAuditResource(c, "agent_access_key", id)
	c.JSON(http.StatusOK, gin.H{"message": "agent access key deleted"})
}

func writeAgentAccessKeyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "agent access key not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
