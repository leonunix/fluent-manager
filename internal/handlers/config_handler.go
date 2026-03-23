package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	Svc services.ConfigService
}

func (h *ConfigHandler) ListTemplates(c *gin.Context) {
	fluentType := c.Query("fluent_type")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	templates, total, err := h.Svc.ListTemplates(fluentType, search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *ConfigHandler) GetTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	tpl, err := h.Svc.GetTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	FluentType  string `json:"fluent_type" binding:"required,oneof=fluentbit fluentd"`
	Content     string `json:"content" binding:"required"`
	Variables   string `json:"variables"`
}

func (h *ConfigHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	tpl, err := h.Svc.CreateTemplate(req.Name, req.Description, req.FluentType, req.Content, req.Variables, userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "template name already exists"})
		return
	}
	c.JSON(http.StatusCreated, tpl)
}

func (h *ConfigHandler) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tpl, err := h.Svc.UpdateTemplate(uint(id), req.Name, req.Description, req.FluentType, req.Content, req.Variables)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

func (h *ConfigHandler) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.Svc.DeleteTemplate(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

type CreateVersionRequest struct {
	Content string `json:"content" binding:"required"`
	Comment string `json:"comment"`
}

func (h *ConfigHandler) CreateVersion(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	version, err := h.Svc.CreateVersion(uint(templateID), userID, req.Content, req.Comment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusCreated, version)
}

func (h *ConfigHandler) ListVersions(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	versions, err := h.Svc.ListVersions(uint(templateID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": versions})
}

func (h *ConfigHandler) GetVersion(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 32)
	version, err := h.Svc.GetVersion(uint(versionID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}
	c.JSON(http.StatusOK, version)
}
