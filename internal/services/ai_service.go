package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/openai/openai-go/v3"
	openaioption "github.com/openai/openai-go/v3/option"
	"google.golang.org/genai"
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

type AITestAccountInput struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Provider              string `json:"provider"`
	APIKey                string `json:"api_key"`
	BaseURL               string `json:"base_url"`
	Model                 string `json:"model"`
	RequestTimeoutSeconds int    `json:"request_timeout_seconds"`
}

type AITestAccountResult struct {
	Success     bool   `json:"success"`
	Provider    string `json:"provider"`
	AccountID   string `json:"account_id,omitempty"`
	AccountName string `json:"account_name,omitempty"`
	Model       string `json:"model"`
	Message     string `json:"message"`
	Response    string `json:"response"`
	LatencyMs   int64  `json:"latency_ms"`
}

type AIProviderError struct {
	Provider        string
	Code            string
	UserMessage     string
	ProviderMessage string
	StatusCode      int
	Cause           error
}

func (e *AIProviderError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.UserMessage) != "" {
		return e.UserMessage
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "ai provider request failed"
}

func (e *AIProviderError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *AIProviderError) HTTPStatus() int {
	if e == nil || e.StatusCode == 0 {
		return http.StatusBadGateway
	}
	return e.StatusCode
}

type AIService interface {
	AnalyzeLogSample(input *LogSampleAnalysisInput) (*LogSampleAnalysisResult, error)
	TestAccount(input *AITestAccountInput) (*AITestAccountResult, error)
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

func connectivityTestSystemPrompt() string {
	return "You are a connectivity test assistant for Fluent Manager. Reply with a very short plain-text acknowledgement."
}

func connectivityTestPrompt() string {
	return "Connectivity check from Fluent Manager. Reply with exactly: PONG"
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
	if err := validateAIAccount(account); err != nil {
		return nil, err
	}

	prompt := buildLogSamplePrompt(input)
	systemPrompt := buildSystemPrompt(settings.SystemPrompt)

	raw, err := s.generate(account, settings.RequestTimeoutSeconds, systemPrompt, prompt)
	if err != nil {
		return nil, wrapAIProviderError(account.Provider, err)
	}

	result, err := parseLogSampleAnalysis(raw)
	if err != nil {
		return nil, &AIProviderError{
			Provider:        account.Provider,
			Code:            "invalid_response",
			UserMessage:     "The model returned a response, but it was not in the JSON format required by Fluent Manager.",
			ProviderMessage: err.Error(),
			StatusCode:      http.StatusBadGateway,
			Cause:           err,
		}
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

func (s *aiService) TestAccount(input *AITestAccountInput) (*AITestAccountResult, error) {
	account := &AIAccountDTO{
		ID:       strings.TrimSpace(input.ID),
		Name:     strings.TrimSpace(input.Name),
		Provider: strings.TrimSpace(input.Provider),
		APIKey:   strings.TrimSpace(input.APIKey),
		BaseURL:  strings.TrimSpace(input.BaseURL),
		Model:    strings.TrimSpace(input.Model),
		Enabled:  true,
	}
	if err := validateAIAccount(account); err != nil {
		return nil, err
	}

	start := time.Now()
	raw, err := s.generate(account, input.RequestTimeoutSeconds, connectivityTestSystemPrompt(), connectivityTestPrompt())
	if err != nil {
		return nil, wrapAIProviderError(account.Provider, err)
	}

	response := strings.TrimSpace(raw)
	if response == "" {
		return nil, &AIProviderError{
			Provider:        account.Provider,
			Code:            "empty_response",
			UserMessage:     "The model request succeeded, but the provider returned an empty response.",
			ProviderMessage: "provider returned empty response",
			StatusCode:      http.StatusBadGateway,
		}
	}
	if len(response) > 240 {
		response = response[:240] + "..."
	}

	return &AITestAccountResult{
		Success:     true,
		Provider:    account.Provider,
		AccountID:   account.ID,
		AccountName: account.Name,
		Model:       account.Model,
		Message:     "Connection successful. The model returned a valid response.",
		Response:    response,
		LatencyMs:   time.Since(start).Milliseconds(),
	}, nil
}

func (s *aiService) requestHTTPClient(timeoutSeconds int) *http.Client {
	timeout := 90 * time.Second
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}

	if s.httpClient == nil {
		return &http.Client{Timeout: timeout}
	}

	client := *s.httpClient
	client.Timeout = timeout
	return &client
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
	// If the stored value equals the base prompt, the user never customised it —
	// treat as empty to avoid sending the base text twice.
	if custom == "" || custom == base {
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

Log sample (lines prefixed with "# /path/to/file" indicate the source file path — use the filename and directory as additional context to infer the log type, service, or component even before reading the content):
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

func validateAIAccount(account *AIAccountDTO) error {
	if account == nil {
		return fmt.Errorf("%w: ai account is required", ErrInvalidArgument)
	}
	switch account.Provider {
	case "openai", "claude", "gemini", "deepseek":
	default:
		return fmt.Errorf("%w: unsupported ai provider", ErrInvalidArgument)
	}
	if strings.TrimSpace(account.APIKey) == "" {
		return fmt.Errorf("%w: api key is required for the selected ai account", ErrInvalidArgument)
	}
	if strings.TrimSpace(account.Model) == "" {
		return fmt.Errorf("%w: model is required for the selected ai account", ErrInvalidArgument)
	}
	return nil
}

func (s *aiService) generate(account *AIAccountDTO, timeoutSeconds int, systemPrompt, prompt string) (string, error) {
	httpClient := s.requestHTTPClient(timeoutSeconds)
	switch account.Provider {
	case "openai", "deepseek":
		return s.generateOpenAICompatible(httpClient, account, systemPrompt, prompt)
	case "claude":
		return s.generateClaude(httpClient, account, systemPrompt, prompt)
	case "gemini":
		return s.generateGemini(httpClient, account, systemPrompt, prompt)
	default:
		return "", fmt.Errorf("%w: unsupported ai provider", ErrInvalidArgument)
	}
}

func wrapAIProviderError(provider string, err error) error {
	if err == nil {
		return nil
	}

	var aiErr *AIProviderError
	if errors.As(err, &aiErr) {
		return aiErr
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return newAIProviderError(provider, "timeout", http.StatusGatewayTimeout, err.Error(), err)
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		code := "network_error"
		status := http.StatusBadGateway
		if urlErr.Timeout() {
			code = "timeout"
			status = http.StatusGatewayTimeout
		}
		return newAIProviderError(provider, code, status, urlErr.Error(), err)
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		code := "network_error"
		status := http.StatusBadGateway
		if netErr.Timeout() {
			code = "timeout"
			status = http.StatusGatewayTimeout
		}
		return newAIProviderError(provider, code, status, netErr.Error(), err)
	}

	var openaiErr *openai.Error
	if errors.As(err, &openaiErr) {
		code := strings.TrimSpace(openaiErr.Code)
		if code == "" {
			code = strings.TrimSpace(openaiErr.Type)
		}
		return newAIProviderError(provider, code, openaiErr.StatusCode, openaiErr.Message, err)
	}

	var anthropicErr *anthropic.Error
	if errors.As(err, &anthropicErr) {
		code, message := parseAnthropicProviderError(anthropicErr.RawJSON())
		if message == "" {
			message = anthropicErr.Error()
		}
		return newAIProviderError(provider, code, anthropicErr.StatusCode, message, err)
	}

	var geminiErr genai.APIError
	if errors.As(err, &geminiErr) {
		code := strings.TrimSpace(geminiErr.Status)
		if code == "" && geminiErr.Code > 0 {
			code = fmt.Sprintf("http_%d", geminiErr.Code)
		}
		return newAIProviderError(provider, code, geminiErr.Code, geminiErr.Message, err)
	}

	return newAIProviderError(provider, "request_failed", http.StatusBadGateway, err.Error(), err)
}

func newAIProviderError(provider, code string, statusCode int, providerMessage string, cause error) *AIProviderError {
	providerMessage = clipProviderMessage(providerMessage)
	return &AIProviderError{
		Provider:        provider,
		Code:            normalizeProviderErrorCode(code),
		UserMessage:     userMessageForAIProviderError(statusCode, code, providerMessage),
		ProviderMessage: providerMessage,
		StatusCode:      statusCode,
		Cause:           cause,
	}
}

func normalizeProviderErrorCode(code string) string {
	code = strings.TrimSpace(strings.ToLower(code))
	if code == "" {
		return "request_failed"
	}
	code = strings.ReplaceAll(code, " ", "_")
	return code
}

func clipProviderMessage(message string) string {
	message = strings.TrimSpace(message)
	if len(message) <= 600 {
		return message
	}
	return message[:600] + "..."
}

func userMessageForAIProviderError(statusCode int, code, providerMessage string) string {
	text := strings.ToLower(strings.TrimSpace(code + " " + providerMessage))

	switch {
	case statusCode == http.StatusUnauthorized || strings.Contains(text, "authentication") || strings.Contains(text, "invalid_api_key") || strings.Contains(text, "api key"):
		return "Authentication failed. Check whether the API key is correct and still active."
	case statusCode == http.StatusForbidden || strings.Contains(text, "permission"):
		return "The account does not have permission to use this model or endpoint."
	case statusCode == http.StatusNotFound || strings.Contains(text, "model_not_found") || strings.Contains(text, "not found"):
		return "The model or endpoint could not be found. Check the model name and base URL."
	case statusCode == http.StatusTooManyRequests || strings.Contains(text, "rate_limit"):
		return "The provider rate limit was reached. Try again later or use another account."
	case statusCode == http.StatusGatewayTimeout || statusCode == http.StatusRequestTimeout || strings.Contains(text, "timeout") || strings.Contains(text, "deadline"):
		return "The AI request timed out. Check network connectivity and the base URL, then try again."
	case strings.Contains(text, "network") || strings.Contains(text, "connection refused") || strings.Contains(text, "no such host") || strings.Contains(text, "tls"):
		return "Fluent Manager could not reach the AI endpoint. Check the base URL, network, proxy, or TLS settings."
	case statusCode >= 500:
		return "The AI provider is temporarily unavailable. Try again later."
	case statusCode == http.StatusBadRequest:
		return "The provider rejected the request. Check the model name, base URL, and account configuration."
	default:
		return "The AI request failed. Check the account, model, and network settings, then try again."
	}
}

func parseAnthropicProviderError(raw string) (string, string) {
	var payload struct {
		Type  string `json:"type"`
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", ""
	}

	code := strings.TrimSpace(payload.Error.Type)
	if code == "" {
		code = strings.TrimSpace(payload.Type)
	}
	message := strings.TrimSpace(payload.Error.Message)
	if message == "" {
		message = strings.TrimSpace(payload.Message)
	}
	return code, message
}

func normalizeProviderBaseURL(account *AIAccountDTO) string {
	baseURL := strings.TrimRight(strings.TrimSpace(account.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultAIBaseURL(account.Provider)
	}
	return baseURL
}

func (s *aiService) generateOpenAICompatible(httpClient *http.Client, account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	client := openai.NewClient(
		openaioption.WithAPIKey(account.APIKey),
		openaioption.WithBaseURL(normalizeProviderBaseURL(account)),
		openaioption.WithHTTPClient(httpClient),
	)

	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.DeveloperMessage(systemPrompt),
			openai.UserMessage(prompt),
		},
		Model:       openai.ChatModel(account.Model),
		Temperature: openai.Float(0.2),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("provider returned no choices")
	}
	return resp.Choices[0].Message.Content, nil
}

func (s *aiService) generateClaude(httpClient *http.Client, account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	client := anthropic.NewClient(
		anthropicoption.WithAPIKey(account.APIKey),
		anthropicoption.WithBaseURL(normalizeProviderBaseURL(account)),
		anthropicoption.WithHTTPClient(httpClient),
	)

	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:       anthropic.Model(account.Model),
		MaxTokens:   2500,
		Temperature: anthropic.Float(0.2),
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", err
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			textBlock := block.AsText()
			if strings.TrimSpace(textBlock.Text) != "" {
				return textBlock.Text, nil
			}
		}
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("provider returned no content")
	}
	return "", fmt.Errorf("provider returned no text content")
}

func (s *aiService) generateGemini(httpClient *http.Client, account *AIAccountDTO, systemPrompt, prompt string) (string, error) {
	clientConfig := &genai.ClientConfig{
		APIKey:     account.APIKey,
		Backend:    genai.BackendGeminiAPI,
		HTTPClient: httpClient,
	}
	if baseURL := normalizeProviderBaseURL(account); baseURL != "" {
		clientConfig.HTTPOptions = genai.HTTPOptions{BaseURL: baseURL}
	}

	client, err := genai.NewClient(context.Background(), clientConfig)
	if err != nil {
		return "", err
	}

	resp, err := client.Models.GenerateContent(
		context.Background(),
		account.Model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature: genai.Ptr[float32](0.2),
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemPrompt}},
			},
		},
	)
	if err != nil {
		return "", err
	}

	text := strings.TrimSpace(resp.Text())
	if text == "" {
		return "", fmt.Errorf("provider returned no candidates")
	}
	return text, nil
}
