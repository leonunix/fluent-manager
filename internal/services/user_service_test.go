package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupUserTest(t *testing.T) (*gorm.DB, UserService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewUserService(db)
	return db, svc
}

func TestCreateUser(t *testing.T) {
	_, svc := setupUserTest(t)

	user, err := svc.Create("testuser", "test@example.com", "Test User", "password123", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" || user.Email != "test@example.com" {
		t.Error("user not created correctly")
	}
	if user.PasswordHash == "password123" {
		t.Error("password should be hashed, not stored in plaintext")
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	_, svc := setupUserTest(t)

	svc.Create("testuser", "", "", "pass", nil, nil)
	_, err := svc.Create("testuser", "", "", "pass", nil, nil)
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}

func TestCreateUser_WithRoles(t *testing.T) {
	db, svc := setupUserTest(t)

	role := models.Role{Name: "viewer", Description: "read only"}
	db.Create(&role)

	user, err := svc.Create("testuser", "", "", "pass", []uint{role.ID}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(user.Roles) != 1 || user.Roles[0].Name != "viewer" {
		t.Error("role not assigned")
	}
}

func TestListUsers(t *testing.T) {
	_, svc := setupUserTest(t)

	svc.Create("alice", "alice@test.com", "Alice", "pass", nil, nil)
	svc.Create("bob", "bob@test.com", "Bob", "pass", nil, nil)

	users, total, _ := svc.List("", 1, 10)
	if total != 2 {
		t.Errorf("expected 2 users, got %d", total)
	}

	users, total, _ = svc.List("alice", 1, 10)
	if total != 1 || users[0].Username != "alice" {
		t.Error("search filter not working")
	}
}

func TestGetUser(t *testing.T) {
	_, svc := setupUserTest(t)

	created, _ := svc.Create("testuser", "", "", "pass", nil, nil)
	user, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" {
		t.Error("wrong user returned")
	}
}

func TestUpdateUser(t *testing.T) {
	_, svc := setupUserTest(t)

	user, _ := svc.Create("testuser", "old@test.com", "Old Name", "pass", nil, nil)

	isActive := false
	updated, err := svc.Update(user.ID, "new@test.com", "New Name", "", &isActive, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Email != "new@test.com" || updated.DisplayName != "New Name" {
		t.Error("user not updated")
	}
	if updated.IsActive {
		t.Error("is_active should be false")
	}
}

func TestUpdateUser_ChangePassword(t *testing.T) {
	db, svc := setupUserTest(t)

	user, _ := svc.Create("testuser", "", "", "oldpass", nil, nil)

	var before models.User
	db.First(&before, user.ID)
	oldHash := before.PasswordHash

	svc.Update(user.ID, "", "", "newpass", nil, nil, nil)

	var after models.User
	db.First(&after, user.ID)
	if after.PasswordHash == oldHash {
		t.Error("password hash should have changed")
	}
}

func TestUpdateUser_ChangeRoles(t *testing.T) {
	db, svc := setupUserTest(t)

	r1 := models.Role{Name: "admin"}
	r2 := models.Role{Name: "viewer"}
	db.Create(&r1)
	db.Create(&r2)

	user, _ := svc.Create("testuser", "", "", "pass", []uint{r1.ID}, nil)

	updated, _ := svc.Update(user.ID, "", "", "", nil, []uint{r2.ID}, nil)
	if len(updated.Roles) != 1 || updated.Roles[0].Name != "viewer" {
		t.Error("roles not updated correctly")
	}
}

func TestDeleteUser(t *testing.T) {
	_, svc := setupUserTest(t)

	user, _ := svc.Create("testuser", "", "", "pass", nil, nil)
	if err := svc.Delete(user.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(user.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}
