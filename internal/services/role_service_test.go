package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupRoleTest(t *testing.T) (*gorm.DB, RoleService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewRoleService(db)
	return db, svc
}

func TestCreateRole(t *testing.T) {
	_, svc := setupRoleTest(t)

	role, err := svc.Create("editor", "Can edit things", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "editor" {
		t.Error("role not created correctly")
	}
}

func TestCreateRole_WithPermissions(t *testing.T) {
	db, svc := setupRoleTest(t)

	p1 := models.Permission{Name: "nodes:read", Resource: "nodes", Action: "read"}
	p2 := models.Permission{Name: "nodes:update", Resource: "nodes", Action: "update"}
	db.Create(&p1)
	db.Create(&p2)

	role, err := svc.Create("operator", "Node operator", []uint{p1.ID, p2.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(role.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(role.Permissions))
	}
}

func TestListRoles(t *testing.T) {
	_, svc := setupRoleTest(t)

	svc.Create("admin", "Full access", nil)
	svc.Create("viewer", "Read only", nil)

	roles, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestGetRole(t *testing.T) {
	_, svc := setupRoleTest(t)

	created, _ := svc.Create("admin", "", nil)
	role, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "admin" {
		t.Error("wrong role returned")
	}
}

func TestUpdateRole(t *testing.T) {
	db, svc := setupRoleTest(t)

	p1 := models.Permission{Name: "nodes:read", Resource: "nodes", Action: "read"}
	db.Create(&p1)

	role, _ := svc.Create("editor", "Old desc", nil)

	updated, err := svc.Update(role.ID, "super-editor", "New desc", []uint{p1.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "super-editor" {
		t.Error("role name not updated")
	}
	if len(updated.Permissions) != 1 {
		t.Error("permissions not updated")
	}
}

func TestDeleteRole(t *testing.T) {
	_, svc := setupRoleTest(t)

	role, _ := svc.Create("temp", "", nil)
	if err := svc.Delete(role.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(role.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestDeleteRole_ClearsAssociations(t *testing.T) {
	db, svc := setupRoleTest(t)

	p := models.Permission{Name: "nodes:read", Resource: "nodes", Action: "read"}
	db.Create(&p)

	role, _ := svc.Create("temp", "", []uint{p.ID})

	user := models.User{Username: "test", PasswordHash: "x", AuthSource: "local", IsActive: true}
	db.Create(&user)
	db.Model(&user).Association("Roles").Append(role)

	svc.Delete(role.ID)

	// Verify associations cleared
	var count int64
	db.Table("user_roles").Where("role_id = ?", role.ID).Count(&count)
	if count != 0 {
		t.Error("user_roles association should be cleared")
	}
}

func TestListPermissions(t *testing.T) {
	db, svc := setupRoleTest(t)

	db.Create(&models.Permission{Name: "nodes:read", Resource: "nodes", Action: "read"})
	db.Create(&models.Permission{Name: "nodes:create", Resource: "nodes", Action: "create"})

	perms, err := svc.ListPermissions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(perms))
	}
}
