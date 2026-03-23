package routers

import "github.com/gin-gonic/gin"

func registerSetupRoutes(api *gin.RouterGroup, h *allHandlers) {
	g := api.Group("/setup")
	g.GET("/status", h.Setup.GetStatus)
	g.POST("/test-db", h.Setup.TestDB)
	g.POST("/initialize", h.Setup.Initialize)
}
