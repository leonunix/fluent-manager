package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct{}

// --- Config Templates ---

func (h *ConfigHandler) ListTemplates(c *gin.Context) {
	var templates []models.ConfigTemplate
	query := models.DB.Preload("Creator")

	if fluentType := c.Query("fluent_type"); fluentType != "" {
		query = query.Where("fluent_type = ?", fluentType)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Model(&models.ConfigTemplate{}).Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates)

	c.JSON(http.StatusOK, gin.H{
		"data":      templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *ConfigHandler) GetTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var tpl models.ConfigTemplate
	if err := models.DB.Preload("Creator").Preload("Versions").First(&tpl, id).Error; err != nil {
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
	tpl := models.ConfigTemplate{
		Name:        req.Name,
		Description: req.Description,
		FluentType:  req.FluentType,
		Content:     req.Content,
		Variables:   req.Variables,
		CreatedBy:   userID,
	}

	if err := models.DB.Create(&tpl).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "template name already exists"})
		return
	}
	c.JSON(http.StatusCreated, tpl)
}

func (h *ConfigHandler) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var tpl models.ConfigTemplate
	if err := models.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&tpl).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"fluent_type": req.FluentType,
		"content":     req.Content,
		"variables":   req.Variables,
	})
	c.JSON(http.StatusOK, tpl)
}

func (h *ConfigHandler) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var tpl models.ConfigTemplate
	if err := models.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	models.DB.Where("template_id = ?", id).Delete(&models.ConfigVersion{})
	models.DB.Delete(&tpl)
	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

// --- Config Versions ---

type CreateVersionRequest struct {
	Content string `json:"content" binding:"required"`
	Comment string `json:"comment"`
}

func (h *ConfigHandler) CreateVersion(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var tpl models.ConfigTemplate
	if err := models.DB.First(&tpl, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get next version number
	var maxVersion int
	models.DB.Model(&models.ConfigVersion{}).
		Where("template_id = ?", templateID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	userID := c.GetUint("user_id")
	version := models.ConfigVersion{
		TemplateID: uint(templateID),
		Version:    maxVersion + 1,
		Content:    req.Content,
		Hash:       models.HashConfig(req.Content),
		Comment:    req.Comment,
		CreatedBy:  userID,
	}
	models.DB.Create(&version)

	c.JSON(http.StatusCreated, version)
}

func (h *ConfigHandler) ListVersions(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var versions []models.ConfigVersion
	models.DB.Where("template_id = ?", templateID).
		Preload("Creator").
		Order("version DESC").
		Find(&versions)
	c.JSON(http.StatusOK, gin.H{"data": versions})
}

func (h *ConfigHandler) GetVersion(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 32)
	var version models.ConfigVersion
	if err := models.DB.Preload("Template").Preload("Creator").First(&version, versionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}
	c.JSON(http.StatusOK, version)
}
