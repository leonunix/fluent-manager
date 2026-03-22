package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct{}

func (h *RoleHandler) List(c *gin.Context) {
	var roles []models.Role
	models.DB.Preload("Permissions").Find(&roles)
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RoleHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var role models.Role
	if err := models.DB.Preload("Permissions").First(&role, id).Error; err != nil {
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

	role := models.Role{Name: req.Name, Description: req.Description}
	if err := models.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role name already exists"})
		return
	}

	if len(req.PermissionIDs) > 0 {
		var perms []models.Permission
		models.DB.Where("id IN ?", req.PermissionIDs).Find(&perms)
		models.DB.Model(&role).Association("Permissions").Replace(perms)
	}

	models.DB.Preload("Permissions").First(&role, role.ID)
	c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var role models.Role
	if err := models.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&role).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	})

	if req.PermissionIDs != nil {
		var perms []models.Permission
		models.DB.Where("id IN ?", req.PermissionIDs).Find(&perms)
		models.DB.Model(&role).Association("Permissions").Replace(perms)
	}

	models.DB.Preload("Permissions").First(&role, role.ID)
	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var role models.Role
	if err := models.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	if role.Name == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin role"})
		return
	}

	models.DB.Model(&role).Association("Permissions").Clear()
	models.DB.Model(&role).Association("Users").Clear()
	models.DB.Delete(&role)
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	var perms []models.Permission
	models.DB.Find(&perms)
	c.JSON(http.StatusOK, gin.H{"data": perms})
}
