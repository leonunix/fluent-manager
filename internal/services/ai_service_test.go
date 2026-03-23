package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/testutil"
)

func seedAISettingsForTest(t *testing.T, provider, baseURL string) AuthSettingsService {
	t.Helper()

	db := testutil.NewTestDB()
	settingsSvc := NewAuthSettingsService(db)
	if err := settingsSvc.UpdateAISettings(AISettingsDTO{
		Enabled:               true,
		ActiveAccountID:       provider + "-primary",
		RequestTimeoutSeconds: 15,
		Accounts: []AIAccountDTO{
			{
				ID:       provider + "-primary",
				Name:     strings.ToUpper(provider[:1]) + provider[1:] + " Primary",
				Provider: provider,
				Enabled:  true,
				APIKey:   "secret",
				BaseURL:  baseURL,
				Model:    "test-model",
			},
		},
	}); err != nil {
		t.Fatalf("failed to seed ai settings: %v", err)
	}
	return settingsSvc
}

func TestAnalyzeLogSampleWithOpenAICompatibleProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"detected_format\":\"json line\",\"summary\":\"tail input with parser\",\"recommended_module_name\":\"nginx_input\",\"recommended_template_name\":\"nginx_pipeline\",\"module_type\":\"input\",\"variables_json\":\"{\\\"path\\\":\\\"/var/log/nginx/access.log\\\"}\",\"module_content\":\"[INPUT]\\n    Name tail\",\"template_content\":\"# template\",\"assembly_steps\":[\"Create input module\"],\"notes\":[\"Review parser fields\"]}"}}]}`)
	}))
	defer server.Close()

	settingsSvc := seedAISettingsForTest(t, "openai", server.URL)
	svc := &aiService{settingsSvc: settingsSvc, httpClient: server.Client()}

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

func TestAnalyzeLogSampleWithClaudeProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "secret" {
			t.Fatalf("unexpected api key header: %s", got)
		}
		if got := r.Header.Get("anthropic-version"); got == "" {
			t.Fatal("expected anthropic-version header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"msg_1","type":"message","role":"assistant","model":"test-model","content":[{"type":"text","text":"{\"detected_format\":\"plain text\",\"summary\":\"use parser module\",\"recommended_module_name\":\"app_parser\",\"recommended_template_name\":\"app_pipeline\",\"module_type\":\"parser\",\"variables_json\":\"{\\\"format\\\":\\\"regex\\\"}\",\"module_content\":\"<parse>\\n  @type regexp\\n</parse>\",\"template_content\":\"# template\",\"assembly_steps\":[\"Create parser module\"],\"notes\":[\"Verify regex fields\"]}"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()

	settingsSvc := seedAISettingsForTest(t, "claude", server.URL)
	svc := &aiService{settingsSvc: settingsSvc, httpClient: server.Client()}

	result, err := svc.AnalyzeLogSample(&LogSampleAnalysisInput{
		FluentType: "fluentd",
		Goal:       "module",
		ModuleType: "parser",
		Sample:     `2026-03-23T10:00:00Z INFO request completed`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provider != "claude" {
		t.Fatalf("expected provider claude, got %q", result.Provider)
	}
	if result.ModuleType != "parser" {
		t.Fatalf("expected parser module type, got %q", result.ModuleType)
	}
	if !strings.Contains(result.ModuleContent, "@type regexp") {
		t.Fatalf("expected parser content, got %q", result.ModuleContent)
	}
}

func TestAnalyzeLogSampleWithGeminiProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":generateContent") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "secret" {
			t.Fatalf("unexpected api key header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"text":"{\"detected_format\":\"json\",\"summary\":\"json logs can be tailed directly\",\"recommended_module_name\":\"json_input\",\"recommended_template_name\":\"json_pipeline\",\"module_type\":\"input\",\"variables_json\":\"{\\\"path\\\":\\\"/var/log/app.json\\\"}\",\"module_content\":\"[INPUT]\\n    Name tail\",\"template_content\":\"[INPUT]\\n    Name tail\",\"assembly_steps\":[\"Create tail input\"],\"notes\":[\"Confirm file path\"]}"}],"role":"model"}}]}`)
	}))
	defer server.Close()

	settingsSvc := seedAISettingsForTest(t, "gemini", server.URL)
	svc := &aiService{settingsSvc: settingsSvc, httpClient: server.Client()}

	result, err := svc.AnalyzeLogSample(&LogSampleAnalysisInput{
		FluentType: "fluentbit",
		Goal:       "both",
		ModuleType: "input",
		Sample:     `{"message":"hello","level":"info"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provider != "gemini" {
		t.Fatalf("expected provider gemini, got %q", result.Provider)
	}
	if result.RecommendedTemplateName != "json_pipeline" {
		t.Fatalf("expected template recommendation, got %q", result.RecommendedTemplateName)
	}
	if !strings.Contains(result.VariablesJSON, "/var/log/app.json") {
		t.Fatalf("expected normalized variables json, got %q", result.VariablesJSON)
	}
}
