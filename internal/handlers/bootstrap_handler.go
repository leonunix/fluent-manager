package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BootstrapHandler struct {
	Svc services.BootstrapService
}

func (h *BootstrapHandler) Capability(c *gin.Context) {
	c.JSON(http.StatusOK, h.Svc.GetCapability())
}

func (h *BootstrapHandler) ListHosts(c *gin.Context) {
	hosts, err := h.Svc.ListHosts(middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": hosts})
}

func (h *BootstrapHandler) CreateHost(c *gin.Context) {
	var req services.BootstrapHostInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	host, err := h.Svc.CreateHost(req, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	middleware.SetAuditResource(c, "bootstrap_host", host.ID)
	c.JSON(http.StatusCreated, host)
}

func (h *BootstrapHandler) CreateHosts(c *gin.Context) {
	var req struct {
		Hosts []services.BootstrapHostInput `json:"hosts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hosts, err := h.Svc.CreateHosts(req.Hosts, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": hosts, "count": len(hosts)})
}

func (h *BootstrapHandler) UpdateHost(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req services.BootstrapHostInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	host, err := h.Svc.UpdateHost(id, req, middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	middleware.SetAuditResource(c, "bootstrap_host", host.ID)
	c.JSON(http.StatusOK, host)
}

func (h *BootstrapHandler) DeleteHost(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.Svc.DeleteHost(id, middleware.GetAllowedClusters(c)); err != nil {
		writeBootstrapError(c, err)
		return
	}
	middleware.SetAuditResource(c, "bootstrap_host", id)
	c.JSON(http.StatusOK, gin.H{"message": "bootstrap host deleted"})
}

func (h *BootstrapHandler) CreateTask(c *gin.Context) {
	var req services.BootstrapTaskInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.Svc.Create(req, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	middleware.SetAuditResource(c, "bootstrap_task", task.ID)
	c.JSON(http.StatusCreated, task)
}

func (h *BootstrapHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := h.Svc.List(page, pageSize, middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *BootstrapHandler) GetTask(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	task, records, err := h.Svc.Get(id, middleware.GetAllowedClusters(c))
	if err != nil {
		writeBootstrapError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"records": records,
	})
}

func writeBootstrapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "bootstrap resource not found"})
	default:
		log.Printf("WARNING: bootstrap handler internal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
