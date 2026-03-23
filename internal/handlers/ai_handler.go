package handlers

import (
	"errors"
	"net/http"

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
	existingKeys := map[string]string{}
	if existing != nil {
		for _, account := range existing.Accounts {
			existingKeys[account.ID] = account.APIKey
		}
	}
	for idx := range dto.Accounts {
		if dto.Accounts[idx].APIKey == aiMaskedSecretValue {
			dto.Accounts[idx].APIKey = existingKeys[dto.Accounts[idx].ID]
		}
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
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrInvalidArgument):
			status = http.StatusBadRequest
		case errors.Is(err, services.ErrForbidden):
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
