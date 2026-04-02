package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
)

func TestUpdateHostPreservesPrivateKeyWhenOmitted(t *testing.T) {
	db := testutil.NewTestDB()
	svc := NewBootstrapService(db, BootstrapSettings{Secret: "test-secret"}, nil)

	created, err := svc.CreateHost(BootstrapHostInput{
		Hostname:   "node-private",
		SSHUser:    "root",
		SSHPort:    22,
		AuthType:   "private_key",
		PrivateKey: "PRIVATE_KEY_DATA",
	}, 1, nil)
	if err != nil {
		t.Fatalf("CreateHost() error = %v", err)
	}

	var before models.BootstrapHost
	if err := db.First(&before, created.ID).Error; err != nil {
		t.Fatalf("load created host: %v", err)
	}

	updated, err := svc.UpdateHost(created.ID, BootstrapHostInput{
		Hostname:    "node-private",
		SSHUser:     "root",
		SSHPort:     22,
		AuthType:    "private_key",
		Description: "updated without resending key",
	}, nil)
	if err != nil {
		t.Fatalf("UpdateHost() error = %v", err)
	}
	if !updated.HasPrivateKey {
		t.Fatalf("expected HasPrivateKey to remain true")
	}
	if updated.Description != "updated without resending key" {
		t.Fatalf("unexpected description: %q", updated.Description)
	}

	var after models.BootstrapHost
	if err := db.First(&after, created.ID).Error; err != nil {
		t.Fatalf("load updated host: %v", err)
	}
	if after.PrivateKeyEncrypted == "" {
		t.Fatalf("expected private key ciphertext to remain present")
	}
	if after.PrivateKeyEncrypted != before.PrivateKeyEncrypted {
		t.Fatalf("expected private key ciphertext to be preserved when omitted")
	}
}

func TestUpdateHostPreservesPasswordWhenOmitted(t *testing.T) {
	installFakeExecutable(t, "sshpass")

	db := testutil.NewTestDB()
	svc := NewBootstrapService(db, BootstrapSettings{Secret: "test-secret"}, nil)

	created, err := svc.CreateHost(BootstrapHostInput{
		Hostname: "node-password",
		SSHUser:  "root",
		SSHPort:  22,
		AuthType: "password",
		Password: "super-secret",
	}, 1, nil)
	if err != nil {
		t.Fatalf("CreateHost() error = %v", err)
	}

	var before models.BootstrapHost
	if err := db.First(&before, created.ID).Error; err != nil {
		t.Fatalf("load created host: %v", err)
	}

	updated, err := svc.UpdateHost(created.ID, BootstrapHostInput{
		Hostname:    "node-password",
		SSHUser:     "root",
		SSHPort:     22,
		AuthType:    "password",
		Description: "updated without resending password",
	}, nil)
	if err != nil {
		t.Fatalf("UpdateHost() error = %v", err)
	}
	if !updated.HasPassword {
		t.Fatalf("expected HasPassword to remain true")
	}

	var after models.BootstrapHost
	if err := db.First(&after, created.ID).Error; err != nil {
		t.Fatalf("load updated host: %v", err)
	}
	if after.PasswordEncrypted == "" {
		t.Fatalf("expected password ciphertext to remain present")
	}
	if after.PasswordEncrypted != before.PasswordEncrypted {
		t.Fatalf("expected password ciphertext to be preserved when omitted")
	}
}

func TestUpdateHostSwitchingAuthClearsPreviousSecretFields(t *testing.T) {
	installFakeExecutable(t, "sshpass")

	db := testutil.NewTestDB()
	svc := NewBootstrapService(db, BootstrapSettings{Secret: "test-secret"}, nil)

	created, err := svc.CreateHost(BootstrapHostInput{
		Hostname: "node-switch",
		SSHUser:  "root",
		SSHPort:  22,
		AuthType: "password",
		Password: "super-secret",
	}, 1, nil)
	if err != nil {
		t.Fatalf("CreateHost() error = %v", err)
	}

	updated, err := svc.UpdateHost(created.ID, BootstrapHostInput{
		Hostname:   "node-switch",
		SSHUser:    "root",
		SSHPort:    22,
		AuthType:   "private_key",
		PrivateKey: "NEW_PRIVATE_KEY_DATA",
	}, nil)
	if err != nil {
		t.Fatalf("UpdateHost() error = %v", err)
	}
	if updated.HasPassword {
		t.Fatalf("expected HasPassword to be false after switching auth type")
	}
	if !updated.HasPrivateKey {
		t.Fatalf("expected HasPrivateKey to be true after switching auth type")
	}

	var after models.BootstrapHost
	if err := db.First(&after, created.ID).Error; err != nil {
		t.Fatalf("load updated host: %v", err)
	}
	if after.HasPassword {
		t.Fatalf("expected HasPassword to be false in database")
	}
	if after.PasswordEncrypted != "" {
		t.Fatalf("expected PasswordEncrypted to be cleared in database")
	}
	if !after.HasPrivateKey {
		t.Fatalf("expected HasPrivateKey to be true in database")
	}
	if after.PrivateKeyEncrypted == "" {
		t.Fatalf("expected PrivateKeyEncrypted to be set in database")
	}
}

func installFakeExecutable(t *testing.T, name string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	content := []byte("#!/bin/sh\nexit 0\n")
	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatalf("write fake executable %s: %v", name, err)
	}

	currentPath := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+currentPath)
}
