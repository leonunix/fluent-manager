package routers

import (
	"log"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/handlers"
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

// Deps holds all dependencies needed for route registration.
type Deps struct {
	Cfg       *config.Config
	Svc       *services.Registry
	JWTSvc    *auth.JWTService
	SAMLAuth  *auth.SAMLAuth
	CfgPath   string
	RestartCh chan struct{}
}

// SetupRouter creates the gin.Engine with all routes registered.
func SetupRouter(deps Deps) *gin.Engine {
	cfg := deps.Cfg

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

	// SAML routes
	if deps.SAMLAuth != nil && deps.SAMLAuth.SP != nil {
		r.Any("/saml/*action", gin.WrapH(deps.SAMLAuth.SP))
	}

	api := r.Group("/api/v1")

	// Build handlers
	h := buildHandlers(deps)

	// Public routes
	registerAuthRoutes(api, h)
	registerSetupRoutes(api, h)

	// Agent API (authenticated via API key)
	registerAgentRoutes(api, cfg, h)

	// Authenticated routes
	authed := api.Group("")
	authed.Use(middleware.JWTAuth(deps.JWTSvc))
	authed.Use(middleware.ScopeFilter())
	authed.Use(middleware.AuditLog())

	registerUserRoutes(authed, h)
	registerTopologyRoutes(authed, h)
	registerFluentRoutes(authed, h)
	registerNodeRoutes(authed, h)
	registerConfigRoutes(authed, h)
	registerDeployRoutes(authed, h)
	registerMetricsRoutes(authed, h)

	return r
}

// allHandlers holds all handler instances for route registration.
type allHandlers struct {
	Auth        *handlers.AuthHandler
	User        *handlers.UserHandler
	Role        *handlers.RoleHandler
	Node        *handlers.NodeHandler
	Topology    *handlers.TopologyHandler
	Fluent      *handlers.FluentHandler
	FluentOps   *handlers.FluentOpsHandler
	Config      *handlers.ConfigHandler
	Deploy      *handlers.DeployHandler
	Agent       *handlers.AgentHandler
	AgentPolicy *handlers.AgentPolicyHandler
	Metrics     *handlers.MetricsHandler
	Setup       *handlers.SetupHandler
}

func buildHandlers(deps Deps) *allHandlers {
	var ldapAuth *auth.LDAPAuth
	if deps.Cfg.Auth.LDAP.Enabled {
		ldapAuth = auth.NewLDAPAuth(deps.Cfg.Auth.LDAP)
		log.Println("LDAP authentication enabled")
	}

	return &allHandlers{
		Auth:        &handlers.AuthHandler{JWT: deps.JWTSvc, LDAP: ldapAuth, SAML: deps.SAMLAuth, Svc: deps.Svc.Auth},
		User:        &handlers.UserHandler{Svc: deps.Svc.User},
		Role:        &handlers.RoleHandler{Svc: deps.Svc.Role},
		Node:        &handlers.NodeHandler{Svc: deps.Svc.Node},
		Topology:    &handlers.TopologyHandler{Svc: deps.Svc.Topology},
		Fluent:      &handlers.FluentHandler{Svc: deps.Svc.Fluent, NodeSvc: deps.Svc.Node},
		FluentOps:   &handlers.FluentOpsHandler{Svc: deps.Svc.FluentOps},
		Config:      &handlers.ConfigHandler{Svc: deps.Svc.Config},
		Deploy:      &handlers.DeployHandler{Svc: deps.Svc.Deploy},
		Agent:       &handlers.AgentHandler{Svc: deps.Svc.Agent, NodeSvc: deps.Svc.Node},
		AgentPolicy: &handlers.AgentPolicyHandler{Svc: deps.Svc.AgentPolicy, NodeSvc: deps.Svc.Node},
		Metrics:     &handlers.MetricsHandler{Svc: deps.Svc.Metrics, TopoSvc: deps.Svc.Topology},
		Setup:       &handlers.SetupHandler{Svc: deps.Svc.Setup, JWT: deps.JWTSvc, CfgPath: deps.CfgPath, RestartCh: deps.RestartCh},
	}
}
