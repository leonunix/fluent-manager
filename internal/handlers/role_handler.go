package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	Svc services.RoleService
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.Svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RoleHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	role, err := h.Svc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.Svc.Create(req.Name, req.Description, req.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role name already exists"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.Svc.Update(uint(id), req.Name, req.Description, req.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Check if it's admin role first
	role, err := h.Svc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.Name == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin role"})
		return
	}

	if err := h.Svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.Svc.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}
