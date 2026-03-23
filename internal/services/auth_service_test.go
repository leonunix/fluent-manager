package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupAuthTest(t *testing.T) (*gorm.DB, AuthService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewAuthService(db)
	return db, svc
}

func createLocalUser(t *testing.T, db *gorm.DB, username, password string) *models.User {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	user := models.User{
		Username:     username,
		PasswordHash: string(hash),
		AuthSource:   "local",
		IsActive:     true,
	}
	db.Create(&user)
	return &user
}

func TestLocalLogin_Success(t *testing.T) {
	db, svc := setupAuthTest(t)
	createLocalUser(t, db, "admin", "secret123")

	user, err := svc.LocalLogin("admin", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "admin" {
		t.Error("wrong user returned")
	}
}

func TestLocalLogin_WrongPassword(t *testing.T) {
	db, svc := setupAuthTest(t)
	createLocalUser(t, db, "admin", "secret123")

	_, err := svc.LocalLogin("admin", "wrongpass")
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestLocalLogin_UserNotFound(t *testing.T) {
	_, svc := setupAuthTest(t)

	_, err := svc.LocalLogin("nonexistent", "pass")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestLocalLogin_LDAPUserNotMatched(t *testing.T) {
	db, svc := setupAuthTest(t)

	// Create LDAP user - should not be found by local login
	db.Create(&models.User{Username: "ldapuser", AuthSource: "ldap", IsActive: true})

	_, err := svc.LocalLogin("ldapuser", "pass")
	if err == nil {
		t.Error("LDAP user should not authenticate via local login")
	}
}

func TestFindOrCreateLDAPUser_Create(t *testing.T) {
	db, svc := setupAuthTest(t)

	// Seed viewer role
	db.Create(&models.Role{Name: "viewer", Description: "read-only"})

	user, err := svc.FindOrCreateLDAPUser("ldapuser", "ldap@test.com", "LDAP User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "ldapuser" || user.AuthSource != "ldap" {
		t.Error("LDAP user not created correctly")
	}
}

func TestFindOrCreateLDAPUser_FindExisting(t *testing.T) {
	db, svc := setupAuthTest(t)

	db.Create(&models.User{Username: "ldapuser", AuthSource: "ldap", Email: "old@test.com", IsActive: true})

	user, err := svc.FindOrCreateLDAPUser("ldapuser", "new@test.com", "Updated Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return existing, not create new
	if user.Email != "old@test.com" {
		t.Error("should return existing user without updating")
	}
}

func TestGetProfile(t *testing.T) {
	db, svc := setupAuthTest(t)
	u := createLocalUser(t, db, "admin", "pass")

	role := models.Role{Name: "admin"}
	db.Create(&role)
	db.Model(u).Association("Roles").Append(&role)

	profile, err := svc.GetProfile(u.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Username != "admin" {
		t.Error("wrong profile")
	}
	if len(profile.Roles) != 1 {
		t.Error("roles should be preloaded")
	}
}

func TestChangePassword_Success(t *testing.T) {
	db, svc := setupAuthTest(t)
	u := createLocalUser(t, db, "admin", "oldpass")

	err := svc.ChangePassword(u.ID, "oldpass", "newpass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify new password works
	_, err = svc.LocalLogin("admin", "newpass")
	if err != nil {
		t.Error("new password should work after change")
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	db, svc := setupAuthTest(t)
	u := createLocalUser(t, db, "admin", "oldpass")

	err := svc.ChangePassword(u.ID, "wrongold", "newpass")
	if err == nil {
		t.Error("expected error for wrong old password")
	}
}

func TestChangePassword_NonLocalUser(t *testing.T) {
	db, svc := setupAuthTest(t)

	user := models.User{Username: "ldapuser", AuthSource: "ldap", IsActive: true}
	db.Create(&user)

	err := svc.ChangePassword(user.ID, "", "newpass")
	if err == nil {
		t.Error("expected error for non-local user")
	}
}

func TestUpdateLastLogin(t *testing.T) {
	db, svc := setupAuthTest(t)
	u := createLocalUser(t, db, "admin", "pass")

	svc.UpdateLastLogin(u.ID)

	var updated models.User
	db.First(&updated, u.ID)
	if updated.LastLoginAt == nil {
		t.Error("last_login_at should be set")
	}
}
