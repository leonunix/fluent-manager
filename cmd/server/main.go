package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fluent-manager/fluent-manager/internal/agent"
	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/cache"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/handlers"
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
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

	// Cache (Redis)
	cache.Init(&cfg.Cache)

	// Node monitor
	monitor := agent.NewMonitor(cfg.Agent.HeartbeatInterval)
	monitor.Start()
	defer monitor.Stop()

	// Service layer
	fluentSharedKeySecret := cfg.Fluent.SharedKeySecret
	if fluentSharedKeySecret == "" {
		fluentSharedKeySecret = cfg.Auth.JWTSecret
	}
	svc := services.NewRegistry(models.DB, fluentSharedKeySecret, services.AgentSettings{
		HeartbeatInterval:   cfg.Agent.HeartbeatInterval,
		MetricsInterval:     cfg.Agent.MetricsInterval,
		LogUploadInterval:   cfg.Agent.LogUploadInterval,
		LogBufferLines:      cfg.Agent.LogBufferLines,
		HealthPort:          cfg.Agent.HealthPort,
		MaxRetries:          cfg.Agent.MaxRetries,
		RetryBaseDelay:      cfg.Agent.RetryBaseDelay,
		FluentType:          cfg.Agent.FluentType,
		FluentConfigPath:    cfg.Agent.FluentConfigPath,
		FluentConfigDir:     cfg.Agent.FluentConfigDir,
		FluentBinary:        cfg.Agent.FluentBinary,
		FluentServiceUnit:   cfg.Agent.FluentServiceUnit,
		FluentRestartCmd:    cfg.Agent.FluentRestartCmd,
		FluentReloadCmd:     cfg.Agent.FluentReloadCmd,
		FluentDryRunCmd:     cfg.Agent.FluentDryRunCmd,
		FluentLogPath:       cfg.Agent.FluentLogPath,
		FluentMetricsURL:    cfg.Agent.FluentMetricsURL,
		FluentMetricsFormat: cfg.Agent.FluentMetricsFormat,
		BackupDir:           cfg.Agent.BackupDir,
		MaxBackups:          cfg.Agent.MaxBackups,
	})

	// Handlers
	authHandler := &handlers.AuthHandler{JWT: jwtSvc, LDAP: ldapAuth, SAML: samlAuth, Svc: svc.Auth}
	userHandler := &handlers.UserHandler{Svc: svc.User}
	roleHandler := &handlers.RoleHandler{Svc: svc.Role}
	nodeHandler := &handlers.NodeHandler{Svc: svc.Node}
	topoHandler := &handlers.TopologyHandler{Svc: svc.Topology}
	fluentHandler := &handlers.FluentHandler{Svc: svc.Fluent, NodeSvc: svc.Node}
	fluentOpsHandler := &handlers.FluentOpsHandler{Svc: svc.FluentOps}
	configHandler := &handlers.ConfigHandler{Svc: svc.Config}
	deployHandler := &handlers.DeployHandler{Svc: svc.Deploy}
	agentHandler := &handlers.AgentHandler{Svc: svc.Agent, NodeSvc: svc.Node}
	agentPolicyHandler := &handlers.AgentPolicyHandler{Svc: svc.AgentPolicy, NodeSvc: svc.Node}
	metricsHandler := &handlers.MetricsHandler{Svc: svc.Metrics, TopoSvc: svc.Topology}

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
	authed.Use(middleware.ScopeFilter())
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
			// Match rules
			clusters.GET("/:id/rules", middleware.RequirePermission("topology", "read"), topoHandler.ListMatchRules)
			clusters.POST("/:id/rules", middleware.RequirePermission("topology", "create"), topoHandler.CreateMatchRule)
			clusters.PUT("/:id/rules/:rule_id", middleware.RequirePermission("topology", "update"), topoHandler.UpdateMatchRule)
			clusters.DELETE("/:id/rules/:rule_id", middleware.RequirePermission("topology", "delete"), topoHandler.DeleteMatchRule)
		}

		// User Scopes (admin only)
		authed.GET("/users/:id/scopes", middleware.RequirePermission("users", "read"), topoHandler.ListUserScopes)
		authed.PUT("/users/:id/scopes", middleware.RequirePermission("users", "update"), topoHandler.SetUserScopes)

		// Fluent aggregation groups
		aggGroups := authed.Group("/aggregation-groups")
		{
			aggGroups.GET("", middleware.RequirePermission("topology", "read"), fluentHandler.ListAggregationGroups)
			aggGroups.GET("/deleted", middleware.RequirePermission("topology", "read"), fluentHandler.ListDeletedAggregationGroups)
			aggGroups.GET("/:id", middleware.RequirePermission("topology", "read"), fluentHandler.GetAggregationGroup)
			aggGroups.POST("", middleware.RequirePermission("topology", "create"), fluentHandler.CreateAggregationGroup)
			aggGroups.PUT("/:id", middleware.RequirePermission("topology", "update"), fluentHandler.UpdateAggregationGroup)
			aggGroups.DELETE("/:id", middleware.RequirePermission("topology", "delete"), fluentHandler.DeleteAggregationGroup)
			aggGroups.POST("/:id/restore", middleware.RequirePermission("topology", "update"), fluentHandler.RestoreAggregationGroup)
			aggGroups.GET("/:id/metrics", middleware.RequirePermission("topology", "read"), fluentOpsHandler.AggregationGroupMetrics)
		}

		// Fluent log pipelines and flow graph
		pipelines := authed.Group("/log-pipelines")
		{
			pipelines.GET("", middleware.RequirePermission("topology", "read"), fluentOpsHandler.ListPipelines)
			pipelines.GET("/graph", middleware.RequirePermission("topology", "read"), fluentOpsHandler.PipelineGraph)
			pipelines.GET("/:id", middleware.RequirePermission("topology", "read"), fluentOpsHandler.GetPipeline)
			pipelines.POST("", middleware.RequirePermission("topology", "create"), fluentOpsHandler.CreatePipeline)
			pipelines.PUT("/:id", middleware.RequirePermission("topology", "update"), fluentOpsHandler.UpdatePipeline)
			pipelines.DELETE("/:id", middleware.RequirePermission("topology", "delete"), fluentOpsHandler.DeletePipeline)
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
			nodes.GET("/:id/fluent-profile", middleware.RequirePermission("nodes", "read"), fluentHandler.GetNodeProfile)
			nodes.PUT("/:id/fluent-profile", middleware.RequirePermission("nodes", "update"), fluentHandler.UpdateNodeProfile)
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
			configs.GET("/modules", middleware.RequirePermission("configs", "read"), configHandler.ListModules)
			configs.GET("/modules/:id", middleware.RequirePermission("configs", "read"), configHandler.GetModule)
			configs.POST("/modules", middleware.RequirePermission("configs", "create"), configHandler.CreateModule)
			configs.PUT("/modules/:id", middleware.RequirePermission("configs", "update"), configHandler.UpdateModule)
			configs.DELETE("/modules/:id", middleware.RequirePermission("configs", "delete"), configHandler.DeleteModule)
			configs.GET("/modules/:id/versions", middleware.RequirePermission("configs", "read"), configHandler.ListModuleVersions)
			configs.POST("/modules/:id/versions", middleware.RequirePermission("configs", "create"), configHandler.CreateModuleVersion)
			configs.POST("/rendered-configs/preview", middleware.RequirePermission("configs", "read"), configHandler.PreviewRenderedConfig)
			configs.GET("/rendered-configs/:id", middleware.RequirePermission("configs", "read"), configHandler.GetRenderedConfig)
		}

		analysis := authed.Group("/config-analysis")
		{
			analysis.POST("/lint", middleware.RequirePermission("configs", "read"), fluentOpsHandler.LintConfig)
			analysis.POST("/replay", middleware.RequirePermission("configs", "read"), fluentOpsHandler.ReplayConfig)
			analysis.POST("/diff", middleware.RequirePermission("configs", "read"), fluentOpsHandler.SemanticDiff)
			analysis.POST("/compatibility", middleware.RequirePermission("configs", "read"), fluentOpsHandler.CheckCompatibility)
			analysis.GET("/:id", middleware.RequirePermission("configs", "read"), fluentOpsHandler.GetAnalysisResult)
		}

		// Deployments
		deploys := authed.Group("/deploys")
		{
			deploys.GET("", middleware.RequirePermission("configs", "read"), deployHandler.List)
			deploys.GET("/:id", middleware.RequirePermission("configs", "read"), deployHandler.Get)
			deploys.POST("", middleware.RequirePermission("configs", "deploy"), deployHandler.Create)
		}

		// Aggregated Metrics
		metrics := authed.Group("/metrics")
		{
			metrics.GET("/overview", middleware.RequirePermission("nodes", "read"), metricsHandler.Overview)
			metrics.GET("/top-nodes", middleware.RequirePermission("nodes", "read"), metricsHandler.TopNodes)
			metrics.GET("/by-datacenter", middleware.RequirePermission("nodes", "read"), metricsHandler.ByDatacenter)
		}

		runtime := authed.Group("/runtime")
		{
			runtime.GET("/drift", middleware.RequirePermission("nodes", "read"), fluentOpsHandler.RuntimeDrift)
			runtime.GET("/health/graph", middleware.RequirePermission("nodes", "read"), fluentOpsHandler.RuntimeHealthGraph)
			runtime.GET("/recommendations", middleware.RequirePermission("nodes", "read"), fluentOpsHandler.RuntimeRecommendations)
		}

		agentPolicies := authed.Group("/agent-policies")
		{
			agentPolicies.GET("", middleware.RequirePermission("configs", "read"), agentPolicyHandler.List)
			agentPolicies.GET("/:id", middleware.RequirePermission("configs", "read"), agentPolicyHandler.Get)
			agentPolicies.POST("", middleware.RequirePermission("configs", "update"), agentPolicyHandler.Create)
			agentPolicies.PUT("/:id", middleware.RequirePermission("configs", "update"), agentPolicyHandler.Update)
			agentPolicies.DELETE("/:id", middleware.RequirePermission("configs", "update"), agentPolicyHandler.Delete)
			agentPolicies.GET("/resolve/:node_id", middleware.RequirePermission("nodes", "read"), agentPolicyHandler.ResolveForNode)
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
