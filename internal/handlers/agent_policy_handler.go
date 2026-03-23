package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgentPolicyHandler struct {
	Svc     services.AgentPolicyService
	NodeSvc services.NodeService
}

func (h *AgentPolicyHandler) List(c *gin.Context) {
	policies, err := h.Svc.ListPolicies(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":                policies,
		"defaults":            h.Svc.GetDefaultSettings(),
		"allowed_scope_types": allowedAgentPolicyScopeTypes(middleware.GetAllowedClusters(c)),
	})
}

func (h *AgentPolicyHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	policy, err := h.Svc.GetPolicy(id, middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	c.JSON(http.StatusOK, policy)
}

func (h *AgentPolicyHandler) Create(c *gin.Context) {
	var req services.AgentPolicyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	policy, err := h.Svc.CreatePolicy(&req, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	applyAgentPolicyAudit(c, nil, policy, "create")
	c.JSON(http.StatusCreated, policy)
}

func (h *AgentPolicyHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	allowedClusters := middleware.GetAllowedClusters(c)
	before, err := h.Svc.GetPolicy(id, allowedClusters)
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	var req services.AgentPolicyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	policy, err := h.Svc.UpdatePolicy(id, &req, allowedClusters)
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	applyAgentPolicyAudit(c, before, policy, "update")
	c.JSON(http.StatusOK, policy)
}

func (h *AgentPolicyHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	allowedClusters := middleware.GetAllowedClusters(c)
	before, err := h.Svc.GetPolicy(id, allowedClusters)
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	if err := h.Svc.DeletePolicy(id, allowedClusters); err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	applyAgentPolicyAudit(c, before, nil, "delete")
	c.JSON(http.StatusOK, gin.H{"message": "agent policy deleted"})
}

func (h *AgentPolicyHandler) ResolveForNode(c *gin.Context) {
	nodeID, ok := parseUintParam(c, "node_id")
	if !ok {
		return
	}
	node, err := h.NodeSvc.Get(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	resolved, err := h.Svc.ResolveForNodeID(nodeID)
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resolved)
}

func writeAgentPolicyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "agent policy not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func allowedAgentPolicyScopeTypes(allowedClusters []uint) []string {
	if allowedClusters == nil {
		return []string{"global", "environment", "cluster", "label_selector"}
	}
	return []string{"cluster"}
}

type agentPolicyAuditSnapshot struct {
	ID            uint                        `json:"id"`
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	ScopeType     string                      `json:"scope_type"`
	EnvironmentID *uint                       `json:"environment_id,omitempty"`
	ClusterID     *uint                       `json:"cluster_id,omitempty"`
	LabelSelector string                      `json:"label_selector"`
	Priority      int                         `json:"priority"`
	IsEnabled     bool                        `json:"is_enabled"`
	Settings      services.AgentSettingsPatch `json:"settings"`
}

type agentPolicyAuditChange struct {
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
}

type agentPolicyAuditPayload struct {
	Operation string                            `json:"operation"`
	Before    *agentPolicyAuditSnapshot         `json:"before,omitempty"`
	After     *agentPolicyAuditSnapshot         `json:"after,omitempty"`
	Changes   map[string]agentPolicyAuditChange `json:"changes,omitempty"`
}

func applyAgentPolicyAudit(c *gin.Context, before, after *services.AgentPolicyView, operation string) {
	payload := agentPolicyAuditPayload{
		Operation: operation,
		Before:    toAgentPolicyAuditSnapshot(before),
		After:     toAgentPolicyAuditSnapshot(after),
		Changes:   buildAgentPolicyAuditChanges(before, after),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	resourceID := uint(0)
	if after != nil {
		resourceID = after.ID
	} else if before != nil {
		resourceID = before.ID
	}
	middleware.SetAuditResource(c, "agent_policy", resourceID)
	middleware.SetAuditDetail(c, string(data))
}

func toAgentPolicyAuditSnapshot(policy *services.AgentPolicyView) *agentPolicyAuditSnapshot {
	if policy == nil {
		return nil
	}
	return &agentPolicyAuditSnapshot{
		ID:            policy.ID,
		Name:          policy.Name,
		Description:   policy.Description,
		ScopeType:     policy.ScopeType,
		EnvironmentID: policy.EnvironmentID,
		ClusterID:     policy.ClusterID,
		LabelSelector: policy.LabelSelector,
		Priority:      policy.Priority,
		IsEnabled:     policy.IsEnabled,
		Settings:      policy.Settings,
	}
}

func buildAgentPolicyAuditChanges(before, after *services.AgentPolicyView) map[string]agentPolicyAuditChange {
	if before == nil || after == nil {
		return nil
	}

	changes := map[string]agentPolicyAuditChange{}
	addAgentPolicyAuditChange(changes, "name", before.Name, after.Name)
	addAgentPolicyAuditChange(changes, "description", before.Description, after.Description)
	addAgentPolicyAuditChange(changes, "scope_type", before.ScopeType, after.ScopeType)
	addAgentPolicyAuditChange(changes, "environment_id", before.EnvironmentID, after.EnvironmentID)
	addAgentPolicyAuditChange(changes, "cluster_id", before.ClusterID, after.ClusterID)
	addAgentPolicyAuditChange(changes, "label_selector", before.LabelSelector, after.LabelSelector)
	addAgentPolicyAuditChange(changes, "priority", before.Priority, after.Priority)
	addAgentPolicyAuditChange(changes, "is_enabled", before.IsEnabled, after.IsEnabled)
	addAgentPolicyAuditChange(changes, "settings", before.Settings, after.Settings)
	if len(changes) == 0 {
		return nil
	}
	return changes
}

func addAgentPolicyAuditChange(changes map[string]agentPolicyAuditChange, field string, before, after interface{}) {
	if reflect.DeepEqual(before, after) {
		return
	}
	changes[field] = agentPolicyAuditChange{
		Before: before,
		After:  after,
	}
}
