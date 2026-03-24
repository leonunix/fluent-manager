package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	FluentType    string `json:"fluent_type" binding:"required,oneof=fluentbit fluentd"`
	Content       string `json:"content" binding:"required"`
	Variables     string `json:"variables"`
	SourceType    string `json:"source_type"`
	SourceModules string `json:"source_modules"`
	FlowLayout    string `json:"flow_layout"`
}

func (h *ConfigHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	tpl, err := h.Svc.CreateTemplate(&services.ConfigTemplateInput{
		Name:          req.Name,
		Description:   req.Description,
		FluentType:    req.FluentType,
		Content:       req.Content,
		Variables:     req.Variables,
		SourceType:    req.SourceType,
		SourceModules: req.SourceModules,
		FlowLayout:    req.FlowLayout,
	}, userID)
	if err != nil {
		writeConfigError(c, err, "template")
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

	tpl, err := h.Svc.UpdateTemplate(uint(id), &services.ConfigTemplateInput{
		Name:          req.Name,
		Description:   req.Description,
		FluentType:    req.FluentType,
		Content:       req.Content,
		Variables:     req.Variables,
		SourceType:    req.SourceType,
		SourceModules: req.SourceModules,
		FlowLayout:    req.FlowLayout,
	})
	if err != nil {
		writeConfigError(c, err, "template")
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

func (h *ConfigHandler) ListModules(c *gin.Context) {
	fluentType := c.Query("fluent_type")
	moduleType := c.Query("module_type")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	modules, total, err := h.Svc.ListModules(fluentType, moduleType, search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      modules,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *ConfigHandler) GetModule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	module, err := h.Svc.GetModule(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return
	}
	c.JSON(http.StatusOK, module)
}

type ConfigModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ModuleType  string `json:"module_type" binding:"required"`
	FluentType  string `json:"fluent_type" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Variables   string `json:"variables"`
	IsBuiltin   bool   `json:"is_builtin"`
	PresetKind  string `json:"preset_kind"`
	PresetKey   string `json:"preset_key"`
}

type DeleteModulesRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

func (h *ConfigHandler) CreateModule(c *gin.Context) {
	var req ConfigModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	module, err := h.Svc.CreateModule(&services.ConfigModuleInput{
		Name:        req.Name,
		Description: req.Description,
		ModuleType:  req.ModuleType,
		FluentType:  req.FluentType,
		Content:     req.Content,
		Variables:   req.Variables,
		IsBuiltin:   req.IsBuiltin,
		PresetKind:  req.PresetKind,
		PresetKey:   req.PresetKey,
	}, c.GetUint("user_id"))
	if err != nil {
		writeConfigError(c, err, "module")
		return
	}
	c.JSON(http.StatusCreated, module)
}

func (h *ConfigHandler) UpdateModule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req ConfigModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	module, err := h.Svc.UpdateModule(uint(id), &services.ConfigModuleInput{
		Name:        req.Name,
		Description: req.Description,
		ModuleType:  req.ModuleType,
		FluentType:  req.FluentType,
		Content:     req.Content,
		Variables:   req.Variables,
		IsBuiltin:   req.IsBuiltin,
		PresetKind:  req.PresetKind,
		PresetKey:   req.PresetKey,
	})
	if err != nil {
		writeConfigError(c, err, "module")
		return
	}
	c.JSON(http.StatusOK, module)
}

func (h *ConfigHandler) DeleteModule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.Svc.DeleteModule(uint(id)); err != nil {
		writeConfigError(c, err, "module")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "module deleted"})
}

func (h *ConfigHandler) DeleteModules(c *gin.Context) {
	var req DeleteModulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Svc.DeleteModules(req.IDs); err != nil {
		writeConfigError(c, err, "module")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "modules deleted",
		"count":   len(req.IDs),
	})
}

type CreateModuleVersionRequest struct {
	Content   string `json:"content" binding:"required"`
	Variables string `json:"variables"`
	Comment   string `json:"comment"`
}

func (h *ConfigHandler) CreateModuleVersion(c *gin.Context) {
	moduleID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req CreateModuleVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	version, err := h.Svc.CreateModuleVersion(uint(moduleID), c.GetUint("user_id"), req.Content, req.Variables, req.Comment)
	if err != nil {
		writeConfigError(c, err, "module")
		return
	}
	c.JSON(http.StatusCreated, version)
}

func (h *ConfigHandler) ListModuleVersions(c *gin.Context) {
	moduleID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	versions, err := h.Svc.ListModuleVersions(uint(moduleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": versions})
}

type RenderedConfigPreviewRequest struct {
	Name           string                     `json:"name"`
	FluentType     string                     `json:"fluent_type" binding:"required"`
	RuntimeVersion string                     `json:"runtime_version"`
	Modules        []services.RenderModuleRef `json:"modules" binding:"required"`
	Variables      string                     `json:"variables"`
}

func (h *ConfigHandler) PreviewRenderedConfig(c *gin.Context) {
	var req RenderedConfigPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rendered, err := h.Svc.PreviewRenderedConfig(&services.RenderedConfigPreviewInput{
		Name:           req.Name,
		FluentType:     req.FluentType,
		RuntimeVersion: req.RuntimeVersion,
		Modules:        req.Modules,
		Variables:      req.Variables,
	}, c.GetUint("user_id"))
	if err != nil {
		writeConfigError(c, err, "rendered config")
		return
	}
	c.JSON(http.StatusCreated, rendered)
}

func (h *ConfigHandler) GetRenderedConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	rendered, err := h.Svc.GetRenderedConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rendered config not found"})
		return
	}
	c.JSON(http.StatusOK, rendered)
}

func writeConfigError(c *gin.Context, err error, resource string) {
	switch {
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": resource + " not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
