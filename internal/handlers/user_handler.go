package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct{}

func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	query := models.DB.Preload("Roles")

	if search := c.Query("search"); search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
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
	query.Model(&models.User{}).Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password" binding:"required,min=6"`
	RoleIDs     []uint `json:"role_ids"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if models.DB.Where("username = ?", req.Username).First(&existing).RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: string(hash),
		AuthSource:   "local",
		IsActive:     true,
	}
	models.DB.Create(&user)

	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		models.DB.Where("id IN ?", req.RoleIDs).Find(&roles)
		models.DB.Model(&user).Association("Roles").Replace(roles)
	}

	models.DB.Preload("Roles").First(&user, user.ID)
	c.JSON(http.StatusCreated, user)
}

type UpdateUserRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsActive    *bool  `json:"is_active"`
	Password    string `json:"password"`
	RoleIDs     []uint `json:"role_ids"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		models.DB.Model(&user).Updates(updates)
	}

	if req.RoleIDs != nil {
		var roles []models.Role
		models.DB.Where("id IN ?", req.RoleIDs).Find(&roles)
		models.DB.Model(&user).Association("Roles").Replace(roles)
	}

	models.DB.Preload("Roles").First(&user, user.ID)
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Username == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin user"})
		return
	}

	models.DB.Model(&user).Association("Roles").Clear()
	models.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var user models.User
	if err := models.DB.Preload("Roles.Permissions").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
