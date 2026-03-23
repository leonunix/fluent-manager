package services

import (
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type RoleService interface {
	List() ([]models.Role, error)
	Get(id uint) (*models.Role, error)
	Create(name, description string, permissionIDs []uint) (*models.Role, error)
	Update(id uint, name, description string, permissionIDs []uint) (*models.Role, error)
	Delete(id uint) error
	ListPermissions() ([]models.Permission, error)
}

type roleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) RoleService {
	return &roleService{db: db}
}

func (s *roleService) List() ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (s *roleService) Get(id uint) (*models.Role, error) {
	var role models.Role
	if err := s.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *roleService) Create(name, description string, permissionIDs []uint) (*models.Role, error) {
	role := models.Role{Name: name, Description: description}
	if err := s.db.Create(&role).Error; err != nil {
		return nil, err
	}

	if len(permissionIDs) > 0 {
		var perms []models.Permission
		s.db.Where("id IN ?", permissionIDs).Find(&perms)
		s.db.Model(&role).Association("Permissions").Replace(perms)
	}

	s.db.Preload("Permissions").First(&role, role.ID)
	return &role, nil
}

func (s *roleService) Update(id uint, name, description string, permissionIDs []uint) (*models.Role, error) {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return nil, err
	}

	s.db.Model(&role).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	})

	if permissionIDs != nil {
		var perms []models.Permission
		s.db.Where("id IN ?", permissionIDs).Find(&perms)
		s.db.Model(&role).Association("Permissions").Replace(perms)
	}

	s.db.Preload("Permissions").First(&role, role.ID)
	return &role, nil
}

func (s *roleService) Delete(id uint) error {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return err
	}
	s.db.Model(&role).Association("Permissions").Clear()
	s.db.Model(&role).Association("Users").Clear()
	return s.db.Delete(&role).Error
}

func (s *roleService) ListPermissions() ([]models.Permission, error) {
	var perms []models.Permission
	err := s.db.Find(&perms).Error
	return perms, err
}
