package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type ConfigReplayInput struct {
	FluentType     string `json:"fluent_type"`
	RuntimeVersion string `json:"runtime_version"`
	Content        string `json:"content"`
	SampleLog      string `json:"sample_log"`
	SampleTag      string `json:"sample_tag"`
}

type ConfigReplayStep struct {
	Stage  string `json:"stage"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type ConfigReplayResult struct {
	FluentType      string                 `json:"fluent_type"`
	RuntimeVersion  string                 `json:"runtime_version"`
	SampleTag       string                 `json:"sample_tag"`
	DetectedParser  string                 `json:"detected_parser"`
	ParsedRecord    map[string]interface{} `json:"parsed_record"`
	MatchedFilters  []string               `json:"matched_filters"`
	RouteMatched    bool                   `json:"route_matched"`
	FinalOutput     string                 `json:"final_output"`
	FinalOutputType string                 `json:"final_output_type"`
	Warnings        []string               `json:"warnings"`
	Steps           []ConfigReplayStep     `json:"steps"`
}

type ConfigSemanticDiffInput struct {
	FluentType    string `json:"fluent_type"`
	BeforeContent string `json:"before_content"`
	AfterContent  string `json:"after_content"`
}

type SemanticChange struct {
	Category   string `json:"category"`
	ChangeType string `json:"change_type"`
	Item       string `json:"item"`
	Detail     string `json:"detail"`
}

type ConfigSemanticDiffResult struct {
	FluentType string           `json:"fluent_type"`
	Summary    string           `json:"summary"`
	Changes    []SemanticChange `json:"changes"`
}

type CompatibilityCheckInput struct {
	FluentType     string `json:"fluent_type"`
	RuntimeVersion string `json:"runtime_version"`
	Content        string `json:"content"`
	NodeID         *uint  `json:"node_id"`
}

type CompatibilityCheckResult struct {
	FluentType         string                         `json:"fluent_type"`
	RuntimeVersion     string                         `json:"runtime_version"`
	Compatible         bool                           `json:"compatible"`
	HotReloadSupported bool                           `json:"hot_reload_supported"`
	CheckedNodeID      *uint                          `json:"checked_node_id,omitempty"`
	MissingPlugins     []string                       `json:"missing_plugins"`
	Findings           []models.ConfigAnalysisFinding `json:"findings"`
}

type RuntimeRecommendation struct {
	Severity   string `json:"severity"`
	ScopeType  string `json:"scope_type"`
	ScopeID    uint   `json:"scope_id"`
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion"`
}

type semanticSnapshot struct {
	Inputs      []string
	Parsers     []string
	Filters     []string
	Routes      []string
	Outputs     []string
	Plugins     []string
	FilterRules []matchRule
	OutputRules []matchRule
}

type matchRule struct {
	Pattern string
	Name    string
	Type    string
}

var wildcardRegexCache sync.Map

func (s *fluentOpsService) ReplayConfig(input *ConfigReplayInput) (*ConfigReplayResult, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: replay payload is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validRenderedConfigTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}
	sampleLog := strings.TrimSpace(input.SampleLog)
	if sampleLog == "" {
		return nil, fmt.Errorf("%w: sample_log is required", ErrInvalidArgument)
	}

	tag := strings.TrimSpace(input.SampleTag)
	if tag == "" {
		tag = "replay.sample"
	}

	snapshot := analyzeConfigSemantics(fluentType, content)
	record, parser := parseReplaySample(sampleLog, snapshot.Parsers)
	steps := []ConfigReplayStep{
		{
			Stage:  "ingest",
			Name:   "sample accepted",
			Status: "ok",
			Detail: "sample log received for replay",
		},
		{
			Stage:  "parse",
			Name:   parser,
			Status: "ok",
			Detail: fmt.Sprintf("record materialized with %d field(s)", len(record)),
		},
	}

	matchedFilters := make([]string, 0, len(snapshot.FilterRules))
	for _, rule := range snapshot.FilterRules {
		if matchTagPattern(rule.Pattern, tag) {
			matchedFilters = append(matchedFilters, rule.Name)
			steps = append(steps, ConfigReplayStep{
				Stage:  "filter",
				Name:   rule.Name,
				Status: "matched",
				Detail: fmt.Sprintf("tag %q matched filter pattern %q", tag, rule.Pattern),
			})
		}
	}

	result := &ConfigReplayResult{
		FluentType:     fluentType,
		RuntimeVersion: strings.TrimSpace(input.RuntimeVersion),
		SampleTag:      tag,
		DetectedParser: parser,
		ParsedRecord:   record,
		MatchedFilters: uniqueSorted(matchedFilters),
		Warnings:       []string{},
		Steps:          steps,
	}

	for _, rule := range snapshot.OutputRules {
		if matchTagPattern(rule.Pattern, tag) {
			result.RouteMatched = true
			result.FinalOutput = rule.Name
			result.FinalOutputType = rule.Type
			result.Steps = append(result.Steps, ConfigReplayStep{
				Stage:  "route",
				Name:   rule.Name,
				Status: "matched",
				Detail: fmt.Sprintf("tag %q routed by pattern %q", tag, rule.Pattern),
			})
			return result, nil
		}
	}

	result.Warnings = append(result.Warnings, "sample tag did not match any output route in the baseline replay engine")
	result.Steps = append(result.Steps, ConfigReplayStep{
		Stage:  "route",
		Name:   "no route",
		Status: "warning",
		Detail: fmt.Sprintf("tag %q did not match any output rule", tag),
	})
	return result, nil
}

func (s *fluentOpsService) SemanticDiff(input *ConfigSemanticDiffInput) (*ConfigSemanticDiffResult, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: diff payload is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validRenderedConfigTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}

	before := analyzeConfigSemantics(fluentType, strings.TrimSpace(input.BeforeContent))
	after := analyzeConfigSemantics(fluentType, strings.TrimSpace(input.AfterContent))

	changes := make([]SemanticChange, 0, 12)
	changes = append(changes, diffSemanticCategory("input", before.Inputs, after.Inputs)...)
	changes = append(changes, diffSemanticCategory("parser", before.Parsers, after.Parsers)...)
	changes = append(changes, diffSemanticCategory("filter", before.Filters, after.Filters)...)
	changes = append(changes, diffSemanticCategory("route", before.Routes, after.Routes)...)
	changes = append(changes, diffSemanticCategory("output", before.Outputs, after.Outputs)...)
	changes = append(changes, diffSemanticCategory("plugin", before.Plugins, after.Plugins)...)

	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Category != changes[j].Category {
			return changes[i].Category < changes[j].Category
		}
		if changes[i].ChangeType != changes[j].ChangeType {
			return changes[i].ChangeType < changes[j].ChangeType
		}
		return changes[i].Item < changes[j].Item
	})

	added, removed := 0, 0
	for _, change := range changes {
		if change.ChangeType == "added" {
			added++
		} else if change.ChangeType == "removed" {
			removed++
		}
	}

	summary := "no semantic changes detected"
	if len(changes) > 0 {
		summary = fmt.Sprintf("%d change(s): %d added, %d removed", len(changes), added, removed)
	}

	return &ConfigSemanticDiffResult{
		FluentType: fluentType,
		Summary:    summary,
		Changes:    changes,
	}, nil
}

func (s *fluentOpsService) CheckCompatibility(input *CompatibilityCheckInput, allowedClusters []uint) (*CompatibilityCheckResult, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: compatibility payload is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validRenderedConfigTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}

	snapshot := analyzeConfigSemantics(fluentType, content)
	findings := make([]models.ConfigAnalysisFinding, 0, 8)
	missingPlugins := []string{}
	hotReloadSupported := false
	lowerContent := strings.ToLower(content)

	if strings.TrimSpace(input.RuntimeVersion) == "" {
		findings = append(findings, models.ConfigAnalysisFinding{
			Severity:   "info",
			RuleCode:   "RUNTIME_VERSION_UNSET",
			Message:    "runtime version is not provided, version-specific compatibility checks are limited",
			Suggestion: "supply the target Fluent Bit / Fluentd version for stricter checks",
			Line:       1,
		})
	}

	if input.NodeID != nil {
		var node models.Node
		if err := s.db.Preload("Cluster").Preload("FluentProfile").First(&node, *input.NodeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("%w: node not found", ErrInvalidArgument)
			}
			return nil, err
		}
		if node.ClusterID != nil && !clusterAllowed(*node.ClusterID, allowedClusters) {
			return nil, ErrForbidden
		}
		if node.FluentType != "" && node.FluentType != fluentType {
			findings = append(findings, models.ConfigAnalysisFinding{
				Severity:   "error",
				RuleCode:   "NODE_RUNTIME_MISMATCH",
				Message:    fmt.Sprintf("node runtime %q does not match requested fluent_type %q", node.FluentType, fluentType),
				Suggestion: "run the check against a node with the same runtime family",
				Line:       1,
			})
		}

		if node.FluentProfile == nil {
			findings = append(findings, models.ConfigAnalysisFinding{
				Severity:   "warning",
				RuleCode:   "NODE_PROFILE_MISSING",
				Message:    "node has not reported a Fluent profile yet",
				Suggestion: "wait for the agent heartbeat or refresh the node fluent profile before strict compatibility validation",
				Line:       1,
			})
		} else {
			hotReloadSupported = node.FluentProfile.SupportsHotReload
			availablePlugins := normalizePluginInventory(node.FluentProfile.LoadedPlugins)
			for _, plugin := range snapshot.Plugins {
				if !availablePlugins[strings.ToLower(plugin)] {
					missingPlugins = append(missingPlugins, plugin)
				}
			}
			if len(missingPlugins) > 0 {
				findings = append(findings, models.ConfigAnalysisFinding{
					Severity:   "error",
					RuleCode:   "PLUGIN_MISSING",
					Message:    fmt.Sprintf("node is missing plugin(s): %s", strings.Join(uniqueSorted(missingPlugins), ", ")),
					Suggestion: "install the missing plugin set or render a node-specific config variant",
					Line:       1,
				})
			}
			if strings.Contains(lowerContent, "multiline") && !node.FluentProfile.SupportsMultiline {
				findings = append(findings, models.ConfigAnalysisFinding{
					Severity:   "error",
					RuleCode:   "MULTILINE_UNSUPPORTED",
					Message:    "config uses multiline behavior but the node profile reports no multiline support",
					Suggestion: "enable multiline support on the node runtime or use a compatible parser chain",
					Line:       firstLineContaining(strings.Split(content, "\n"), "multiline"),
				})
			}
			if strings.Contains(lowerContent, "storage.") && !node.FluentProfile.SupportsStorageLayer {
				findings = append(findings, models.ConfigAnalysisFinding{
					Severity:   "warning",
					RuleCode:   "STORAGE_LAYER_UNSUPPORTED",
					Message:    "config uses storage settings but the node profile reports no storage layer support",
					Suggestion: "use memory-only buffering or upgrade the runtime capabilities",
					Line:       firstLineContaining(strings.Split(content, "\n"), "storage."),
				})
			}
			if (strings.Contains(lowerContent, "tls") || strings.Contains(lowerContent, "shared_key")) && !node.FluentProfile.SupportsForwardTLS {
				findings = append(findings, models.ConfigAnalysisFinding{
					Severity:   "warning",
					RuleCode:   "FORWARD_TLS_UNSUPPORTED",
					Message:    "config enables TLS/shared key semantics but the node profile reports no forward TLS support",
					Suggestion: "deploy to a runtime with forward TLS support or adjust the transport mode",
					Line:       firstLineContaining(strings.Split(content, "\n"), "tls"),
				})
			}
			if !hotReloadSupported {
				findings = append(findings, models.ConfigAnalysisFinding{
					Severity:   "info",
					RuleCode:   "HOT_RELOAD_UNAVAILABLE",
					Message:    "node profile reports that hot reload is unavailable",
					Suggestion: "plan a restart-based rollout for this node",
					Line:       1,
				})
			}
		}
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
		}
		return findings[i].RuleCode < findings[j].RuleCode
	})

	compatible := true
	for _, finding := range findings {
		if finding.Severity == "error" {
			compatible = false
			break
		}
	}

	return &CompatibilityCheckResult{
		FluentType:         fluentType,
		RuntimeVersion:     strings.TrimSpace(input.RuntimeVersion),
		Compatible:         compatible,
		HotReloadSupported: hotReloadSupported,
		CheckedNodeID:      input.NodeID,
		MissingPlugins:     uniqueSorted(missingPlugins),
		Findings:           findings,
	}, nil
}

func (s *fluentOpsService) RuntimeRecommendations(allowedClusters []uint) ([]RuntimeRecommendation, error) {
	var nodes []models.Node
	query := s.db.Preload("Cluster").Preload("AggregationGroup").Preload("FluentProfile").Order("hostname")
	if allowedClusters != nil {
		if len(allowedClusters) == 0 {
			return []RuntimeRecommendation{}, nil
		}
		query = query.Where("cluster_id IN ?", allowedClusters)
	}
	if err := query.Find(&nodes).Error; err != nil {
		return nil, err
	}

	nodeIDs := make([]uint, 0, len(nodes))
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}

	stateMap := map[uint]models.NodeRuntimeState{}
	if len(nodeIDs) > 0 {
		var states []models.NodeRuntimeState
		if err := s.db.Where("node_id IN ?", nodeIDs).Find(&states).Error; err != nil {
			return nil, err
		}
		for _, state := range states {
			stateMap[state.NodeID] = state
		}
	}

	recommendations := make([]RuntimeRecommendation, 0, 16)
	for _, node := range nodes {
		state := stateMap[node.ID]
		if node.NodeRole == models.NodeRoleEdgeCollector && node.AggregationGroupID == nil {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "high",
				ScopeType:  "node",
				ScopeID:    node.ID,
				Title:      "edge collector has no aggregation target",
				Detail:     fmt.Sprintf("node %s is marked as an edge collector but is not assigned to an aggregation group", node.Hostname),
				Suggestion: "bind the node to a Fluentd / Fluent Bit aggregation group or change its role",
			})
		}
		if node.FluentType == "fluentbit" && node.NodeRole == models.NodeRoleEdgeCollector && node.FluentProfile != nil && !node.FluentProfile.SupportsStorageLayer {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "medium",
				ScopeType:  "node",
				ScopeID:    node.ID,
				Title:      "edge collector lacks persistent buffering support",
				Detail:     fmt.Sprintf("node %s reports no storage layer support, which increases loss risk during downstream outages", node.Hostname),
				Suggestion: "enable filesystem buffering or move the node to a runtime profile with storage support",
			})
		}
		if state.QueueDepth >= 1000 || state.RetryCount >= 10 || state.FlushLatencyMS >= 5000 {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "high",
				ScopeType:  "node",
				ScopeID:    node.ID,
				Title:      "runtime backpressure is building up",
				Detail:     fmt.Sprintf("node %s reports queue_depth=%d retry_count=%d flush_latency_ms=%d", node.Hostname, state.QueueDepth, state.RetryCount, state.FlushLatencyMS),
				Suggestion: "scale out downstream capacity, reduce filter cost, or increase buffering before the node falls behind",
			})
		}
		if state.LastError != "" {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "medium",
				ScopeType:  "node",
				ScopeID:    node.ID,
				Title:      "recent config apply failure detected",
				Detail:     fmt.Sprintf("node %s reported a deployment error: %s", node.Hostname, state.LastError),
				Suggestion: "inspect the rendered config and rerun compatibility or replay checks before redeploying",
			})
		}
		if state.DesiredConfigHash != "" && state.EffectiveConfigHash != "" && state.DesiredConfigHash != state.EffectiveConfigHash {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "medium",
				ScopeType:  "node",
				ScopeID:    node.ID,
				Title:      "node config is drifted from desired state",
				Detail:     fmt.Sprintf("node %s effective hash %s differs from desired hash %s", node.Hostname, shortConfigHash(state.EffectiveConfigHash), shortConfigHash(state.DesiredConfigHash)),
				Suggestion: "redeploy the latest rendered config and confirm agent reload succeeded",
			})
		}
	}

	var groups []models.AggregationGroup
	groupQuery := s.db.Preload("Cluster").Order("name")
	groupQuery = applyAggregationGroupScope(groupQuery, allowedClusters)
	if err := groupQuery.Find(&groups).Error; err != nil {
		return nil, err
	}
	groupMetrics, err := s.aggregationGroupMetricsForGroups(groups)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		metric := groupMetrics[group.ID]
		if metric == nil {
			continue
		}
		if metric.AssignedNodes >= 10 && metric.AvgCPU >= 75 {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "high",
				ScopeType:  "aggregation_group",
				ScopeID:    group.ID,
				Title:      "aggregation group is approaching saturation",
				Detail:     fmt.Sprintf("group %s has %d assigned nodes and average CPU %.1f%%", metric.Name, metric.AssignedNodes, metric.AvgCPU),
				Suggestion: "rebalance edge collectors, add aggregators, or split high-volume pipelines",
			})
		}
		if metric.AssignedNodes > 0 && metric.OnlineNodes < metric.AssignedNodes {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "medium",
				ScopeType:  "aggregation_group",
				ScopeID:    group.ID,
				Title:      "aggregation group has offline members",
				Detail:     fmt.Sprintf("group %s has %d/%d nodes online", metric.Name, metric.OnlineNodes, metric.AssignedNodes),
				Suggestion: "check node reachability and failover coverage before peak traffic periods",
			})
		}
	}

	pipelines, err := s.ListPipelines(allowedClusters)
	if err != nil {
		return nil, err
	}
	for _, pipeline := range pipelines {
		if pipeline.Protocol == "forward" && pipeline.DestinationAggregationGroup != nil && !pipeline.DestinationAggregationGroup.EnableTLS {
			recommendations = append(recommendations, RuntimeRecommendation{
				Severity:   "medium",
				ScopeType:  "pipeline",
				ScopeID:    pipeline.ID,
				Title:      "forward pipeline is not protected by TLS",
				Detail:     fmt.Sprintf("pipeline %s forwards to aggregation group %s without TLS enabled", pipeline.Name, firstNonEmpty(pipeline.DestinationAggregationGroup.Alias, pipeline.DestinationAggregationGroup.Name)),
				Suggestion: "enable TLS on the destination aggregation group and rotate a protected shared key",
			})
		}
	}

	sort.SliceStable(recommendations, func(i, j int) bool {
		if recommendationRank(recommendations[i].Severity) != recommendationRank(recommendations[j].Severity) {
			return recommendationRank(recommendations[i].Severity) < recommendationRank(recommendations[j].Severity)
		}
		if recommendations[i].ScopeType != recommendations[j].ScopeType {
			return recommendations[i].ScopeType < recommendations[j].ScopeType
		}
		return recommendations[i].Title < recommendations[j].Title
	})

	return recommendations, nil
}

func analyzeConfigSemantics(fluentType, content string) semanticSnapshot {
	if fluentType == "fluentbit" {
		return analyzeFluentBitSemantics(content)
	}
	return analyzeFluentdSemantics(content)
}

func analyzeFluentBitSemantics(content string) semanticSnapshot {
	blocks := parseFluentBitBlocks(content)
	snapshot := semanticSnapshot{}

	for _, block := range blocks {
		pluginType := firstNonEmpty(block.Fields["name"], block.Fields["@type"])
		name := firstNonEmpty(block.Fields["alias"], block.Fields["id"], pluginType)
		match := firstNonEmpty(block.Fields["match"], "*")
		switch block.Kind {
		case "INPUT":
			if pluginType != "" {
				snapshot.Inputs = append(snapshot.Inputs, pluginType)
				snapshot.Plugins = append(snapshot.Plugins, pluginType)
			}
			if parser := block.Fields["parser"]; parser != "" {
				snapshot.Parsers = append(snapshot.Parsers, parser)
			}
		case "PARSER":
			if parserName := firstNonEmpty(block.Fields["name"], block.Fields["format"]); parserName != "" {
				snapshot.Parsers = append(snapshot.Parsers, parserName)
				snapshot.Plugins = append(snapshot.Plugins, parserName)
			}
		case "FILTER":
			if pluginType != "" {
				snapshot.Filters = append(snapshot.Filters, pluginType)
				snapshot.Plugins = append(snapshot.Plugins, pluginType)
				snapshot.FilterRules = append(snapshot.FilterRules, matchRule{Pattern: match, Name: name, Type: pluginType})
			}
		case "OUTPUT":
			if pluginType != "" {
				snapshot.Outputs = append(snapshot.Outputs, pluginType)
				snapshot.Plugins = append(snapshot.Plugins, pluginType)
				snapshot.Routes = append(snapshot.Routes, fmt.Sprintf("%s -> %s", match, name))
				snapshot.OutputRules = append(snapshot.OutputRules, matchRule{Pattern: match, Name: name, Type: pluginType})
			}
		}
	}

	snapshot.Inputs = uniqueSorted(snapshot.Inputs)
	snapshot.Parsers = uniqueSorted(snapshot.Parsers)
	snapshot.Filters = uniqueSorted(snapshot.Filters)
	snapshot.Routes = uniqueSorted(snapshot.Routes)
	snapshot.Outputs = uniqueSorted(snapshot.Outputs)
	snapshot.Plugins = uniqueSorted(snapshot.Plugins)
	return snapshot
}

func analyzeFluentdSemantics(content string) semanticSnapshot {
	blocks := parseFluentdBlocks(content)
	snapshot := semanticSnapshot{}

	for _, block := range blocks {
		switch block.Kind {
		case "source":
			if block.Plugin != "" {
				snapshot.Inputs = append(snapshot.Inputs, block.Plugin)
				snapshot.Plugins = append(snapshot.Plugins, block.Plugin)
			}
		case "parse":
			if block.Plugin != "" {
				snapshot.Parsers = append(snapshot.Parsers, block.Plugin)
				snapshot.Plugins = append(snapshot.Plugins, block.Plugin)
			}
		case "filter":
			if block.Plugin != "" {
				snapshot.Filters = append(snapshot.Filters, block.Plugin)
				snapshot.Plugins = append(snapshot.Plugins, block.Plugin)
				snapshot.FilterRules = append(snapshot.FilterRules, matchRule{
					Pattern: firstNonEmpty(block.Match, "**"),
					Name:    firstNonEmpty(block.ID, block.Plugin),
					Type:    block.Plugin,
				})
			}
		case "match":
			if block.Plugin != "" {
				pattern := firstNonEmpty(block.Match, "**")
				targetName := firstNonEmpty(block.ID, block.Plugin)
				snapshot.Outputs = append(snapshot.Outputs, block.Plugin)
				snapshot.Plugins = append(snapshot.Plugins, block.Plugin)
				snapshot.Routes = append(snapshot.Routes, fmt.Sprintf("%s -> %s", pattern, targetName))
				snapshot.OutputRules = append(snapshot.OutputRules, matchRule{Pattern: pattern, Name: targetName, Type: block.Plugin})
			}
		}
	}

	snapshot.Inputs = uniqueSorted(snapshot.Inputs)
	snapshot.Parsers = uniqueSorted(snapshot.Parsers)
	snapshot.Filters = uniqueSorted(snapshot.Filters)
	snapshot.Routes = uniqueSorted(snapshot.Routes)
	snapshot.Outputs = uniqueSorted(snapshot.Outputs)
	snapshot.Plugins = uniqueSorted(snapshot.Plugins)
	return snapshot
}

type fluentBitBlock struct {
	Kind   string
	Fields map[string]string
}

func parseFluentBitBlocks(content string) []fluentBitBlock {
	sectionRe := regexp.MustCompile(`^\[([A-Za-z_]+)\]$`)
	blocks := []fluentBitBlock{}
	var current *fluentBitBlock

	flush := func() {
		if current != nil {
			blocks = append(blocks, *current)
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if matches := sectionRe.FindStringSubmatch(line); len(matches) == 2 {
			flush()
			current = &fluentBitBlock{
				Kind:   strings.ToUpper(matches[1]),
				Fields: map[string]string{},
			}
			continue
		}
		if current == nil {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(parts[0])
		value := strings.TrimSpace(strings.TrimPrefix(line, parts[0]))
		current.Fields[key] = value
	}
	flush()
	return blocks
}

type fluentdBlock struct {
	Kind   string
	Match  string
	Plugin string
	ID     string
}

func parseFluentdBlocks(content string) []fluentdBlock {
	openRe := regexp.MustCompile(`^<([a-zA-Z_]+)(?:\s+([^>]+))?>$`)
	closeRe := regexp.MustCompile(`^</([a-zA-Z_]+)>$`)
	blocks := []fluentdBlock{}
	stack := []fluentdBlock{}

	appendCurrent := func(block fluentdBlock) {
		if block.Kind != "" {
			blocks = append(blocks, block)
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if matches := openRe.FindStringSubmatch(line); len(matches) >= 2 {
			stack = append(stack, fluentdBlock{
				Kind:  strings.ToLower(matches[1]),
				Match: strings.TrimSpace(matches[2]),
			})
			continue
		}
		if matches := closeRe.FindStringSubmatch(line); len(matches) == 2 {
			if len(stack) == 0 {
				continue
			}
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			appendCurrent(last)
			continue
		}
		if len(stack) == 0 {
			continue
		}
		if strings.HasPrefix(line, "@type ") {
			stack[len(stack)-1].Plugin = strings.TrimSpace(strings.TrimPrefix(line, "@type "))
			continue
		}
		if strings.HasPrefix(line, "@id ") {
			stack[len(stack)-1].ID = strings.TrimSpace(strings.TrimPrefix(line, "@id "))
		}
	}
	for i := len(stack) - 1; i >= 0; i-- {
		appendCurrent(stack[i])
	}
	return blocks
}

func parseReplaySample(sampleLog string, knownParsers []string) (map[string]interface{}, string) {
	record := map[string]interface{}{}
	if err := json.Unmarshal([]byte(sampleLog), &record); err == nil {
		record["__raw"] = sampleLog
		return record, "json"
	}
	record["message"] = sampleLog
	parser := "raw"
	if len(knownParsers) > 0 {
		parser = knownParsers[0]
	}
	return record, parser
}

func diffSemanticCategory(category string, before, after []string) []SemanticChange {
	beforeSet := map[string]bool{}
	afterSet := map[string]bool{}
	for _, item := range before {
		beforeSet[item] = true
	}
	for _, item := range after {
		afterSet[item] = true
	}

	changes := []SemanticChange{}
	for _, item := range uniqueSorted(after) {
		if !beforeSet[item] {
			changes = append(changes, SemanticChange{
				Category:   category,
				ChangeType: "added",
				Item:       item,
				Detail:     fmt.Sprintf("%s %q appears only in the new config", category, item),
			})
		}
	}
	for _, item := range uniqueSorted(before) {
		if !afterSet[item] {
			changes = append(changes, SemanticChange{
				Category:   category,
				ChangeType: "removed",
				Item:       item,
				Detail:     fmt.Sprintf("%s %q is no longer present in the new config", category, item),
			})
		}
	}
	return changes
}

func normalizePluginInventory(raw string) map[string]bool {
	inventory := map[string]bool{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return inventory
	}

	var arr []string
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			for _, item := range arr {
				name := strings.ToLower(strings.TrimSpace(item))
				if name != "" {
					inventory[name] = true
				}
			}
			return inventory
		}
	}

	splitter := func(r rune) bool {
		return r == ',' || r == '\n' || r == '\t' || r == ' '
	}
	for _, item := range strings.FieldsFunc(raw, splitter) {
		name := strings.ToLower(strings.TrimSpace(item))
		if name != "" {
			inventory[name] = true
		}
	}
	return inventory
}

func uniqueSorted(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || set[trimmed] {
			continue
		}
		set[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func matchTagPattern(pattern, tag string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" || pattern == "**" {
		return true
	}
	for _, candidate := range splitMatchPatterns(pattern) {
		if wildcardMatch(candidate, tag) {
			return true
		}
	}
	return false
}

func splitMatchPatterns(pattern string) []string {
	parts := strings.FieldsFunc(pattern, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	if len(parts) == 0 {
		return []string{pattern}
	}
	return parts
}

func wildcardMatch(pattern, value string) bool {
	if pattern == "" {
		return true
	}
	if cached, ok := wildcardRegexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp).MatchString(value)
	}
	quoted := regexp.QuoteMeta(pattern)
	quoted = strings.ReplaceAll(quoted, `\*\*`, ".*")
	quoted = strings.ReplaceAll(quoted, `\*`, `[^.]+`)
	re, err := regexp.Compile("^" + quoted + "$")
	if err != nil {
		return false
	}
	wildcardRegexCache.Store(pattern, re)
	return re.MatchString(value)
}

func shortConfigHash(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func recommendationRank(severity string) int {
	switch severity {
	case "high":
		return 0
	case "medium":
		return 1
	default:
		return 2
	}
}
