package services

import (
	"errors"

	"gorm.io/gorm"
)

// ErrHasChildren is returned when attempting to delete an entity that has child records.
var ErrHasChildren = errors.New("cannot delete: has child records")

// Registry holds all service instances.
type Registry struct {
	Auth     AuthService
	User     UserService
	Role     RoleService
	Topology TopologyService
	Node     NodeService
	Config   ConfigService
	Deploy   DeployService
	Agent    AgentService
	Metrics  MetricsService
}

// NewRegistry creates all services with the given database connection.
func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{
		Auth:     NewAuthService(db),
		User:     NewUserService(db),
		Role:     NewRoleService(db),
		Topology: NewTopologyService(db),
		Node:     NewNodeService(db),
		Config:   NewConfigService(db),
		Deploy:   NewDeployService(db),
		Agent:    NewAgentService(db),
		Metrics:  NewMetricsService(db),
	}
}
