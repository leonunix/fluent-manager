package handlers

import (
	"errors"
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgentPolicyHandler struct {
	Svc     services.AgentPolicyService
	NodeSvc services.NodeService
}

func (h *AgentPolicyHandler) List(c *gin.Context) {
	policies, err := h.Svc.ListPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":     policies,
		"defaults": h.Svc.GetDefaultSettings(),
	})
}

func (h *AgentPolicyHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	policy, err := h.Svc.GetPolicy(id)
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
	policy, err := h.Svc.CreatePolicy(&req, c.GetUint("user_id"))
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, policy)
}

func (h *AgentPolicyHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req services.AgentPolicyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	policy, err := h.Svc.UpdatePolicy(id, &req)
	if err != nil {
		writeAgentPolicyError(c, err)
		return
	}
	c.JSON(http.StatusOK, policy)
}

func (h *AgentPolicyHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Svc.DeletePolicy(id); err != nil {
		writeAgentPolicyError(c, err)
		return
	}
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
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "agent policy not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
