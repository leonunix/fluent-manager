package routers

import (
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(api *gin.RouterGroup, h *allHandlers) {
	api.POST("/auth/login", h.Auth.Login)
}
