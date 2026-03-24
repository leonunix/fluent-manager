package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
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
var importNameTokenRegex = regexp.MustCompile(`[^a-z0-9]+`)
var importVariableTokenRegex = regexp.MustCompile(`[^a-z0-9_]+`)
var importExtractableKeySet = map[string]bool{
	"flush":               true,
	"log_level":           true,
	"workers":             true,
	"path":                true,
	"tag":                 true,
	"db":                  true,
	"db_path":             true,
	"db.sync":             true,
	"db_sync":             true,
	"pos_file":            true,
	"parser":              true,
	"match":               true,
	"host":                true,
	"port":                true,
	"uri":                 true,
	"endpoint":            true,
	"index":               true,
	"index_name":          true,
	"brokers":             true,
	"topic":               true,
	"topics":              true,
	"tenant_id":           true,
	"labels":              true,
	"http_user":           true,
	"http_passwd":         true,
	"http_password":       true,
	"user":                true,
	"password":            true,
	"scheme":              true,
	"tls":                 true,
	"tls.verify":          true,
	"tls_verify":          true,
	"ssl_verify":          true,
	"logstash_format":     true,
	"logstash_prefix":     true,
	"logstash_dateformat": true,
	"generate_id":         true,
	"retry_limit":         true,
	"replace_dots":        true,
	"suppress_type_name":  true,
	"trace_error":         true,
	"http_method":         true,
	"serializer":          true,
	"default_topic":       true,
	"output_data_type":    true,
	"format":              true,
	"time_key":            true,
	"time_format":         true,
}

func (s *fluentOpsService) ImportExistingConfig(input *ConfigImportInput) (*ConfigImportResult, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: import payload is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validRenderedConfigTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}

	prefix := sanitizeImportNameToken(input.NamePrefix)
	if prefix == "" {
		prefix = "imported-config"
	}

	modules, warnings := importConfigModules(fluentType, content, prefix)
	if len(modules) == 0 {
		return nil, fmt.Errorf("%w: no supported module blocks were detected", ErrInvalidArgument)
	}
	modules, matchedExisting, reusedExisting := s.attachImportReuseSuggestions(modules)
	modules, destinations := s.attachImportDestinationSuggestions(fluentType, modules)
	importMode, autoAssembleSupported, workspaceOnlyReason := classifyImportedConfigMode(modules)
	suggestedTemplateName := prefix + "-assembly"
	if !autoAssembleSupported {
		suggestedTemplateName = prefix + "-assets"
	}

	flowPath := buildImportedFlowPath(modules, destinations)
	templateDraftContent := buildImportedTemplateDraft(modules)
	semanticDiff, equivalent := buildImportSemanticValidation(fluentType, content, templateDraftContent)
	lintFindings := lintConfigContent(fluentType, templateDraftContent)
	lintSummary := buildLintSummary(lintFindings)
	verdict := "needs_review"
	switch {
	case equivalent:
		verdict = "equivalent"
	case len(semanticDiff.Changes) <= 2:
		verdict = "mostly_equivalent"
	}
	validationSummary := fmt.Sprintf("%s; lint: %s", semanticDiff.Summary, lintSummary)
	flowLayout := map[string]interface{}{
		"builder":                 "config_import",
		"import_mode":             importMode,
		"auto_assemble_supported": autoAssembleSupported,
		"workspace_only_reason":   workspaceOnlyReason,
		"name_prefix":             prefix,
		"path":                    flowPath,
		"suggested_template":      suggestedTemplateName,
		"module_count":            len(modules),
		"matched_existing_count":  matchedExisting,
		"reused_existing_count":   reusedExisting,
		"destination_count":       len(destinations),
		"destinations":            destinations,
		"retained_warnings":       uniqueSorted(warnings),
		"validation_verdict":      verdict,
	}

	return &ConfigImportResult{
		FluentType:            fluentType,
		NamePrefix:            prefix,
		ImportMode:            importMode,
		AutoAssembleSupported: autoAssembleSupported,
		WorkspaceOnlyReason:   workspaceOnlyReason,
		Summary:               buildImportedConfigSummary(fluentType, modules, reusedExisting, autoAssembleSupported),
		SuggestedTemplateName: suggestedTemplateName,
		Warnings:              uniqueSorted(warnings),
		Modules:               modules,
		Destinations:          destinations,
		FlowPath:              flowPath,
		FlowLayout:            flowLayout,
		TemplateDraftContent:  templateDraftContent,
		Validation: ConfigImportValidation{
			Equivalent:      equivalent,
			Verdict:         verdict,
			Summary:         validationSummary,
			SemanticDiff:    semanticDiff,
			LintSummary:     lintSummary,
			LintFindings:    lintFindings,
			RenderedContent: templateDraftContent,
		},
	}, nil
}

func classifyImportedConfigMode(modules []ImportedConfigModule) (string, bool, string) {
	hasPipelineStage := false
	for _, module := range modules {
		switch module.ModuleType {
		case "input", "filter", "route", "output":
			hasPipelineStage = true
		}
	}
	if hasPipelineStage {
		return "existing_config", true, ""
	}
	return "workspace_assets", false, "imported content contains only global assets such as service/parser blocks, so it will be stored in the module workspace without auto-assembling a runnable config template"
}

func buildImportedConfigSummary(fluentType string, modules []ImportedConfigModule, reusedExisting int, autoAssembleSupported bool) string {
	if autoAssembleSupported {
		return fmt.Sprintf("extracted %d module draft(s) from the imported %s config, with %d reusable match(es)", len(modules), fluentType, reusedExisting)
	}
	return fmt.Sprintf("extracted %d reusable global asset module(s) from the imported %s config; these assets will be stored in the workspace without auto-assembling a pipeline template", len(modules), fluentType)
}

func (s *fluentOpsService) attachImportReuseSuggestions(modules []ImportedConfigModule) ([]ImportedConfigModule, int, int) {
	if len(modules) == 0 {
		return modules, 0, 0
	}

	var existing []models.ConfigModule
	if err := s.db.Where("fluent_type IN ?", []string{"shared", modules[0].FluentType}).Find(&existing).Error; err != nil {
		return modules, 0, 0
	}

	matched, reused := 0, 0
	for index := range modules {
		if modules[index].ModuleType == "output" {
			modules[index].ImportAction = "create_new"
			continue
		}
		match := findReusableImportedModule(modules[index], existing)
		if match == nil {
			modules[index].ImportAction = "create_new"
			continue
		}
		matched++
		modules[index].ExistingModuleID = &match.ID
		modules[index].ExistingModuleName = match.Name
		if normalizeImportContent(match.Content) == normalizeImportContent(modules[index].Content) &&
			normalizeImportVariables(match.Variables) == normalizeImportVariables(modules[index].Variables) {
			modules[index].ImportAction = "reuse_existing"
			reused++
		} else {
			modules[index].ImportAction = "create_new"
		}
	}
	return modules, matched, reused
}

func (s *fluentOpsService) attachImportDestinationSuggestions(fluentType string, modules []ImportedConfigModule) ([]ImportedConfigModule, []ImportedConfigDestination) {
	if len(modules) == 0 {
		return modules, nil
	}

	var targets []models.OutputTarget
	if err := s.db.Where("fluent_type IN ?", []string{"shared", fluentType}).Order("target_type, name").Find(&targets).Error; err != nil {
		return modules, nil
	}
	if len(targets) == 0 {
		return modules, nil
	}

	destinations := make([]ImportedConfigDestination, 0, len(modules))
	for index := range modules {
		module := &modules[index]
		if module.ModuleType != "output" {
			continue
		}
		targetType := inferImportedOutputTargetType(*module)
		if targetType == "" {
			continue
		}
		match := bestOutputTargetMatch(*module, targetType, targets)
		if match == nil {
			continue
		}
		module.OutputTargetID = &match.target.ID
		module.OutputTargetName = match.target.Name
		module.OutputTargetType = match.target.TargetType
		module.OutputTargetEndpoint = match.target.Endpoint
		module.OutputTargetMatchType = match.matchType
		destinations = append(destinations, ImportedConfigDestination{
			OutputModuleName:  module.Name,
			OutputModuleOrder: module.Order,
			OutputTargetID:    match.target.ID,
			Name:              match.target.Name,
			TargetType:        match.target.TargetType,
			Endpoint:          match.target.Endpoint,
			MatchType:         match.matchType,
		})
	}

	return modules, uniqueImportedDestinations(destinations)
}

type outputTargetMatch struct {
	target    models.OutputTarget
	matchType string
	score     int
}

func bestOutputTargetMatch(module ImportedConfigModule, targetType string, targets []models.OutputTarget) *outputTargetMatch {
	candidates := make([]models.OutputTarget, 0, len(targets))
	for _, target := range targets {
		if strings.EqualFold(strings.TrimSpace(target.TargetType), targetType) {
			candidates = append(candidates, target)
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	best := &outputTargetMatch{}
	for _, candidate := range candidates {
		score := scoreImportedOutputTargetMatch(module, candidate)
		if score > best.score {
			best = &outputTargetMatch{
				target:    candidate,
				matchType: "exact",
				score:     score,
			}
		}
	}
	if best.score > 0 {
		return best
	}
	if len(candidates) == 1 {
		return &outputTargetMatch{
			target:    candidates[0],
			matchType: "type_match",
			score:     0,
		}
	}
	return nil
}

func scoreImportedOutputTargetMatch(module ImportedConfigModule, target models.OutputTarget) int {
	moduleVars := parseImportVariablesMap(module.Variables)
	targetSettings := parseOutputTargetSettingsMap(target.Settings)
	score := 0

	moduleHost := normalizedImportValue(firstNonEmpty(importMapString(moduleVars, "host"), importMapString(moduleVars, "endpoint")))
	targetHost := normalizedImportValue(firstNonEmpty(importMapString(targetSettings, "host"), importMapString(targetSettings, "endpoint")))
	if moduleHost != "" && targetHost != "" {
		if moduleHost == targetHost {
			score += 4
		} else if strings.Contains(targetHost, moduleHost) || strings.Contains(moduleHost, targetHost) {
			score += 2
		}
	}

	modulePort := normalizedImportValue(importMapString(moduleVars, "port"))
	targetPort := normalizedImportValue(importMapString(targetSettings, "port"))
	if modulePort != "" && targetPort != "" && modulePort == targetPort {
		score += 2
	}

	moduleIndex := normalizedImportValue(importMapString(moduleVars, "index"))
	targetIndex := normalizedImportValue(importMapString(targetSettings, "index"))
	if moduleIndex != "" && targetIndex != "" && moduleIndex == targetIndex {
		score += 2
	}

	moduleURI := normalizedImportValue(importMapString(moduleVars, "uri"))
	targetURI := normalizedImportValue(firstNonEmpty(importMapString(targetSettings, "uri"), target.Endpoint))
	if moduleURI != "" && targetURI != "" {
		if moduleURI == targetURI {
			score += 3
		} else if strings.Contains(targetURI, moduleURI) || strings.Contains(moduleURI, targetURI) {
			score += 1
		}
	}

	moduleTopic := normalizedImportValue(firstNonEmpty(importMapString(moduleVars, "topics"), importMapString(moduleVars, "topic")))
	targetTopic := normalizedImportValue(firstNonEmpty(importMapString(targetSettings, "topics"), importMapString(targetSettings, "topic")))
	if moduleTopic != "" && targetTopic != "" && moduleTopic == targetTopic {
		score += 2
	}

	moduleBucket := normalizedImportValue(importMapString(moduleVars, "bucket"))
	targetBucket := normalizedImportValue(importMapString(targetSettings, "bucket"))
	if moduleBucket != "" && targetBucket != "" && moduleBucket == targetBucket {
		score += 2
	}

	modulePath := normalizedImportValue(importMapString(moduleVars, "path"))
	targetPath := normalizedImportValue(importMapString(targetSettings, "path"))
	if modulePath != "" && targetPath != "" && modulePath == targetPath {
		score += 1
	}

	if strings.EqualFold(strings.TrimSpace(target.TargetType), "stdout") {
		score++
	}
	return score
}

func inferImportedOutputTargetType(module ImportedConfigModule) string {
	plugin := strings.ToLower(strings.TrimSpace(module.DetectedPlugin))
	switch plugin {
	case "es", "opensearch", "elasticsearch":
		return "opensearch"
	case "loki":
		return "loki"
	case "kafka", "rdkafka":
		return "kafka"
	case "http":
		return "http"
	case "s3":
		return "s3"
	case "stdout":
		return "stdout"
	}

	content := strings.ToLower(strings.TrimSpace(module.Content))
	switch {
	case strings.Contains(content, "opensearch"), strings.Contains(content, "elasticsearch"):
		return "opensearch"
	case strings.Contains(content, "loki"):
		return "loki"
	case strings.Contains(content, "kafka"):
		return "kafka"
	case strings.Contains(content, "stdout"):
		return "stdout"
	case strings.Contains(content, "s3"):
		return "s3"
	case strings.Contains(content, "http"):
		return "http"
	default:
		return ""
	}
}

func parseImportVariablesMap(raw string) map[string]interface{} {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return map[string]interface{}{}
	}
	return parsed
}

func parseOutputTargetSettingsMap(raw string) map[string]interface{} {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return map[string]interface{}{}
	}
	return parsed
}

func importMapString(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}

func normalizedImportValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func uniqueImportedDestinations(destinations []ImportedConfigDestination) []ImportedConfigDestination {
	if len(destinations) == 0 {
		return nil
	}
	seen := map[uint]bool{}
	filtered := make([]ImportedConfigDestination, 0, len(destinations))
	for _, destination := range destinations {
		if seen[destination.OutputTargetID] {
			continue
		}
		seen[destination.OutputTargetID] = true
		filtered = append(filtered, destination)
	}
	return filtered
}

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

func importConfigModules(fluentType, content, prefix string) ([]ImportedConfigModule, []string) {
	if fluentType == "fluentbit" {
		return importFluentBitModules(content, prefix)
	}
	return importFluentdModules(content, prefix)
}

func importFluentBitModules(content, prefix string) ([]ImportedConfigModule, []string) {
	blocks := parseFluentBitBlocks(content)
	modules := make([]ImportedConfigModule, 0, len(blocks))
	warnings := []string{}

	for _, block := range blocks {
		moduleType := mapFluentBitBlockToModuleType(block.Kind)
		if moduleType == "" {
			warnings = append(warnings, fmt.Sprintf("block [%s] is not imported in the baseline importer", block.Kind))
			continue
		}
		plugin := firstNonEmpty(block.Fields["name"], block.Fields["@type"], block.Fields["format"], strings.ToLower(block.Kind))
		order := len(modules) + 1
		extractedContent, extractedVariables, variableKeys := extractImportVariables(block.Raw, "fluentbit")
		modules = append(modules, ImportedConfigModule{
			Order:          order,
			Name:           buildImportedModuleName(prefix, moduleType, plugin, order),
			Summary:        buildImportedModuleSummary(moduleType, plugin, block.Fields["match"]),
			ModuleType:     moduleType,
			FluentType:     "fluentbit",
			DetectedPlugin: plugin,
			Content:        extractedContent,
			Variables:      extractedVariables,
			VariableKeys:   variableKeys,
			ImportAction:   "create_new",
		})
	}

	for _, module := range modules {
		if module.ModuleType == "output" {
			warnings = append(warnings, "routing conditions remain inside output modules during import; dedicated route modules are not split automatically yet")
			break
		}
	}

	return modules, warnings
}

func importFluentdModules(content, prefix string) ([]ImportedConfigModule, []string) {
	blocks := parseFluentdBlocks(content)
	modules := make([]ImportedConfigModule, 0, len(blocks))
	warnings := []string{}

	for _, block := range blocks {
		if shouldCreateFluentdRouteCompanion(block) {
			routeContent, routeVariables, routeVariableKeys := buildFluentdRouteCompanion(block)
			modules = append(modules, ImportedConfigModule{
				Order:          len(modules) + 1,
				Name:           buildImportedModuleName(prefix, "route", firstNonEmpty(block.Match, block.Plugin, "route"), len(modules)+1),
				Summary:        buildImportedModuleSummary("route", block.Plugin, block.Match),
				ModuleType:     "route",
				FluentType:     "fluentd",
				DetectedPlugin: block.Plugin,
				Content:        routeContent,
				Variables:      routeVariables,
				VariableKeys:   routeVariableKeys,
				ImportAction:   "create_new",
			})
		}

		moduleType, warning := mapFluentdBlockToModuleType(block)
		if warning != "" {
			warnings = append(warnings, warning)
		}
		if moduleType == "" {
			continue
		}

		plugin := firstNonEmpty(block.Plugin, block.ID, block.Match, block.Kind)
		order := len(modules) + 1
		extractedContent, extractedVariables, variableKeys := extractImportVariables(block.Raw, "fluentd")
		modules = append(modules, ImportedConfigModule{
			Order:          order,
			Name:           buildImportedModuleName(prefix, moduleType, plugin, order),
			Summary:        buildImportedModuleSummary(moduleType, plugin, block.Match),
			ModuleType:     moduleType,
			FluentType:     "fluentd",
			DetectedPlugin: plugin,
			Content:        extractedContent,
			Variables:      extractedVariables,
			VariableKeys:   variableKeys,
			ImportAction:   "create_new",
		})
	}

	return modules, warnings
}

func shouldCreateFluentdRouteCompanion(block fluentdBlock) bool {
	return block.Kind == "match" && strings.TrimSpace(block.Match) != ""
}

func buildFluentdRouteCompanion(block fluentdBlock) (string, string, []string) {
	variables := map[string]interface{}{
		"match": strings.TrimSpace(block.Match),
	}
	keys := []string{"match"}
	if target := strings.TrimSpace(firstNonEmpty(block.ID, block.Plugin)); target != "" {
		variables["target"] = target
		keys = append(keys, "target")
	}
	content := "# Imported route intent\n# Match {{ .match }}"
	if _, ok := variables["target"]; ok {
		content += " -> {{ .target }}"
	}
	serialized, _ := json.MarshalIndent(variables, "", "  ")
	return content, string(serialized), keys
}

func mapFluentBitBlockToModuleType(kind string) string {
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case "SERVICE":
		return "service"
	case "INPUT":
		return "input"
	case "PARSER":
		return "parser"
	case "FILTER":
		return "filter"
	case "OUTPUT":
		return "output"
	default:
		return ""
	}
}

func mapFluentdBlockToModuleType(block fluentdBlock) (string, string) {
	switch block.Kind {
	case "system":
		return "service", ""
	case "source":
		return "input", ""
	case "filter":
		return "filter", ""
	case "match":
		return "output", "routing and output behavior remain coupled inside <match> blocks during baseline import"
	case "parse":
		if block.ParentKind != "" {
			return "", fmt.Sprintf("nested <%s> parse content was retained inside the parent %s block for safe migration", block.Kind, block.ParentKind)
		}
		return "parser", ""
	case "label":
		return "", "label sections are not split into reusable modules by the baseline importer yet"
	default:
		return "", fmt.Sprintf("block <%s> is not imported in the baseline importer", block.Kind)
	}
}

func buildImportedModuleName(prefix, moduleType, plugin string, order int) string {
	if moduleType == "parser" {
		if parserName := sanitizeImportNameToken(plugin); parserName != "" {
			return parserName
		}
	}
	base := []string{sanitizeImportNameToken(prefix), sanitizeImportNameToken(plugin), sanitizeImportNameToken(moduleType)}
	tokens := make([]string, 0, len(base))
	for _, token := range base {
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	if len(tokens) == 0 {
		tokens = append(tokens, "imported-module")
	}
	return fmt.Sprintf("%s-%02d", strings.Join(tokens, "-"), order)
}

func buildImportedModuleSummary(moduleType, plugin, match string) string {
	plugin = strings.TrimSpace(plugin)
	match = strings.TrimSpace(match)
	switch moduleType {
	case "service":
		if plugin != "" {
			return fmt.Sprintf("service tuning around %s", plugin)
		}
		return "service-level runtime settings"
	case "input":
		if plugin != "" {
			return fmt.Sprintf("input source using %s", plugin)
		}
		return "input source block"
	case "parser":
		if plugin != "" {
			return fmt.Sprintf("parser definition using %s", plugin)
		}
		return "parser definition"
	case "filter":
		if plugin != "" && match != "" {
			return fmt.Sprintf("filter %s for %s", plugin, match)
		}
		if plugin != "" {
			return fmt.Sprintf("filter stage using %s", plugin)
		}
		return "filter stage"
	case "route":
		if plugin != "" && match != "" {
			return fmt.Sprintf("route %s toward %s", match, plugin)
		}
		if match != "" {
			return fmt.Sprintf("route stage for %s", match)
		}
		return "route stage"
	case "output":
		if plugin != "" && match != "" {
			return fmt.Sprintf("output %s for %s", plugin, match)
		}
		if plugin != "" {
			return fmt.Sprintf("output stage using %s", plugin)
		}
		return "output stage"
	default:
		return "imported module"
	}
}

func sanitizeImportNameToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = importNameTokenRegex.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func buildImportedFlowPath(modules []ImportedConfigModule, destinations []ImportedConfigDestination) []string {
	grouped := map[string][]string{
		"service": {},
		"input":   {},
		"parser":  {},
		"filter":  {},
		"route":   {},
		"output":  {},
	}
	for _, module := range modules {
		grouped[module.ModuleType] = append(grouped[module.ModuleType], module.Name)
	}

	path := []string{}
	if len(grouped["service"]) > 0 {
		path = append(path, summarizeImportStage(grouped["service"]))
	}
	if len(grouped["input"]) > 0 {
		path = append(path, summarizeImportStage(grouped["input"]))
	}

	processors := append([]string{}, grouped["parser"]...)
	processors = append(processors, grouped["filter"]...)
	processors = append(processors, grouped["route"]...)
	if len(processors) > 0 {
		path = append(path, summarizeImportStage(processors))
	}
	if len(grouped["output"]) > 0 {
		path = append(path, summarizeImportStage(grouped["output"]))
	}
	if len(destinations) > 0 {
		destinationNames := make([]string, 0, len(destinations))
		for _, destination := range destinations {
			destinationNames = append(destinationNames, destination.Name)
		}
		path = append(path, summarizeImportStage(destinationNames))
	}
	return path
}

func summarizeImportStage(items []string) string {
	items = uniqueSorted(items)
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " + " + items[1]
	default:
		return items[0] + " + " + items[1] + " + ..."
	}
}

func findReusableImportedModule(module ImportedConfigModule, existing []models.ConfigModule) *models.ConfigModule {
	for index := range existing {
		current := existing[index]
		if current.ModuleType != module.ModuleType {
			continue
		}
		if current.FluentType != module.FluentType && current.FluentType != "shared" {
			continue
		}
		if normalizeImportContent(current.Content) == normalizeImportContent(module.Content) {
			return &current
		}
	}
	for index := range existing {
		current := existing[index]
		if current.ModuleType == module.ModuleType &&
			(current.FluentType == module.FluentType || current.FluentType == "shared") &&
			sanitizeImportNameToken(current.Name) == sanitizeImportNameToken(module.Name) {
			return &current
		}
	}
	return nil
}

func normalizeImportContent(content string) string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return strings.Join(normalized, "\n")
}

func normalizeImportVariables(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "{}"
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return value
	}
	normalized, _ := json.Marshal(parsed)
	return string(normalized)
}

func buildImportedTemplateDraft(modules []ImportedConfigModule) string {
	ordered := append([]ImportedConfigModule{}, modules...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Order < ordered[j].Order
	})

	sections := make([]string, 0, len(ordered))
	for _, module := range ordered {
		sections = append(sections, strings.TrimSpace(renderImportedModuleWithDefaults(module)))
	}
	return strings.TrimSpace(strings.Join(sections, "\n\n"))
}

func buildImportSemanticValidation(fluentType, originalContent, renderedContent string) (*ConfigSemanticDiffResult, bool) {
	diff := &ConfigSemanticDiffResult{
		FluentType: fluentType,
		Summary:    "validation unavailable",
		Changes:    []SemanticChange{},
	}
	if strings.TrimSpace(renderedContent) == "" {
		diff.Summary = "rendered content is empty"
		return diff, false
	}

	before := analyzeConfigSemantics(fluentType, strings.TrimSpace(originalContent))
	after := analyzeConfigSemantics(fluentType, strings.TrimSpace(renderedContent))
	changes := make([]SemanticChange, 0, 12)
	changes = append(changes, diffSemanticCategory("input", before.Inputs, after.Inputs)...)
	changes = append(changes, diffSemanticCategory("parser", before.Parsers, after.Parsers)...)
	changes = append(changes, diffSemanticCategory("filter", before.Filters, after.Filters)...)
	changes = append(changes, diffSemanticCategory("route", before.Routes, after.Routes)...)
	changes = append(changes, diffSemanticCategory("output", before.Outputs, after.Outputs)...)
	changes = append(changes, diffSemanticCategory("plugin", before.Plugins, after.Plugins)...)

	diff.Changes = changes
	if len(changes) == 0 {
		diff.Summary = "no semantic changes detected between imported assembly and original config"
		return diff, true
	}

	added, removed := 0, 0
	for _, change := range changes {
		if change.ChangeType == "added" {
			added++
		}
		if change.ChangeType == "removed" {
			removed++
		}
	}
	diff.Summary = fmt.Sprintf("%d semantic change(s): %d added, %d removed", len(changes), added, removed)
	return diff, false
}

func renderImportedModuleWithDefaults(module ImportedConfigModule) string {
	content := strings.TrimSpace(module.Content)
	variables := map[string]interface{}{}
	if parsed, err := parseRenderVariables(module.Variables); err == nil {
		variables = parsed
	}
	rendered, err := renderModuleTemplate(content, variables)
	if err != nil {
		return content
	}
	return rendered
}

func extractImportVariables(raw, runtime string) (string, string, []string) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 {
		return strings.TrimSpace(raw), "{}", nil
	}

	variableMap := map[string]interface{}{}
	variableKeys := []string{}
	usedKeys := map[string]int{}
	transformed := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") ||
			strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "<") || strings.HasPrefix(trimmed, "</") {
			transformed = append(transformed, line)
			continue
		}

		key, rawValue, indent, ok := parseImportLine(line, runtime)
		if !ok || !shouldExtractImportKey(key) {
			transformed = append(transformed, line)
			continue
		}

		variableKey := uniqueImportVariableKey(key, usedKeys)
		variableKeys = append(variableKeys, variableKey)
		variableMap[variableKey] = normalizeImportedVariableValue(rawValue)
		transformed = append(transformed, fmt.Sprintf("%s%s {{ .%s }}", indent, key, variableKey))
	}

	normalizedVariables, _ := json.MarshalIndent(variableMap, "", "  ")
	if len(variableMap) == 0 {
		return strings.TrimSpace(raw), "{}", nil
	}
	return strings.TrimSpace(strings.Join(transformed, "\n")), string(normalizedVariables), variableKeys
}

func parseImportLine(line, runtime string) (string, string, string, bool) {
	indentLength := len(line) - len(strings.TrimLeft(line, " \t"))
	indent := line[:indentLength]
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", "", indent, false
	}

	if runtime == "fluentbit" {
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			return "", "", indent, false
		}
		key := parts[0]
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, key))
		return key, value, indent, true
	}

	if strings.HasPrefix(trimmed, "@type ") || strings.HasPrefix(trimmed, "@id ") {
		return "", "", indent, false
	}
	parts := strings.Fields(trimmed)
	if len(parts) < 2 {
		return "", "", indent, false
	}
	key := parts[0]
	value := strings.TrimSpace(strings.TrimPrefix(trimmed, key))
	return key, value, indent, true
}

func shouldExtractImportKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, ".", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return importExtractableKeySet[normalized]
}

func uniqueImportVariableKey(key string, used map[string]int) string {
	base := sanitizeImportVariableToken(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(key)), ".", "_"))
	if base == "" {
		base = "value"
	}
	used[base]++
	if used[base] == 1 {
		return base
	}
	return fmt.Sprintf("%s_%d", base, used[base])
}

func sanitizeImportVariableToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = importVariableTokenRegex.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return value
}

func normalizeImportedVariableValue(raw string) interface{} {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}
	if number, err := strconv.Atoi(value); err == nil {
		return number
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil && strings.Contains(value, ".") {
		return floatValue
	}
	return value
}

type fluentBitBlock struct {
	Kind   string
	Fields map[string]string
	Raw    string
}

func parseFluentBitBlocks(content string) []fluentBitBlock {
	sectionRe := regexp.MustCompile(`^\[([A-Za-z_]+)\]$`)
	blocks := []fluentBitBlock{}
	var current *fluentBitBlock
	currentLines := []string{}
	pendingBoundaryLines := []string{}
	seenBlockContent := false

	flush := func() {
		if current != nil {
			current.Raw = strings.TrimSpace(strings.Join(currentLines, "\n"))
			blocks = append(blocks, *current)
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(raw)
		if matches := sectionRe.FindStringSubmatch(trimmed); len(matches) == 2 {
			flush()
			current = &fluentBitBlock{
				Kind:   strings.ToUpper(matches[1]),
				Fields: map[string]string{},
			}
			currentLines = append(append([]string{}, pendingBoundaryLines...), raw)
			pendingBoundaryLines = nil
			seenBlockContent = false
			continue
		}
		if current == nil {
			continue
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			if seenBlockContent {
				pendingBoundaryLines = append(pendingBoundaryLines, raw)
			} else {
				currentLines = append(currentLines, raw)
			}
			continue
		}

		if len(pendingBoundaryLines) > 0 {
			currentLines = append(currentLines, pendingBoundaryLines...)
			pendingBoundaryLines = nil
		}
		currentLines = append(currentLines, raw)
		seenBlockContent = true
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(parts[0])
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, parts[0]))
		current.Fields[key] = value
	}
	if len(pendingBoundaryLines) > 0 {
		currentLines = append(currentLines, pendingBoundaryLines...)
	}
	flush()
	return blocks
}

type fluentdBlock struct {
	Kind       string
	Match      string
	Plugin     string
	ID         string
	Raw        string
	ParentKind string
	Depth      int
}

func parseFluentdBlocks(content string) []fluentdBlock {
	openRe := regexp.MustCompile(`^<([a-zA-Z_]+)(?:\s+([^>]+))?>$`)
	closeRe := regexp.MustCompile(`^</([a-zA-Z_]+)>$`)
	blocks := []fluentdBlock{}
	type fluentdParseState struct {
		block fluentdBlock
		lines []string
	}
	stack := []fluentdParseState{}

	appendCurrent := func(block fluentdBlock) {
		if block.Kind != "" {
			blocks = append(blocks, block)
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			for i := range stack {
				stack[i].lines = append(stack[i].lines, raw)
			}
			continue
		}
		if matches := openRe.FindStringSubmatch(line); len(matches) >= 2 {
			for i := range stack {
				stack[i].lines = append(stack[i].lines, raw)
			}
			parentKind := ""
			if len(stack) > 0 {
				parentKind = stack[len(stack)-1].block.Kind
			}
			stack = append(stack, fluentdParseState{
				block: fluentdBlock{
					Kind:       strings.ToLower(matches[1]),
					Match:      strings.TrimSpace(matches[2]),
					ParentKind: parentKind,
					Depth:      len(stack),
				},
				lines: []string{raw},
			})
			continue
		}
		if matches := closeRe.FindStringSubmatch(line); len(matches) == 2 {
			if len(stack) == 0 {
				continue
			}
			for i := range stack {
				stack[i].lines = append(stack[i].lines, raw)
			}
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			last.block.Raw = strings.TrimSpace(strings.Join(last.lines, "\n"))
			appendCurrent(last.block)
			continue
		}
		if len(stack) == 0 {
			continue
		}
		for i := range stack {
			stack[i].lines = append(stack[i].lines, raw)
		}
		if strings.HasPrefix(line, "@type ") {
			stack[len(stack)-1].block.Plugin = strings.TrimSpace(strings.TrimPrefix(line, "@type "))
			continue
		}
		if strings.HasPrefix(line, "@id ") {
			stack[len(stack)-1].block.ID = strings.TrimSpace(strings.TrimPrefix(line, "@id "))
		}
	}
	for i := len(stack) - 1; i >= 0; i-- {
		stack[i].block.Raw = strings.TrimSpace(strings.Join(stack[i].lines, "\n"))
		appendCurrent(stack[i].block)
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
