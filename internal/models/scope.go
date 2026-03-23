package models

import (
	"encoding/json"
	"net"
	"path/filepath"
	"strings"
)

// MatchNode checks if a node matches this rule.
func (r *ClusterMatchRule) MatchNode(hostname, ipAddress, fluentType, os string, labels string) bool {
	if !r.IsActive {
		return false
	}

	// Hostname pattern (glob)
	if r.HostnamePattern != "" {
		matched, _ := filepath.Match(r.HostnamePattern, hostname)
		if !matched {
			return false
		}
	}

	// IP pattern (CIDR or prefix glob)
	if r.IPPattern != "" {
		if !matchIP(r.IPPattern, ipAddress) {
			return false
		}
	}

	// Fluent type
	if r.FluentType != "" && r.FluentType != fluentType {
		return false
	}

	// OS pattern
	if r.OSPattern != "" {
		matched, _ := filepath.Match(r.OSPattern, strings.ToLower(os))
		if !matched {
			return false
		}
	}

	// Label selector
	if r.LabelSelector != "" {
		if !matchLabels(r.LabelSelector, labels) {
			return false
		}
	}

	return true
}

func matchIP(pattern, ip string) bool {
	// Try CIDR first (e.g. 10.0.0.0/16)
	if strings.Contains(pattern, "/") {
		_, cidr, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		return cidr.Contains(net.ParseIP(ip))
	}
	// Glob pattern (e.g. 10.0.1.*)
	matched, _ := filepath.Match(pattern, ip)
	return matched
}

func matchLabels(selectorJSON, nodeLabelsJSON string) bool {
	if nodeLabelsJSON == "" {
		return false
	}
	var selector map[string]string
	if err := json.Unmarshal([]byte(selectorJSON), &selector); err != nil {
		return false
	}
	var nodeLabels map[string]string
	if err := json.Unmarshal([]byte(nodeLabelsJSON), &nodeLabels); err != nil {
		return false
	}
	for k, v := range selector {
		if nodeLabels[k] != v {
			return false
		}
	}
	return true
}

// AutoAssignCluster finds the best matching cluster for a node, or the default cluster.
// Returns nil if no match and no default cluster exists.
func AutoAssignCluster(hostname, ipAddress, fluentType, os, labels string) *uint {
	var rules []ClusterMatchRule
	DB.Where("is_active = ?", true).Order("priority ASC").Find(&rules)

	for _, rule := range rules {
		if rule.MatchNode(hostname, ipAddress, fluentType, os, labels) {
			return &rule.ClusterID
		}
	}

	// Fallback to default cluster
	var defaultCluster Cluster
	if err := DB.Where("is_default = ?", true).First(&defaultCluster).Error; err == nil {
		return &defaultCluster.ID
	}

	return nil
}

// AllowedClusterIDs returns the cluster IDs a user can access based on their
// direct scopes and group-inherited scopes.
// Returns nil if the user has global access (no scopes at all = admin).
func AllowedClusterIDs(userID uint) []uint {
	var userScopes []UserScope
	DB.Where("user_id = ?", userID).Find(&userScopes)

	// Collect group scopes via user_groups join
	var groupScopes []GroupScope
	DB.Joins("JOIN user_groups ON user_groups.group_id = group_scopes.group_id").
		Where("user_groups.user_id = ?", userID).
		Find(&groupScopes)

	if len(userScopes) == 0 && len(groupScopes) == 0 {
		return nil // global access
	}

	clusterSet := map[uint]bool{}
	resolveScopes := func(scopeType string, scopeID uint) {
		switch scopeType {
		case "cluster":
			clusterSet[scopeID] = true
		case "region":
			var clusters []Cluster
			DB.Where("region_id = ?", scopeID).Find(&clusters)
			for _, c := range clusters {
				clusterSet[c.ID] = true
			}
		case "datacenter":
			var clusters []Cluster
			DB.Joins("JOIN regions ON regions.id = clusters.region_id").
				Where("regions.data_center_id = ?", scopeID).Find(&clusters)
			for _, c := range clusters {
				clusterSet[c.ID] = true
			}
		}
	}

	for _, s := range userScopes {
		resolveScopes(s.ScopeType, s.ScopeID)
	}
	for _, s := range groupScopes {
		resolveScopes(s.ScopeType, s.ScopeID)
	}

	ids := make([]uint, 0, len(clusterSet))
	for id := range clusterSet {
		ids = append(ids, id)
	}
	return ids
}
