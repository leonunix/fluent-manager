package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fluent-manager/fluent-manager/internal/agent"
	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/handlers"
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database
	if err := models.InitDB(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Auth services
	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenExpireHours)

	var ldapAuth *auth.LDAPAuth
	if cfg.Auth.LDAP.Enabled {
		ldapAuth = auth.NewLDAPAuth(cfg.Auth.LDAP)
		log.Println("LDAP authentication enabled")
	}

	var samlAuth *auth.SAMLAuth
	if cfg.Auth.SAML.Enabled {
		samlAuth, err = auth.NewSAMLAuth(cfg.Auth.SAML)
		if err != nil {
			log.Printf("WARNING: SAML init failed: %v", err)
		} else {
			log.Println("SAML authentication enabled")
		}
	}

	// Node monitor
	monitor := agent.NewMonitor(cfg.Agent.HeartbeatInterval)
	monitor.Start()
	defer monitor.Stop()

	// Handlers
	authHandler := &handlers.AuthHandler{JWT: jwtSvc, LDAP: ldapAuth, SAML: samlAuth}
	userHandler := &handlers.UserHandler{}
	roleHandler := &handlers.RoleHandler{}
	nodeHandler := &handlers.NodeHandler{}
	topoHandler := &handlers.TopologyHandler{}
	configHandler := &handlers.ConfigHandler{}
	deployHandler := &handlers.DeployHandler{}
	agentHandler := &handlers.AgentHandler{}

	// Router
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Agent-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve frontend static files
	r.Static("/assets", "./web/frontend/dist/assets")
	r.StaticFile("/", "./web/frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/frontend/dist/index.html")
	})

	// API routes
	api := r.Group("/api/v1")

	// Public routes
	api.POST("/auth/login", authHandler.Login)

	// SAML routes
	if samlAuth != nil && samlAuth.SP != nil {
		r.Any("/saml/*action", gin.WrapH(samlAuth.SP))
	}

	// Agent API (authenticated via API key)
	agentAPI := api.Group("/agent")
	agentAPI.Use(middleware.AgentAuth(cfg.Agent.APIKey))
	{
		agentAPI.POST("/register", agentHandler.Register)
		agentAPI.POST("/heartbeat", agentHandler.Heartbeat)
		agentAPI.POST("/report", agentHandler.ReportStatus)
		agentAPI.POST("/command-result", agentHandler.ReportCommandResult)
		agentAPI.POST("/logs", agentHandler.UploadLogs)
	}

	// Authenticated routes
	authed := api.Group("")
	authed.Use(middleware.JWTAuth(jwtSvc))
	authed.Use(middleware.AuditLog())
	{
		// Profile
		authed.GET("/auth/profile", authHandler.GetProfile)
		authed.PUT("/auth/password", authHandler.ChangePassword)

		// Users
		users := authed.Group("/users")
		{
			users.GET("", middleware.RequirePermission("users", "read"), userHandler.List)
			users.GET("/:id", middleware.RequirePermission("users", "read"), userHandler.Get)
			users.POST("", middleware.RequirePermission("users", "create"), userHandler.Create)
			users.PUT("/:id", middleware.RequirePermission("users", "update"), userHandler.Update)
			users.DELETE("/:id", middleware.RequirePermission("users", "delete"), userHandler.Delete)
		}

		// Roles & Permissions
		roles := authed.Group("/roles")
		{
			roles.GET("", middleware.RequirePermission("roles", "read"), roleHandler.List)
			roles.GET("/:id", middleware.RequirePermission("roles", "read"), roleHandler.Get)
			roles.POST("", middleware.RequirePermission("roles", "create"), roleHandler.Create)
			roles.PUT("/:id", middleware.RequirePermission("roles", "update"), roleHandler.Update)
			roles.DELETE("/:id", middleware.RequirePermission("roles", "delete"), roleHandler.Delete)
		}
		authed.GET("/permissions", middleware.RequirePermission("roles", "read"), roleHandler.ListPermissions)

		// ---- Topology: DataCenter → Region → Cluster ----
		authed.GET("/topology/tree", middleware.RequirePermission("topology", "read"), topoHandler.GetTree)

		// Environments
		envs := authed.Group("/environments")
		{
			envs.GET("", middleware.RequirePermission("topology", "read"), topoHandler.ListEnvironments)
			envs.POST("", middleware.RequirePermission("topology", "create"), topoHandler.CreateEnvironment)
			envs.PUT("/:id", middleware.RequirePermission("topology", "update"), topoHandler.UpdateEnvironment)
			envs.DELETE("/:id", middleware.RequirePermission("topology", "delete"), topoHandler.DeleteEnvironment)
		}

		// DataCenters
		dcs := authed.Group("/datacenters")
		{
			dcs.GET("", middleware.RequirePermission("topology", "read"), topoHandler.ListDataCenters)
			dcs.GET("/:id", middleware.RequirePermission("topology", "read"), topoHandler.GetDataCenter)
			dcs.POST("", middleware.RequirePermission("topology", "create"), topoHandler.CreateDataCenter)
			dcs.PUT("/:id", middleware.RequirePermission("topology", "update"), topoHandler.UpdateDataCenter)
			dcs.DELETE("/:id", middleware.RequirePermission("topology", "delete"), topoHandler.DeleteDataCenter)
		}

		// Regions
		regions := authed.Group("/regions")
		{
			regions.GET("", middleware.RequirePermission("topology", "read"), topoHandler.ListRegions)
			regions.GET("/:id", middleware.RequirePermission("topology", "read"), topoHandler.GetRegion)
			regions.POST("", middleware.RequirePermission("topology", "create"), topoHandler.CreateRegion)
			regions.PUT("/:id", middleware.RequirePermission("topology", "update"), topoHandler.UpdateRegion)
			regions.DELETE("/:id", middleware.RequirePermission("topology", "delete"), topoHandler.DeleteRegion)
		}

		// Clusters
		clusters := authed.Group("/clusters")
		{
			clusters.GET("", middleware.RequirePermission("topology", "read"), topoHandler.ListClusters)
			clusters.GET("/:id", middleware.RequirePermission("topology", "read"), topoHandler.GetCluster)
			clusters.POST("", middleware.RequirePermission("topology", "create"), topoHandler.CreateCluster)
			clusters.PUT("/:id", middleware.RequirePermission("topology", "update"), topoHandler.UpdateCluster)
			clusters.DELETE("/:id", middleware.RequirePermission("topology", "delete"), topoHandler.DeleteCluster)
		}

		// Nodes
		nodes := authed.Group("/nodes")
		{
			nodes.GET("", middleware.RequirePermission("nodes", "read"), nodeHandler.List)
			nodes.GET("/stats", middleware.RequirePermission("nodes", "read"), nodeHandler.Stats)
			nodes.GET("/:id", middleware.RequirePermission("nodes", "read"), nodeHandler.Get)
			nodes.PUT("/:id", middleware.RequirePermission("nodes", "update"), nodeHandler.Update)
			nodes.DELETE("/:id", middleware.RequirePermission("nodes", "delete"), nodeHandler.Delete)
			nodes.POST("/batch-move", middleware.RequirePermission("nodes", "update"), nodeHandler.BatchMoveCluster)
			// Node metrics, logs, remote commands
			nodes.GET("/:id/metrics", middleware.RequirePermission("nodes", "read"), agentHandler.GetNodeMetrics)
			nodes.GET("/:id/logs", middleware.RequirePermission("nodes", "read"), agentHandler.GetNodeLogs)
			nodes.POST("/:id/commands", middleware.RequirePermission("nodes", "update"), agentHandler.SendCommand)
			nodes.GET("/:id/commands", middleware.RequirePermission("nodes", "read"), agentHandler.ListNodeCommands)
		}

		// Config Templates
		configs := authed.Group("/configs")
		{
			configs.GET("/templates", middleware.RequirePermission("configs", "read"), configHandler.ListTemplates)
			configs.GET("/templates/:id", middleware.RequirePermission("configs", "read"), configHandler.GetTemplate)
			configs.POST("/templates", middleware.RequirePermission("configs", "create"), configHandler.CreateTemplate)
			configs.PUT("/templates/:id", middleware.RequirePermission("configs", "update"), configHandler.UpdateTemplate)
			configs.DELETE("/templates/:id", middleware.RequirePermission("configs", "delete"), configHandler.DeleteTemplate)
			configs.GET("/templates/:id/versions", middleware.RequirePermission("configs", "read"), configHandler.ListVersions)
			configs.POST("/templates/:id/versions", middleware.RequirePermission("configs", "create"), configHandler.CreateVersion)
			configs.GET("/versions/:version_id", middleware.RequirePermission("configs", "read"), configHandler.GetVersion)
		}

		// Deployments
		deploys := authed.Group("/deploys")
		{
			deploys.GET("", middleware.RequirePermission("configs", "read"), deployHandler.List)
			deploys.GET("/:id", middleware.RequirePermission("configs", "read"), deployHandler.Get)
			deploys.POST("", middleware.RequirePermission("configs", "deploy"), deployHandler.Create)
		}

		// Audit Logs
		authed.GET("/audit-logs", middleware.RequirePermission("audit", "read"), deployHandler.GetAuditLogs)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Fluent Manager server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
