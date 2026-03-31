package services

import (
	"errors"
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

// newModulesResponse builds a minimal valid AI JSON response with the new modules+pipelines schema.
func newModulesResponse(modules, pipelines string) string {
	return fmt.Sprintf(
		`{"detected_format":"json line","summary":"tail input","modules":[%s],"pipelines":[%s],"assembly_steps":["step1"],"notes":["note1"]}`,
		modules, pipelines,
	)
}

func TestAnalyzeLogSampleWithOpenAICompatibleProvider(t *testing.T) {
	moduleJSON := `{"name":"nginx-input","module_type":"input","variables_json":"{\"path\":\"/var/log/nginx/access.log\"}","content":"[INPUT]\n    Name tail","note":""}`
	pipelineJSON := `{"name":"nginx-pipeline","description":"nginx access log pipeline","module_names":["nginx-input"],"template_content":"[INPUT]\n    Name tail","note":""}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		body := fmt.Sprintf(`{"choices":[{"message":{"content":%q}}]}`, newModulesResponse(moduleJSON, pipelineJSON))
		fmt.Fprint(w, body)
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
	if len(result.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result.Modules))
	}
	if result.Modules[0].Name != "nginx-input" {
		t.Fatalf("expected module name nginx-input, got %q", result.Modules[0].Name)
	}
	if result.Modules[0].ModuleType != "input" {
		t.Fatalf("expected module type input, got %q", result.Modules[0].ModuleType)
	}
	if result.Modules[0].VariablesJSON == "" {
		t.Fatal("expected variables json to be populated")
	}
	if len(result.Pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(result.Pipelines))
	}
	if result.Pipelines[0].Name != "nginx-pipeline" {
		t.Fatalf("expected pipeline name nginx-pipeline, got %q", result.Pipelines[0].Name)
	}
}

func TestAnalyzeLogSampleWithClaudeProvider(t *testing.T) {
	moduleJSON := `{"name":"app-parser","module_type":"parser","variables_json":"{\"format\":\"regex\"}","content":"<parse>\n  @type regexp\n</parse>","note":""}`
	pipelineJSON := `{"name":"app-pipeline","description":"app log pipeline","module_names":["app-parser"],"template_content":"<source>\n  @type tail\n</source>","note":""}`

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
		body := fmt.Sprintf(
			`{"id":"msg_1","type":"message","role":"assistant","model":"test-model","content":[{"type":"text","text":%q}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`,
			newModulesResponse(moduleJSON, pipelineJSON),
		)
		fmt.Fprint(w, body)
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
	if len(result.Modules) != 1 || result.Modules[0].ModuleType != "parser" {
		t.Fatalf("expected parser module, got %+v", result.Modules)
	}
	if !strings.Contains(result.Modules[0].Content, "@type regexp") {
		t.Fatalf("expected parser content, got %q", result.Modules[0].Content)
	}
}

func TestAnalyzeLogSampleWithGeminiProvider(t *testing.T) {
	moduleJSON := `{"name":"json-input","module_type":"input","variables_json":"{\"path\":\"/var/log/app.json\"}","content":"[INPUT]\n    Name tail","note":""}`
	pipelineJSON := `{"name":"json-pipeline","description":"json log pipeline","module_names":["json-input"],"template_content":"[INPUT]\n    Name tail","note":""}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":generateContent") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "secret" {
			t.Fatalf("unexpected api key header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		body := fmt.Sprintf(
			`{"candidates":[{"content":{"parts":[{"text":%q}],"role":"model"}}]}`,
			newModulesResponse(moduleJSON, pipelineJSON),
		)
		fmt.Fprint(w, body)
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
	if len(result.Pipelines) != 1 || result.Pipelines[0].Name != "json-pipeline" {
		t.Fatalf("expected json-pipeline, got %+v", result.Pipelines)
	}
	if !strings.Contains(result.Modules[0].VariablesJSON, "/var/log/app.json") {
		t.Fatalf("expected normalized variables json, got %q", result.Modules[0].VariablesJSON)
	}
}

func TestTestAccountWithOpenAICompatibleProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"content":"PONG"}}]}`)
	}))
	defer server.Close()

	svc := &aiService{httpClient: server.Client()}
	result, err := svc.TestAccount(&AITestAccountInput{
		ID:                    "openai-primary",
		Name:                  "OpenAI Primary",
		Provider:              "openai",
		APIKey:                "secret",
		BaseURL:               server.URL,
		Model:                 "test-model",
		RequestTimeoutSeconds: 15,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success result")
	}
	if result.Provider != "openai" {
		t.Fatalf("expected provider openai, got %q", result.Provider)
	}
	if result.Response != "PONG" {
		t.Fatalf("expected response PONG, got %q", result.Response)
	}
}

func TestTestAccountReturnsStructuredProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"message":"invalid api key","type":"invalid_request_error","code":"invalid_api_key"}}`)
	}))
	defer server.Close()

	svc := &aiService{httpClient: server.Client()}
	_, err := svc.TestAccount(&AITestAccountInput{
		Provider:              "openai",
		APIKey:                "secret",
		BaseURL:               server.URL,
		Model:                 "test-model",
		RequestTimeoutSeconds: 15,
	})
	if err == nil {
		t.Fatal("expected provider error")
	}

	var providerErr *AIProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected AIProviderError, got %T", err)
	}
	if providerErr.HTTPStatus() != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", providerErr.HTTPStatus())
	}
	if providerErr.Code != "invalid_api_key" {
		t.Fatalf("expected code invalid_api_key, got %q", providerErr.Code)
	}
	if !strings.Contains(providerErr.ProviderMessage, "invalid api key") {
		t.Fatalf("expected provider message to mention invalid api key, got %q", providerErr.ProviderMessage)
	}
}
