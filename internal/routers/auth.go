package routers

import (
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(api *gin.RouterGroup, h *allHandlers) {
	api.POST("/auth/login", h.Auth.Login)
	api.GET("/auth/saml/callback", h.Auth.SAMLCallback)
	api.POST("/auth/saml/exchange", h.Auth.ExchangeSAMLCode)
	api.GET("/auth/methods", h.AuthSettings.GetEnabledMethods)
}
