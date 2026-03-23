package services

import (
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/testutil"
)

func setupAuthSettingsTest(t *testing.T) AuthSettingsService {
	t.Helper()
	db := testutil.NewTestDB()
	return NewAuthSettingsService(db)
}

func TestAISettingsDefaults(t *testing.T) {
	svc := setupAuthSettingsTest(t)

	settings, err := svc.GetAISettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings.ActiveProvider != "" {
		t.Fatalf("expected no default provider without accounts, got %q", settings.ActiveProvider)
	}
	if settings.RequestTimeoutSeconds != 60 {
		t.Fatalf("expected default timeout 60, got %d", settings.RequestTimeoutSeconds)
	}
}

func TestUpdateAISettingsRoundTrip(t *testing.T) {
	svc := setupAuthSettingsTest(t)

	err := svc.UpdateAISettings(AISettingsDTO{
		Enabled:               true,
		ActiveAccountID:       "deepseek-account",
		RequestTimeoutSeconds: 45,
		SystemPrompt:          "Summarize the log sample into modules",
		Accounts: []AIAccountDTO{
			{
				ID:       "openai-account",
				Name:     "OpenAI Primary",
				Provider: "openai",
				Enabled:  true,
				APIKey:   "sk-openai",
				BaseURL:  "https://api.openai.com/v1",
				Model:    "gpt-ops",
			},
			{
				ID:       "deepseek-account",
				Name:     "DeepSeek Backup",
				Provider: "deepseek",
				Enabled:  true,
				APIKey:   "sk-deepseek",
				BaseURL:  "https://api.deepseek.com/v1",
				Model:    "deepseek-chat",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings, err := svc.GetAISettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !settings.Enabled {
		t.Fatal("expected ai settings to be enabled")
	}
	if settings.ActiveAccountID != "deepseek-account" {
		t.Fatalf("expected active account deepseek-account, got %q", settings.ActiveAccountID)
	}
	if settings.ActiveProvider != "deepseek" {
		t.Fatalf("expected active provider deepseek, got %q", settings.ActiveProvider)
	}
	if len(settings.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(settings.Accounts))
	}
	if settings.Accounts[0].APIKey != "sk-openai" {
		t.Fatalf("expected first account key to persist, got %q", settings.Accounts[0].APIKey)
	}
	if settings.Accounts[1].Model != "deepseek-chat" {
		t.Fatalf("expected deepseek model to persist, got %q", settings.Accounts[1].Model)
	}
	if settings.SystemPrompt == "" {
		t.Fatal("expected system prompt to persist")
	}
}

func TestUpdateAISettingsNormalizesInvalidActiveAccount(t *testing.T) {
	svc := setupAuthSettingsTest(t)

	if err := svc.UpdateAISettings(AISettingsDTO{
		ActiveAccountID:       "missing-account",
		RequestTimeoutSeconds: 0,
		Accounts: []AIAccountDTO{
			{
				ID:       "gemini-one",
				Name:     "Gemini Team A",
				Provider: "gemini",
				Enabled:  true,
			},
		},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings, err := svc.GetAISettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings.ActiveAccountID != "gemini-one" {
		t.Fatalf("expected invalid active account to normalize to gemini-one, got %q", settings.ActiveAccountID)
	}
	if settings.ActiveProvider != "gemini" {
		t.Fatalf("expected active provider to normalize to gemini, got %q", settings.ActiveProvider)
	}
	if settings.RequestTimeoutSeconds != 60 {
		t.Fatalf("expected timeout to normalize to 60, got %d", settings.RequestTimeoutSeconds)
	}
}
