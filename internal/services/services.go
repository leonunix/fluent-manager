package services

import (
	"errors"

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
	Auth         AuthService
	User         UserService
	Role         RoleService
	Group        GroupService
	AuthSettings AuthSettingsService
	AI           AIService
	Topology     TopologyService
	Node         NodeService
	Fluent       FluentService
	FluentOps    FluentOpsService
	AgentPolicy  AgentPolicyService
	Config       ConfigService
	Deploy       DeployService
	Agent        AgentService
	Metrics      MetricsService
	Setup        SetupService
}

// NewRegistry creates all services with the given database connection.
func NewRegistry(db *gorm.DB, fluentSharedKeySecret string, agentSettings AgentSettings) *Registry {
	agentPolicySvc := NewAgentPolicyService(db, agentSettings)
	authSettingsSvc := NewAuthSettingsService(db)
	return &Registry{
		Auth:         NewAuthService(db),
		User:         NewUserService(db),
		Role:         NewRoleService(db),
		Group:        NewGroupService(db),
		AuthSettings: authSettingsSvc,
		AI:           NewAIService(authSettingsSvc),
		Topology:     NewTopologyService(db),
		Node:         NewNodeService(db),
		Fluent:       NewFluentService(db, fluentSharedKeySecret),
		FluentOps:    NewFluentOpsService(db),
		AgentPolicy:  agentPolicySvc,
		Config:       NewConfigService(db),
		Deploy:       NewDeployService(db),
		Agent:        NewAgentService(db, agentPolicySvc),
		Metrics:      NewMetricsService(db),
		Setup:        NewSetupService(db),
	}
}
