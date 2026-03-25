package routers

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/handlers"
	"github.com/fluent-manager/fluent-manager/internal/logwriter"
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

// Deps holds all dependencies needed for route registration.
type Deps struct {
	Cfg          *config.Config
	Svc          *services.Registry
	JWTSvc       *auth.JWTService
	SAMLProvider *auth.SAMLProvider
	CfgPath      string
	RestartCh    chan struct{}
	FrontendFS   fs.FS                 // non-nil = embedded frontend (all-in-one), nil = API-only
	LogWriter    *logwriter.FileLogger // optional file logger for audit logs
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

	// Serve frontend static files (all-in-one mode only)
	if deps.FrontendFS != nil {
		frontendHTTP := http.FS(deps.FrontendFS)
		r.StaticFS("/assets", http.FS(mustSub(deps.FrontendFS, "assets")))
		r.GET("/", func(c *gin.Context) {
			c.FileFromFS("/index.html", frontendHTTP)
		})
		r.NoRoute(func(c *gin.Context) {
			c.FileFromFS("/index.html", frontendHTTP)
		})
		// Serve other root-level static files (favicon, brand, etc.)
		for _, name := range []string{"favicon.ico", "favicon.svg"} {
			n := name
			r.GET("/"+n, func(c *gin.Context) {
				c.FileFromFS("/"+n, frontendHTTP)
			})
		}
		r.StaticFS("/brand", http.FS(mustSub(deps.FrontendFS, "brand")))
		log.Println("Frontend embedded — serving SPA from binary")
	} else {
		log.Println("API-only mode — no frontend served")
	}

	// SAML routes — always registered; provider delegates dynamically
	if deps.SAMLProvider != nil {
		r.Any("/saml/*action", gin.WrapH(deps.SAMLProvider))
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
	authed.Use(middleware.AuditLog(deps.LogWriter))

	registerUserRoutes(authed, h)
	registerGroupRoutes(authed, h)
	registerAIRoutes(authed, h)
	registerTopologyRoutes(authed, h)
	registerAgentAccessKeyRoutes(authed, h)
	registerAgentArtifactRoutes(authed, h)
	registerFluentRoutes(authed, h)
	registerNodeRoutes(authed, h)
	registerConfigRoutes(authed, h)
	registerDeployRoutes(authed, h)
	registerBootstrapRoutes(authed, h)
	registerAgentUpgradeRoutes(authed, h)
	registerMetricsRoutes(authed, h)

	return r
}

// allHandlers holds all handler instances for route registration.
type allHandlers struct {
	Auth           *handlers.AuthHandler
	User           *handlers.UserHandler
	Role           *handlers.RoleHandler
	Group          *handlers.GroupHandler
	AuthSettings   *handlers.AuthSettingsHandler
	AI             *handlers.AIHandler
	Node           *handlers.NodeHandler
	Topology       *handlers.TopologyHandler
	Fluent         *handlers.FluentHandler
	FluentOps      *handlers.FluentOpsHandler
	Config         *handlers.ConfigHandler
	Deploy         *handlers.DeployHandler
	Bootstrap      *handlers.BootstrapHandler
	AgentUpgrade   *handlers.AgentUpgradeHandler
	AgentArtifact  *handlers.AgentArtifactHandler
	Agent          *handlers.AgentHandler
	AgentAccessKey *handlers.AgentAccessKeyHandler
	AgentPolicy    *handlers.AgentPolicyHandler
	Metrics        *handlers.MetricsHandler
	Setup          *handlers.SetupHandler
}

func buildHandlers(deps Deps) *allHandlers {
	var ldapAuth *auth.LDAPAuth
	if deps.Cfg.Auth.LDAP.Enabled {
		ldapAuth = auth.NewLDAPAuth(deps.Cfg.Auth.LDAP)
		log.Println("LDAP authentication enabled")
	}

	return &allHandlers{
		Auth:           &handlers.AuthHandler{JWT: deps.JWTSvc, LDAP: ldapAuth, SAMLProvider: deps.SAMLProvider, Svc: deps.Svc.Auth, AuthSettingsSvc: deps.Svc.AuthSettings},
		User:           &handlers.UserHandler{Svc: deps.Svc.User},
		Role:           &handlers.RoleHandler{Svc: deps.Svc.Role},
		Group:          &handlers.GroupHandler{Svc: deps.Svc.Group},
		AuthSettings:   &handlers.AuthSettingsHandler{Svc: deps.Svc.AuthSettings, SAMLProvider: deps.SAMLProvider},
		AI:             &handlers.AIHandler{SettingsSvc: deps.Svc.AuthSettings, Svc: deps.Svc.AI},
		Node:           &handlers.NodeHandler{Svc: deps.Svc.Node},
		Topology:       &handlers.TopologyHandler{Svc: deps.Svc.Topology},
		Fluent:         &handlers.FluentHandler{Svc: deps.Svc.Fluent, NodeSvc: deps.Svc.Node},
		FluentOps:      &handlers.FluentOpsHandler{Svc: deps.Svc.FluentOps},
		Config:         &handlers.ConfigHandler{Svc: deps.Svc.Config},
		Deploy:         &handlers.DeployHandler{Svc: deps.Svc.Deploy},
		Bootstrap:      &handlers.BootstrapHandler{Svc: deps.Svc.Bootstrap},
		AgentUpgrade:   &handlers.AgentUpgradeHandler{Svc: deps.Svc.AgentUpgrade},
		AgentArtifact:  &handlers.AgentArtifactHandler{Svc: deps.Svc.AgentArtifact},
		Agent:          &handlers.AgentHandler{Svc: deps.Svc.Agent, NodeSvc: deps.Svc.Node},
		AgentAccessKey: &handlers.AgentAccessKeyHandler{Svc: deps.Svc.AgentAccessKey},
		AgentPolicy:    &handlers.AgentPolicyHandler{Svc: deps.Svc.AgentPolicy, NodeSvc: deps.Svc.Node},
		Metrics:        &handlers.MetricsHandler{Svc: deps.Svc.Metrics, TopoSvc: deps.Svc.Topology},
		Setup:          &handlers.SetupHandler{Svc: deps.Svc.Setup, JWT: deps.JWTSvc, CfgPath: deps.CfgPath, RestartCh: deps.RestartCh},
	}
}

// mustSub returns a sub-filesystem or panics. Used for embedded frontend subdirectories.
func mustSub(parent fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(parent, dir)
	if err != nil {
		panic("frontend: missing embedded directory: " + dir)
	}
	return sub
}
