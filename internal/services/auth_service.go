package services

import (
	"errors"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	LocalLogin(username, password string) (*models.User, error)
	FindOrCreateLDAPUser(username, email, displayName string) (*models.User, error)
	GetProfile(userID uint) (*models.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	UpdateLastLogin(userID uint)
}

type authService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) AuthService {
	return &authService{db: db}
}

func (s *authService) LocalLogin(username, password string) (*models.User, error) {
	var user models.User
	result := s.db.Where("username = ? AND auth_source = ?", username, "local").First(&user)
	if result.RowsAffected == 0 {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}

func (s *authService) FindOrCreateLDAPUser(username, email, displayName string) (*models.User, error) {
	var user models.User
	result := s.db.Where("username = ? AND auth_source = ?", username, "ldap").First(&user)
	if result.RowsAffected == 0 {
		user = models.User{
			Username:    username,
			Email:       email,
			DisplayName: displayName,
			AuthSource:  "ldap",
			IsActive:    true,
		}
		s.db.Create(&user)
		var viewerRole models.Role
		s.db.Where("name = ?", "viewer").First(&viewerRole)
		s.db.Model(&user).Association("Roles").Append(&viewerRole)
	}
	return &user, nil
}

func (s *authService) GetProfile(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *authService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	if user.AuthSource != "local" {
		return errors.New("password change only available for local accounts")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Update("password_hash", string(hash)).Error
}

func (s *authService) UpdateLastLogin(userID uint) {
	now := time.Now()
	s.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", &now)
}
