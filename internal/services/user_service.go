package services

import (
	"errors"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	List(search string, page, pageSize int) ([]models.User, int64, error)
	Get(id uint) (*models.User, error)
	Create(username, email, displayName, password string, roleIDs []uint) (*models.User, error)
	Update(id uint, email, displayName, password string, isActive *bool, roleIDs []uint) (*models.User, error)
	Delete(id uint) error
}

type userService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) UserService {
	return &userService{db: db}
}

func (s *userService) List(search string, page, pageSize int) ([]models.User, int64, error) {
	query := s.db.Preload("Roles")
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Model(&models.User{}).Count(&total)

	var users []models.User
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (s *userService) Get(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) Create(username, email, displayName, password string, roleIDs []uint) (*models.User, error) {
	var existing models.User
	if s.db.Where("username = ?", username).First(&existing).RowsAffected > 0 {
		return nil, errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Username:     username,
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: string(hash),
		AuthSource:   "local",
		IsActive:     true,
	}
	s.db.Create(&user)

	if len(roleIDs) > 0 {
		var roles []models.Role
		s.db.Where("id IN ?", roleIDs).Find(&roles)
		s.db.Model(&user).Association("Roles").Replace(roles)
	}

	s.db.Preload("Roles").First(&user, user.ID)
	return &user, nil
}

func (s *userService) Update(id uint, email, displayName, password string, isActive *bool, roleIDs []uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if email != "" {
		updates["email"] = email
	}
	if displayName != "" {
		updates["display_name"] = displayName
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		s.db.Model(&user).Updates(updates)
	}

	if roleIDs != nil {
		var roles []models.Role
		s.db.Where("id IN ?", roleIDs).Find(&roles)
		s.db.Model(&user).Association("Roles").Replace(roles)
	}

	s.db.Preload("Roles").First(&user, user.ID)
	return &user, nil
}

func (s *userService) Delete(id uint) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return err
	}
	s.db.Model(&user).Association("Roles").Clear()
	return s.db.Delete(&user).Error
}
