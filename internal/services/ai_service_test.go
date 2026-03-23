package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/testutil"
)

func TestAnalyzeLogSampleWithOpenAICompatibleProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"detected_format\":\"json line\",\"summary\":\"tail input with parser\",\"recommended_module_name\":\"nginx_input\",\"recommended_template_name\":\"nginx_pipeline\",\"module_type\":\"input\",\"variables_json\":\"{\\\"path\\\":\\\"/var/log/nginx/access.log\\\"}\",\"module_content\":\"[INPUT]\\n    Name tail\",\"template_content\":\"# template\",\"assembly_steps\":[\"Create input module\"],\"notes\":[\"Review parser fields\"]}"}}]}`)
	}))
	defer server.Close()

	db := testutil.NewTestDB()
	settingsSvc := NewAuthSettingsService(db)
	if err := settingsSvc.UpdateAISettings(AISettingsDTO{
		Enabled:               true,
		ActiveAccountID:       "openai-primary",
		RequestTimeoutSeconds: 60,
		Accounts: []AIAccountDTO{
			{
				ID:       "openai-primary",
				Name:     "OpenAI Primary",
				Provider: "openai",
				Enabled:  true,
				APIKey:   "secret",
				BaseURL:  server.URL,
				Model:    "gpt-test",
			},
		},
	}); err != nil {
		t.Fatalf("failed to seed ai settings: %v", err)
	}

	svc := NewAIService(settingsSvc)
	result, err := svc.AnalyzeLogSample(&LogSampleAnalysisInput{
		FluentType: "fluentbit",
		Goal:       "both",
		ModuleType: "input",
		Sample:     `{"level":"info","msg":"hello"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provider != "openai" {
		t.Fatalf("expected provider openai, got %q", result.Provider)
	}
	if result.AccountID != "openai-primary" {
		t.Fatalf("expected account openai-primary, got %q", result.AccountID)
	}
	if result.RecommendedModuleName != "nginx_input" {
		t.Fatalf("expected recommended module name, got %q", result.RecommendedModuleName)
	}
	if result.ModuleType != "input" {
		t.Fatalf("expected module type input, got %q", result.ModuleType)
	}
	if result.VariablesJSON == "" {
		t.Fatal("expected variables json to be populated")
	}
}
