package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

const aiMaskedSecretValue = "********"

type AIHandler struct {
	SettingsSvc services.AuthSettingsService
	Svc         services.AIService
}

func (h *AIHandler) GetSettings(c *gin.Context) {
	settings, err := h.SettingsSvc.GetAISettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for idx := range settings.Accounts {
		if settings.Accounts[idx].APIKey != "" {
			settings.Accounts[idx].APIKey = aiMaskedSecretValue
		}
	}
	c.JSON(http.StatusOK, settings)
}

func (h *AIHandler) UpdateSettings(c *gin.Context) {
	var dto services.AISettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, _ := h.SettingsSvc.GetAISettings()
	existingKeys := aiExistingAccountKeys(existing)
	for idx := range dto.Accounts {
		dto.Accounts[idx].APIKey = restoreMaskedAISecret(dto.Accounts[idx].ID, dto.Accounts[idx].APIKey, existingKeys)
	}

	if err := h.SettingsSvc.UpdateAISettings(dto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "AI settings updated"})
}

func (h *AIHandler) AnalyzeLogSample(c *gin.Context) {
	var req services.LogSampleAnalysisInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.Svc.AnalyzeLogSample(&req)
	if err != nil {
		writeAIError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AIHandler) TestAccount(c *gin.Context) {
	var req services.AITestAccountInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, _ := h.SettingsSvc.GetAISettings()
	req.APIKey = restoreMaskedAISecret(req.ID, req.APIKey, aiExistingAccountKeys(existing))

	result, err := h.Svc.TestAccount(&req)
	if err != nil {
		writeAIError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func aiExistingAccountKeys(settings *services.AISettingsDTO) map[string]string {
	keys := map[string]string{}
	if settings == nil {
		return keys
	}
	for _, account := range settings.Accounts {
		keys[account.ID] = account.APIKey
	}
	return keys
}

func restoreMaskedAISecret(accountID, value string, existingKeys map[string]string) string {
	if value != aiMaskedSecretValue {
		return value
	}
	return existingKeys[accountID]
}

func writeAIError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	response := gin.H{
		"error": err.Error(),
	}

	var providerErr *services.AIProviderError
	switch {
	case errors.As(err, &providerErr):
		status = providerErr.HTTPStatus()
		response = gin.H{
			"error":            providerErr.Error(),
			"error_code":       providerErr.Code,
			"user_message":     providerErr.Error(),
			"provider_message": providerErr.ProviderMessage,
			"provider":         providerErr.Provider,
		}
	case errors.Is(err, services.ErrInvalidArgument):
		status = http.StatusBadRequest
		message := cleanServiceErrorMessage(err, services.ErrInvalidArgument)
		response = gin.H{
			"error":        message,
			"error_code":   "invalid_argument",
			"user_message": message,
		}
	case errors.Is(err, services.ErrForbidden):
		status = http.StatusForbidden
		message := cleanServiceErrorMessage(err, services.ErrForbidden)
		response = gin.H{
			"error":        message,
			"error_code":   "forbidden",
			"user_message": message,
		}
	}

	c.JSON(status, response)
}

func cleanServiceErrorMessage(err error, sentinel error) string {
	message := strings.TrimSpace(err.Error())
	prefix := sentinel.Error() + ":"
	if strings.HasPrefix(message, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(message, prefix))
	}
	return message
}
