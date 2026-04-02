package services

import (
	"errors"

	"github.com/fluent-manager/fluent-manager/internal/logwriter"
	"gorm.io/gorm"
)

// ErrHasChildren is returned when attempting to delete an entity that has child records.
var ErrHasChildren = errors.New("cannot delete: has child records")

// ErrInvalidArgument is returned when the request fails validation.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrConflict is returned when the request conflicts with existing state.
var ErrConflict = errors.New("conflict")

// ErrForbidden is returned when the current scope cannot access the target resource.
var ErrForbidden = errors.New("forbidden")

// Registry holds all service instances.
type Registry struct {
	Auth           AuthService
	User           UserService
	Role           RoleService
	Group          GroupService
	AuthSettings   AuthSettingsService
	AI             AIService
	Topology       TopologyService
	Node           NodeService
	Fluent         FluentService
	FluentOps      FluentOpsService
	AgentAccessKey AgentAccessKeyService
	AgentArtifact  AgentArtifactService
	AgentPolicy    AgentPolicyService
	Config         ConfigService
	Deploy         DeployService
	Bootstrap      BootstrapService
	AgentUpgrade   AgentUpgradeService
	Agent          AgentService
	Metrics        MetricsService
	Setup          SetupService
}

// NewRegistry creates all services with the given database connection.
func NewRegistry(db *gorm.DB, fluentSharedKeySecret string, agentSettings AgentSettings, bootstrapSettings BootstrapSettings, artifactDir string, logWriter *logwriter.FileLogger) *Registry {
	agentPolicySvc := NewAgentPolicyService(db, agentSettings)
	agentAccessKeySvc := NewAgentAccessKeyService(db, bootstrapSettings.Secret)
	agentArtifactSvc := NewAgentArtifactService(db, artifactDir)
	authSettingsSvc := NewAuthSettingsService(db)
	fluentOpsSvc := NewFluentOpsService(db)
	return &Registry{
		Auth:           NewAuthService(db),
		User:           NewUserService(db),
		Role:           NewRoleService(db),
		Group:          NewGroupService(db),
		AuthSettings:   authSettingsSvc,
		AI:             NewAIService(authSettingsSvc),
		Topology:       NewTopologyService(db),
		Node:           NewNodeService(db),
		Fluent:         NewFluentService(db, fluentSharedKeySecret),
		FluentOps:      fluentOpsSvc,
		AgentAccessKey: agentAccessKeySvc,
		AgentArtifact:  agentArtifactSvc,
		AgentPolicy:    agentPolicySvc,
		Config:         NewConfigService(db),
		Deploy:         NewDeployService(db, fluentOpsSvc),
		Bootstrap:      NewBootstrapService(db, bootstrapSettings, agentAccessKeySvc),
		AgentUpgrade:   NewAgentUpgradeService(db),
		Agent:          NewAgentService(db, agentPolicySvc, logWriter),
		Metrics:        NewMetricsService(db),
		Setup:          NewSetupService(db),
	}
}
