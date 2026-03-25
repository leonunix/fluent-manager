package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgentArtifactHandler struct {
	Svc services.AgentArtifactService
}

func (h *AgentArtifactHandler) List(c *gin.Context) {
	artifacts, err := h.Svc.List()
	if err != nil {
		writeAgentArtifactError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": artifacts})
}

func (h *AgentArtifactHandler) Create(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	artifact, err := h.Svc.Create(services.AgentArtifactInput{
		Name:        c.PostForm("name"),
		Version:     c.PostForm("version"),
		Description: c.PostForm("description"),
	}, fileHeader, c.GetUint("user_id"))
	if err != nil {
		writeAgentArtifactError(c, err)
		return
	}
	middleware.SetAuditResource(c, "agent_artifact", artifact.ID)
	c.JSON(http.StatusCreated, artifact)
}

func (h *AgentArtifactHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Svc.Delete(id); err != nil {
		writeAgentArtifactError(c, err)
		return
	}
	middleware.SetAuditResource(c, "agent_artifact", id)
	c.JSON(http.StatusOK, gin.H{"message": "agent artifact deleted"})
}

func (h *AgentArtifactHandler) DownloadForAgent(c *gin.Context) {
	rawID := c.Param("id")
	id64, err := strconv.ParseUint(rawID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid artifact id"})
		return
	}

	artifact, err := h.Svc.Get(uint(id64))
	if err != nil {
		writeAgentArtifactError(c, err)
		return
	}
	c.FileAttachment(artifact.StoragePath, artifact.FileName)
}

func writeAgentArtifactError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "agent artifact not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
