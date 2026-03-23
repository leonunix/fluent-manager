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
	FindOrCreateExternalUser(username, email, displayName, authSource string, externalGroups []string) (*models.User, error)
	GetProfile(userID uint) (*models.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	UpdateLastLogin(userID uint)
}

type authService struct {
	db              *gorm.DB
	authSettingsSvc AuthSettingsService
}

func NewAuthService(db *gorm.DB) AuthService {
	return &authService{
		db:              db,
		authSettingsSvc: NewAuthSettingsService(db),
	}
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
	return s.FindOrCreateExternalUser(username, email, displayName, "ldap", nil)
}

func (s *authService) FindOrCreateExternalUser(username, email, displayName, authSource string, externalGroups []string) (*models.User, error) {
	var user models.User
	result := s.db.Where("username = ? AND auth_source = ?", username, authSource).First(&user)
	isNew := result.RowsAffected == 0

	if isNew {
		user = models.User{
			Username:    username,
			Email:       email,
			DisplayName: displayName,
			AuthSource:  authSource,
			IsActive:    true,
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}

		// Resolve external groups to system groups
		groups, _ := s.authSettingsSvc.ResolveExternalGroups(authSource, externalGroups)
		if len(groups) > 0 {
			s.db.Model(&user).Association("Groups").Replace(groups)
		} else {
			// Fallback: assign viewer role if no group mapping matched
			var viewerRole models.Role
			s.db.Where("name = ?", "viewer").First(&viewerRole)
			s.db.Model(&user).Association("Roles").Append(&viewerRole)
		}
	} else {
		// Existing user — check sync strategy
		syncStrategy := s.authSettingsSvc.GetGroupSyncStrategy(authSource)
		if syncStrategy == "always" && externalGroups != nil {
			// Resolve whatever the IdP returned (may be empty → revoke all groups)
			resolved, _ := s.authSettingsSvc.ResolveExternalGroups(authSource, externalGroups)
			// Replace always: empty slice clears all groups, non-empty sets the new set
			s.db.Model(&user).Association("Groups").Replace(resolved)
		}
		// Update profile fields
		s.db.Model(&user).Updates(map[string]interface{}{
			"email":        email,
			"display_name": displayName,
		})
	}

	return &user, nil
}

func (s *authService) GetProfile(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").Preload("Groups.Roles.Permissions").Preload("Groups.Scopes").Preload("Scopes").First(&user, userID).Error; err != nil {
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
