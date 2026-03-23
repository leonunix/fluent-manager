package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LogSampleAnalysisInput struct {
	AccountID         string `json:"account_id"`
	FluentType        string `json:"fluent_type"`
	Goal              string `json:"goal"`
	ModuleType        string `json:"module_type"`
	Sample            string `json:"sample"`
	ExtraRequirements string `json:"extra_requirements"`
}

type LogSampleAnalysisResult struct {
	Provider                string   `json:"provider"`
	AccountID               string   `json:"account_id"`
	AccountName             string   `json:"account_name"`
	DetectedFormat          string   `json:"detected_format"`
	Summary                 string   `json:"summary"`
	RecommendedModuleName   string   `json:"recommended_module_name"`
	RecommendedTemplateName string   `json:"recommended_template_name"`
	ModuleType              string   `json:"module_type"`
	VariablesJSON           string   `json:"variables_json"`
	ModuleContent           string   `json:"module_content"`
	TemplateContent         string   `json:"template_content"`
	AssemblySteps           []string `json:"assembly_steps"`
	Notes                   []string `json:"notes"`
}

type AIService interface {
	AnalyzeLogSample(input *LogSampleAnalysisInput) (*LogSampleAnalysisResult, error)
}

type aiService struct {
	settingsSvc AuthSettingsService
	httpClient  *http.Client
}

func recommendedAISystemPrompt() string {
	return strings.TrimSpace(`
You are the built-in AI configuration architect for Fluent Manager.

Your job is to help enterprise users convert raw log samples into reusable Fluent Bit / Fluentd assets that are safe, understandable, and directly operable in production.

Follow these rules strictly:
1. Always return clean JSON only. Do not wrap the response in markdown fences.
2. Prefer practical, minimal, production-usable suggestions over theoretical explanations.
3. When generating Fluent config snippets, use Go template placeholders for variable fields, for example {{ .path }}, {{ .tag }}, {{ .match }}, {{ .host }}.
4. Variables JSON must be concise, realistic, and directly usable by the module editor.
5. If the log sample looks like plain text, suggest parser-related structure when useful.
6. If the log sample already looks like JSON, avoid over-designing parsers unless necessary.
7. Prefer naming that is readable in enterprise environments:
   - module names should be short, stable, and descriptive
   - template names should reflect the business purpose or log source
8. Assembly steps must be written for operators, not developers. Each step should be short and actionable.
9. Notes must highlight operational risk, assumptions, or fields that should be verified before deployment.
10. Do not invent unsupported plugins unless the sample clearly requires them.
11. For Fluent Bit, prefer modern, common plugin patterns. For Fluentd, keep structure clear and conventional.
12. Keep outputs focused on helping users create:
   - a reusable module
   - an understandable template
   - a variable set that can be edited later in the GUI

Default enterprise writing style:
- be concise
- be implementation-oriented
- avoid jargon when simpler wording works
- make results understandable to platform operators
`)
}

func NewAIService(settingsSvc AuthSettingsService) AIService {
	return &aiService{
		settingsSvc: settingsSvc,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (s *aiService) AnalyzeLogSample(input *LogSampleAnalysisInput) (*LogSampleAnalysisResult, error) {
	if strings.TrimSpace(input.Sample) == "" {
		return nil, fmt.Errorf("%w: log sample is required", ErrInvalidArgument)
	}
	if !validRenderedConfigTypes[input.FluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent type", ErrInvalidArgument)
	}
	if input.ModuleType != "" && !validConfigModuleTypes[input.ModuleType] {
		return nil, fmt.Errorf("%w: unsupported module type", ErrInvalidArgument)
	}

	settings, err := s.settingsSvc.GetAISettings()
	if err != nil {
		return nil, err
	}
	if settings == nil || !settings.Enabled {
		return nil, fmt.Errorf("%w: ai settings are not enabled", ErrForbidden)
	}

	account, err := selectAIAccount(settings, input.AccountID)
	if err != nil {
		return nil, err
	}
	if !account.Enabled {
		return nil, fmt.Errorf("%w: selected ai account is disabled", ErrForbidden)
	}
	if strings.TrimSpace(account.APIKey) == "" {
		return nil, fmt.Errorf("%w: api key is required for the selected ai account", ErrInvalidArgument)
	}
	if strings.TrimSpace(account.Model) == "" {
		return nil, fmt.Errorf("%w: model is required for the selected ai account", ErrInvalidArgument)
	}

	prompt := buildLogSamplePrompt(input)
	systemPrompt := buildSystemPrompt(settings.SystemPrompt)

	raw, err := s.generate(account, systemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	result, err := parseLogSampleAnalysis(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ai response: %w", err)
	}

	if result.ModuleType == "" {
		result.ModuleType = input.ModuleType
	}
	if result.ModuleType != "" && !validConfigModuleTypes[result.ModuleType] {
		result.ModuleType = input.ModuleType
	}
	if result.VariablesJSON != "" {
		result.VariablesJSON = prettyJSONString(result.VariablesJSON)
	}
	result.Provider = account.Provider
	result.AccountID = account.ID
	result.AccountName = account.Name
	return result, nil
}

func selectAIAccount(settings *AISettingsDTO, requestedID string) (*AIAccountDTO, error) {
	accountID := strings.TrimSpace(requestedID)
	if accountID == "" {
		accountID = settings.ActiveAccountID
	}
	for _, account := range settings.Accounts {
		if account.ID == accountID {
			copy := account
			return &copy, nil
		}
	}
	if len(settings.Accounts) == 0 {
		return nil, fmt.Errorf("%w: no ai accounts configured", ErrForbidden)
	}
	return nil, fmt.Errorf("%w: ai account not found", ErrForbidden)
}

func buildSystemPrompt(custom string) string {
	base := recommendedAISystemPrompt()
	custom = strings.TrimSpace(custom)
	if custom == "" {
		return base
	}
	return base + "\nAdditional enterprise rules:\n" + custom
}

func buildLogSamplePrompt(input *LogSampleAnalysisInput) string {
	goal := input.Goal
	if goal == "" {
		goal = "both"
	}
	moduleTypeHint := input.ModuleType
	if moduleTypeHint == "" {
		moduleTypeHint = "choose the most suitable one from service/input/parser/filter/route/output"
	}
	return fmt.Sprintf(`Analyze the following log sample and propose Fluent configuration assets.

Target runtime: %s
Desired output goal: %s
Preferred module type: %s
Additional requirements:
%s

Return JSON with exactly these fields:
{
  "detected_format": "short explanation",
  "summary": "short explanation",
  "recommended_module_name": "string",
  "recommended_template_name": "string",
  "module_type": "service|input|parser|filter|route|output",
  "variables_json": "{ ... pretty JSON string ... }",
  "module_content": "config snippet string",
  "template_content": "config template string",
  "assembly_steps": ["step 1", "step 2"],
  "notes": ["note 1", "note 2"]
}

Log sample:
%s
`, input.FluentType, goal, moduleTypeHint, strings.TrimSpace(input.ExtraRequirements), strings.TrimSpace(input.Sample))
}

func parseLogSampleAnalysis(raw string) (*LogSampleAnalysisResult, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	if !strings.HasPrefix(clean, "{") {
		start := strings.Index(clean, "{")
		end := strings.LastIndex(clean, "}")
		if start >= 0 && end > start {
			clean = clean[start : end+1]
		}
	}

	var result LogSampleAnalysisResult
	if err := json.Unmarshal([]byte(clean), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func prettyJSONString(value string) string {
	var payload interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return value
	}
	formatted, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return value
	}
	return string(formatted)
}

func defaultAIBaseURL(provider string) string {
	switch provider {
	case "openai":
		return "https://api.openai.com/v1"
	case "claude":
		return "https://api.anthropic.com"
	case "gemini":
		return "https://generativelanguage.googleapis.com"
	case "deepseek":
		return "https://api.deepseek.com/v1"
	default:
		return ""
	}
}

func (s *aiService) generate(account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	switch account.Provider {
	case "openai", "deepseek":
		return s.generateOpenAICompatible(account, systemPrompt, prompt)
	case "claude":
		return s.generateClaude(account, systemPrompt, prompt)
	case "gemini":
		return s.generateGemini(account, systemPrompt, prompt)
	default:
		return "", fmt.Errorf("%w: unsupported ai provider", ErrInvalidArgument)
	}
}

func (s *aiService) generateOpenAICompatible(account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(account.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultAIBaseURL(account.Provider)
	}
	reqBody := map[string]interface{}{
		"model": account.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+account.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("provider returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func (s *aiService) generateClaude(account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(account.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultAIBaseURL(account.Provider)
	}
	reqBody := map[string]interface{}{
		"model":       account.Model,
		"max_tokens":  2500,
		"system":      systemPrompt,
		"temperature": 0.2,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", account.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Content) == 0 {
		return "", fmt.Errorf("provider returned no content")
	}
	return parsed.Content[0].Text, nil
}

func (s *aiService) generateGemini(account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(account.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultAIBaseURL(account.Provider)
	}
	reqBody := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": systemPrompt},
			},
		},
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]float64{
			"temperature": 0.2,
		},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, account.Model, account.APIKey)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("provider returned no candidates")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
