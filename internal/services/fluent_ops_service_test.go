package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupFluentOpsTest(t *testing.T) (*gorm.DB, FluentOpsService) {
	t.Helper()
	db := testutil.NewTestDB()
	return db, NewFluentOpsService(db)
}

func seedOpsCluster(t *testing.T, db *gorm.DB, suffix string) models.Cluster {
	t.Helper()
	dc := models.DataCenter{Name: "ops-dc-" + suffix}
	if err := db.Create(&dc).Error; err != nil {
		t.Fatalf("create dc: %v", err)
	}
	region := models.Region{Name: "ops-region-" + suffix, DataCenterID: dc.ID}
	if err := db.Create(&region).Error; err != nil {
		t.Fatalf("create region: %v", err)
	}
	cluster := models.Cluster{Name: "ops-cluster-" + suffix, RegionID: region.ID}
	if err := db.Create(&cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	return cluster
}

func seedOpsGroup(t *testing.T, db *gorm.DB, clusterID uint, suffix string) models.AggregationGroup {
	t.Helper()
	group := models.AggregationGroup{
		Name:      "ops-group-" + suffix,
		ClusterID: &clusterID,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	return group
}

func seedOpsOutputTarget(t *testing.T, db *gorm.DB, suffix string) models.OutputTarget {
	t.Helper()
	target := models.OutputTarget{
		Name:       "ops-output-" + suffix,
		FluentType: "shared",
		TargetType: "opensearch",
		Endpoint:   "https://opensearch.internal:9200",
		Settings:   `{"match":"*","host":"opensearch.internal","port":9200}`,
	}
	if err := db.Create(&target).Error; err != nil {
		t.Fatalf("create output target: %v", err)
	}
	return target
}

func TestPipelineCRUDAndGraph(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	clusterA := seedOpsCluster(t, db, "a")
	clusterB := seedOpsCluster(t, db, "b")
	group := seedOpsGroup(t, db, clusterB.ID, "agg")
	outputTarget := seedOpsOutputTarget(t, db, "main")

	pipeline, err := svc.CreatePipeline(&LogPipelineInput{
		Name:                          "edge-to-agg",
		FluentType:                    "fluentbit",
		Protocol:                      "forward",
		SourceClusterID:               &clusterA.ID,
		DestinationAggregationGroupID: &group.ID,
		TagStrategy:                   "cluster.hostname",
		Enabled:                       true,
	}, 1, nil)
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}
	if pipeline.Name != "edge-to-agg" {
		t.Fatalf("unexpected pipeline name: %s", pipeline.Name)
	}

	graph, err := svc.PipelineGraph(nil)
	if err != nil {
		t.Fatalf("graph: %v", err)
	}
	if len(graph.Nodes) < 2 || len(graph.Edges) != 1 {
		t.Fatalf("expected graph with nodes and 1 edge, got nodes=%d edges=%d", len(graph.Nodes), len(graph.Edges))
	}

	_, err = svc.UpdatePipeline(pipeline.ID, &LogPipelineInput{
		Name:                      "edge-to-output",
		FluentType:                "fluentbit",
		Protocol:                  "http",
		SourceClusterID:           &clusterA.ID,
		DestinationOutputTargetID: &outputTarget.ID,
		Enabled:                   true,
	}, nil)
	if err != nil {
		t.Fatalf("update pipeline: %v", err)
	}
}

func TestPipelineScopeEnforced(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	clusterA := seedOpsCluster(t, db, "scope-a")
	clusterB := seedOpsCluster(t, db, "scope-b")
	groupB := seedOpsGroup(t, db, clusterB.ID, "scope-b")

	_, err := svc.CreatePipeline(&LogPipelineInput{
		Name:                          "forbidden-pipeline",
		FluentType:                    "fluentbit",
		Protocol:                      "forward",
		SourceClusterID:               &clusterA.ID,
		DestinationAggregationGroupID: &groupB.ID,
		Enabled:                       true,
	}, 1, []uint{clusterA.ID})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestOutputTargetCRUD(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	target, err := svc.CreateOutputTarget(&OutputTargetInput{
		Name:       "opensearch-prod",
		FluentType: "shared",
		TargetType: "opensearch",
		Endpoint:   "https://opensearch.internal:9200",
		Settings:   `{"match":"*","host":"opensearch.internal","port":9200,"index":"logs-%Y.%m.%d"}`,
	}, 1)
	if err != nil {
		t.Fatalf("create output target: %v", err)
	}
	if target.TargetType != "opensearch" {
		t.Fatalf("unexpected target type: %s", target.TargetType)
	}

	updated, err := svc.UpdateOutputTarget(target.ID, &OutputTargetInput{
		Name:       "opensearch-prod",
		FluentType: "shared",
		TargetType: "opensearch",
		Endpoint:   "https://os.example.com:9200",
		Settings:   `{"match":"*","host":"os.example.com","port":9200,"index":"logs-live"}`,
	})
	if err != nil {
		t.Fatalf("update output target: %v", err)
	}
	if updated.Endpoint != "https://os.example.com:9200" {
		t.Fatalf("unexpected endpoint after update: %s", updated.Endpoint)
	}

	targets, err := svc.ListOutputTargets()
	if err != nil {
		t.Fatalf("list output targets: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 output target, got %d", len(targets))
	}
}

func TestLintConfigPersistsFindings(t *testing.T) {
	_, svc := setupFluentOpsTest(t)
	result, err := svc.LintConfig(&ConfigLintInput{
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail\n[OUTPUT]\n  Name forward\n  Host example\n{{.missing}}",
	}, 1)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Summary == "" {
		t.Fatal("expected summary")
	}
}

func TestLintConfigDetectsUndefinedFluentBitParser(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.LintConfig(&ConfigLintInput{
		FluentType: "fluentbit",
		Content: `[INPUT]
  Name tail
  Path /var/log/app.log
  Parser missing_parser

[OUTPUT]
  Name stdout
  Match *`,
	}, 1)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	if !containsFinding(result.Findings, "PARSER_UNDEFINED") {
		t.Fatalf("expected PARSER_UNDEFINED finding, got %#v", result.Findings)
	}
}

func TestLintConfigWarnsWhenParserDependsOnExternalFile(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.LintConfig(&ConfigLintInput{
		FluentType: "fluentbit",
		Content: `[SERVICE]
  Parsers_File /etc/td-agent-bit/custom_parsers.conf

[INPUT]
  Name tail
  Path /var/log/app.log
  Parser openstack_reqid

[OUTPUT]
  Name stdout
  Match *`,
	}, 1)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	if !containsFinding(result.Findings, "PARSER_EXTERNAL_VERIFY") {
		t.Fatalf("expected PARSER_EXTERNAL_VERIFY finding, got %#v", result.Findings)
	}
}

func TestLintConfigDetectsUndefinedFluentdParseBlock(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.LintConfig(&ConfigLintInput{
		FluentType: "fluentd",
		Content: `<source>
  @type tail
  path /var/log/app.log
  <parse>
    time_key time
  </parse>
</source>

<match app.**>
  @type stdout
</match>`,
	}, 1)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	if !containsFinding(result.Findings, "PARSER_UNDEFINED") {
		t.Fatalf("expected PARSER_UNDEFINED finding, got %#v", result.Findings)
	}
}

func TestLintConfigDetectsFluentdParserPluginWithoutNestedParse(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.LintConfig(&ConfigLintInput{
		FluentType: "fluentd",
		Content: `<filter app.**>
  @type parser
  key_name log
</filter>

<match app.**>
  @type stdout
</match>`,
	}, 1)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	if !containsFinding(result.Findings, "PARSER_UNDEFINED") {
		t.Fatalf("expected PARSER_UNDEFINED finding, got %#v", result.Findings)
	}
}

func TestLintConfigAcceptsFluentdParseFormatShorthand(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.LintConfig(&ConfigLintInput{
		FluentType: "fluentd",
		Content: `<source>
  @type tail
  path /var/log/app.log
  <parse>
    format json
  </parse>
</source>

<match app.**>
  @type stdout
</match>`,
	}, 1)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	if containsFinding(result.Findings, "PARSER_UNDEFINED") {
		t.Fatalf("did not expect PARSER_UNDEFINED finding, got %#v", result.Findings)
	}
}

func containsFinding(findings []models.ConfigAnalysisFinding, ruleCode string) bool {
	for _, finding := range findings {
		if finding.RuleCode == ruleCode {
			return true
		}
	}
	return false
}

func TestImportExistingConfigForFluentBit(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentbit",
		NamePrefix: "legacy-nginx",
		Content: `[SERVICE]
  Flush 1

[INPUT]
  Name tail
  Path /var/log/nginx/access.log

[FILTER]
  Name modify
  Match nginx.*

[OUTPUT]
  Name forward
  Match nginx.*`,
	})
	if err != nil {
		t.Fatalf("import existing config: %v", err)
	}
	if len(result.Modules) != 4 {
		t.Fatalf("expected 4 imported modules, got %d", len(result.Modules))
	}
	if result.Modules[0].ModuleType != "service" || result.Modules[1].ModuleType != "input" {
		t.Fatalf("unexpected module order: %#v", result.Modules)
	}
	if !strings.Contains(result.Modules[1].Content, "Path {{ .path }}") {
		t.Fatalf("expected path to be extracted into variables, got %q", result.Modules[1].Content)
	}
	if !strings.Contains(result.Modules[1].Variables, `"path": "/var/log/nginx/access.log"`) {
		t.Fatalf("expected extracted variables to contain path, got %q", result.Modules[1].Variables)
	}
	if len(result.FlowPath) == 0 {
		t.Fatal("expected flow path to be populated")
	}
	if result.Validation.Verdict == "" {
		t.Fatal("expected validation verdict")
	}
	if strings.TrimSpace(result.TemplateDraftContent) == "" {
		t.Fatal("expected template draft content")
	}
}

func TestImportExistingParserLibraryUsesWorkspaceAssetMode(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentbit",
		NamePrefix: "legacy-parsers",
		Content: `[PARSER]
  Name apache
  Format regex
  Regex ^(?<message>.*)$

[PARSER]
  Name nginx
  Format regex
  Regex ^(?<message>.*)$`,
	})
	if err != nil {
		t.Fatalf("import existing config: %v", err)
	}
	if result.ImportMode != "workspace_assets" {
		t.Fatalf("expected workspace_assets import mode, got %q", result.ImportMode)
	}
	if result.AutoAssembleSupported {
		t.Fatal("expected auto assembly to be disabled for parser-only import")
	}
	if len(result.Modules) != 2 {
		t.Fatalf("expected 2 parser modules, got %d", len(result.Modules))
	}
	if result.Modules[0].Name != "apache" {
		t.Fatalf("expected parser module name to use parser Name directly, got %q", result.Modules[0].Name)
	}
	if result.Modules[1].Name != "nginx" {
		t.Fatalf("expected parser module name to use parser Name directly, got %q", result.Modules[1].Name)
	}
	if result.SuggestedTemplateName != "legacy-parsers-assets" {
		t.Fatalf("expected asset-oriented suggested name, got %q", result.SuggestedTemplateName)
	}
}

func TestImportExistingConfigForFluentBitMovesSectionBannerToNextBlock(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentbit",
		NamePrefix: "legacy-openstack",
		Content: `[FILTER]
    Name   record_modifier
    Match vm.*
    Record kind openstack
    Record host ${HOSTNAME}
    Record component libvirt

# ============================================
# OUTPUT: Syslog -> OpenSearch
# ============================================
[OUTPUT]
    Name  es
    Match vm.*`,
	})
	if err != nil {
		t.Fatalf("import existing config: %v", err)
	}
	if len(result.Modules) != 2 {
		t.Fatalf("expected 2 imported modules, got %d", len(result.Modules))
	}
	filterModule := result.Modules[0]
	outputModule := result.Modules[1]
	if strings.Contains(filterModule.Content, "OUTPUT: Syslog -> OpenSearch") {
		t.Fatalf("expected output banner to stay out of filter module, got %q", filterModule.Content)
	}
	if !strings.Contains(outputModule.Content, "OUTPUT: Syslog -> OpenSearch") {
		t.Fatalf("expected output banner to move with output module, got %q", outputModule.Content)
	}
}

func TestImportExistingConfigMatchesExistingOutputTarget(t *testing.T) {
	db, svc := setupFluentOpsTest(t)

	if err := db.Create(&models.OutputTarget{
		Name:       "openstack-opensearch",
		FluentType: "shared",
		TargetType: "opensearch",
		Endpoint:   "https://opensearch.internal:9200",
		Settings:   `{"host":"opensearch.internal","port":9200,"index":"openstack-%Y.%m.%d"}`,
	}).Error; err != nil {
		t.Fatalf("create output target: %v", err)
	}

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentbit",
		NamePrefix: "legacy-openstack",
		Content: `[OUTPUT]
    Name  es
    Match openstack.*
    Host  opensearch.internal
    Port  9200
    Index openstack-%Y.%m.%d`,
	})
	if err != nil {
		t.Fatalf("import existing config: %v", err)
	}
	if len(result.Destinations) != 1 {
		t.Fatalf("expected 1 matched destination, got %d", len(result.Destinations))
	}
	if result.Destinations[0].Name != "openstack-opensearch" {
		t.Fatalf("expected matched destination name, got %#v", result.Destinations[0])
	}
	if result.Destinations[0].MatchType != "exact" {
		t.Fatalf("expected exact destination match, got %#v", result.Destinations[0])
	}
	if result.Modules[0].OutputTargetName != "openstack-opensearch" {
		t.Fatalf("expected output module to carry destination metadata, got %#v", result.Modules[0])
	}
}

func TestImportExistingConfigSanitizesVariableKeysForTemplateRendering(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentbit",
		NamePrefix: "legacy-service",
		Content: `[SERVICE]
    Flush 1
    DB.Sync normal
    Log-Level info`,
	})
	if err != nil {
		t.Fatalf("import existing config: %v", err)
	}
	if len(result.Modules) != 1 {
		t.Fatalf("expected 1 imported module, got %d", len(result.Modules))
	}
	module := result.Modules[0]
	if !strings.Contains(module.Content, "{{ .db_sync }}") {
		t.Fatalf("expected dotted key to become underscore placeholder, got %q", module.Content)
	}
	if !strings.Contains(module.Content, "{{ .log_level }}") {
		t.Fatalf("expected dashed key to become underscore placeholder, got %q", module.Content)
	}
	if _, err := renderModuleTemplate(module.Content, map[string]interface{}{
		"flush":     1,
		"db_sync":   "normal",
		"log_level": "info",
	}); err != nil {
		t.Fatalf("expected imported module template to render successfully, got %v", err)
	}
}

func TestImportExistingConfigForFluentdRetainsNestedParseSafely(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentd",
		NamePrefix: "legacy-app",
		Content: `<system>
  log_level info
</system>

<source>
  @type tail
  path /var/log/app.log
  <parse>
    @type json
  </parse>
</source>

<match app.**>
  @type forward
</match>`,
	})
	if err != nil {
		t.Fatalf("import existing fluentd config: %v", err)
	}
	if len(result.Modules) != 4 {
		t.Fatalf("expected 4 imported modules, got %d", len(result.Modules))
	}
	if result.Modules[1].ModuleType != "input" {
		t.Fatalf("expected source to become input, got %s", result.Modules[1].ModuleType)
	}
	if !strings.Contains(result.Modules[1].Content, "<parse>") {
		t.Fatalf("expected nested parse to remain inside source content, got %q", result.Modules[1].Content)
	}
	if result.Modules[2].ModuleType != "route" {
		t.Fatalf("expected companion route module, got %s", result.Modules[2].ModuleType)
	}
	if !strings.Contains(result.Modules[2].Content, "Match {{ .match }}") {
		t.Fatalf("expected route companion to expose match variable, got %q", result.Modules[2].Content)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected import warnings for baseline fluentd splitting")
	}
}

func TestImportExistingConfigSuggestsReuseForExactModuleMatch(t *testing.T) {
	db, svc := setupFluentOpsTest(t)

	if err := db.Create(&models.ConfigModule{
		Name:       "legacy-nginx-tail-input-01",
		ModuleType: "input",
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail\n  Path {{ .path }}",
		Variables:  `{"path":"/var/log/nginx/access.log"}`,
		CreatedBy:  1,
	}).Error; err != nil {
		t.Fatalf("create existing module: %v", err)
	}

	result, err := svc.ImportExistingConfig(&ConfigImportInput{
		FluentType: "fluentbit",
		NamePrefix: "legacy-nginx",
		Content: `[INPUT]
  Name tail
  Path /var/log/nginx/access.log`,
	})
	if err != nil {
		t.Fatalf("import existing config with reuse: %v", err)
	}
	if len(result.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result.Modules))
	}
	if result.Modules[0].ImportAction != "reuse_existing" {
		t.Fatalf("expected reuse_existing, got %s", result.Modules[0].ImportAction)
	}
	if result.Modules[0].ExistingModuleID == nil {
		t.Fatal("expected existing module id to be attached")
	}
}

func TestReplayConfigMatchesFiltersAndRoute(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.ReplayConfig(&ConfigReplayInput{
		FluentType: "fluentbit",
		Content: `[INPUT]
  Name tail
[FILTER]
  Name modify
  Match app.*
[OUTPUT]
  Name forward
  Alias forward-primary
  Match app.*`,
		SampleLog: `{"message":"hello"}`,
		SampleTag: "app.logs",
	})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if !result.RouteMatched {
		t.Fatal("expected route to match")
	}
	if result.FinalOutput != "forward-primary" {
		t.Fatalf("expected final output alias forward-primary, got %s", result.FinalOutput)
	}
	if result.FinalOutputType != "forward" {
		t.Fatalf("expected final output type forward, got %s", result.FinalOutputType)
	}
	if len(result.MatchedFilters) != 1 || result.MatchedFilters[0] != "modify" {
		t.Fatalf("expected matched modify filter, got %#v", result.MatchedFilters)
	}
	if result.DetectedParser != "json" {
		t.Fatalf("expected json parser detection, got %s", result.DetectedParser)
	}
}

func TestSemanticDiffDetectsMeaningfulChanges(t *testing.T) {
	_, svc := setupFluentOpsTest(t)

	result, err := svc.SemanticDiff(&ConfigSemanticDiffInput{
		FluentType: "fluentbit",
		BeforeContent: `[INPUT]
  Name tail
[OUTPUT]
  Name forward
  Match *`,
		AfterContent: `[INPUT]
  Name tail
[FILTER]
  Name grep
  Match *
[OUTPUT]
  Name loki
  Match *`,
	})
	if err != nil {
		t.Fatalf("semantic diff: %v", err)
	}
	if len(result.Changes) == 0 {
		t.Fatal("expected semantic changes")
	}
	foundOutputChange := false
	for _, change := range result.Changes {
		if change.Category == "output" && change.Item == "loki" && change.ChangeType == "added" {
			foundOutputChange = true
			break
		}
	}
	if !foundOutputChange {
		t.Fatalf("expected output change in semantic diff, got %#v", result.Changes)
	}
}

func TestCompatibilityCheckDetectsMissingPlugins(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	cluster := seedOpsCluster(t, db, "compat")
	node := models.Node{
		NodeUID:       "compat-node",
		Hostname:      "compat-host",
		FluentType:    "fluentbit",
		ClusterID:     &cluster.ID,
		Status:        "online",
		AgentVersion:  "1.0.0",
		FluentVersion: "3.0.0",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	profile := models.NodeFluentProfile{
		NodeID:               node.ID,
		LoadedPlugins:        `["tail","modify"]`,
		SupportsHotReload:    false,
		SupportsMultiline:    true,
		SupportsStorageLayer: true,
		SupportsForwardTLS:   false,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	result, err := svc.CheckCompatibility(&CompatibilityCheckInput{
		FluentType: "fluentbit",
		Content: `[INPUT]
  Name tail
[OUTPUT]
  Name forward
  Match *`,
		NodeID: &node.ID,
	}, []uint{cluster.ID})
	if err != nil {
		t.Fatalf("compatibility: %v", err)
	}
	if result.Compatible {
		t.Fatal("expected incompatible result due to missing forward plugin")
	}
	if len(result.MissingPlugins) != 1 || result.MissingPlugins[0] != "forward" {
		t.Fatalf("expected missing forward plugin, got %#v", result.MissingPlugins)
	}
}

func TestRuntimeRecommendationsSurfacePipelineAndNodeRisk(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	cluster := seedOpsCluster(t, db, "recommend")
	group := seedOpsGroup(t, db, cluster.ID, "recommend")

	node := models.Node{
		NodeUID:       "recommend-node",
		Hostname:      "recommend-host",
		FluentType:    "fluentbit",
		ClusterID:     &cluster.ID,
		NodeRole:      models.NodeRoleEdgeCollector,
		Status:        "online",
		AgentVersion:  "1.0.0",
		FluentVersion: "3.0.0",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	if err := db.Create(&models.NodeFluentProfile{
		NodeID:               node.ID,
		SupportsStorageLayer: false,
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := db.Create(&models.NodeRuntimeState{
		NodeID:              node.ID,
		DesiredConfigHash:   "desired-hash",
		EffectiveConfigHash: "actual-hash",
		QueueDepth:          2000,
		RetryCount:          20,
		LastError:           "reload failed",
	}).Error; err != nil {
		t.Fatalf("create runtime state: %v", err)
	}

	if err := db.Create(&models.LogPipeline{
		Name:                          "recommend-pipeline",
		FluentType:                    "fluentbit",
		Protocol:                      "forward",
		SourceClusterID:               &cluster.ID,
		DestinationAggregationGroupID: &group.ID,
		Enabled:                       true,
		CreatedBy:                     1,
	}).Error; err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	recommendations, err := svc.RuntimeRecommendations(nil)
	if err != nil {
		t.Fatalf("runtime recommendations: %v", err)
	}
	if len(recommendations) == 0 {
		t.Fatal("expected recommendations")
	}

	foundAggregationTarget := false
	foundBackpressure := false
	foundTLS := false
	for _, item := range recommendations {
		switch item.Title {
		case "edge collector has no aggregation target":
			foundAggregationTarget = true
		case "runtime backpressure is building up":
			foundBackpressure = true
		case "forward pipeline is not protected by TLS":
			foundTLS = true
		}
	}
	if !foundAggregationTarget || !foundBackpressure || !foundTLS {
		t.Fatalf("expected key recommendations, got %#v", recommendations)
	}
}

func TestRuntimeDriftUsesRuntimeState(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	cluster := seedOpsCluster(t, db, "drift")

	template := models.ConfigTemplate{Name: "ops-template", FluentType: "fluentbit", Content: "content", CreatedBy: 1}
	if err := db.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	version := models.ConfigVersion{TemplateID: template.ID, Version: 1, Content: "content", Hash: "desired-hash", CreatedBy: 1}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("create config version: %v", err)
	}
	cluster.ConfigID = &version.ID
	if err := db.Save(&cluster).Error; err != nil {
		t.Fatalf("save cluster config: %v", err)
	}

	node := models.Node{
		NodeUID:       "drift-node",
		Hostname:      "drift-host",
		FluentType:    "fluentbit",
		ClusterID:     &cluster.ID,
		Status:        "online",
		AgentVersion:  "1.0.0",
		FluentVersion: "3.0.0",
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	now := time.Now()
	state := models.NodeRuntimeState{
		NodeID:              node.ID,
		DesiredConfigHash:   "desired-hash",
		EffectiveConfigHash: "actual-hash",
		LastSyncAt:          &now,
	}
	if err := db.Create(&state).Error; err != nil {
		t.Fatalf("create runtime state: %v", err)
	}

	items, err := svc.ListRuntimeDrift(nil)
	if err != nil {
		t.Fatalf("list drift: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 drift item, got %d", len(items))
	}
	if items[0].Status != "drifted" {
		t.Fatalf("expected drifted status, got %s", items[0].Status)
	}
}

func TestAggregationGroupMetricsUseNodeTLSSupport(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	cluster := seedOpsCluster(t, db, "metrics")
	group := seedOpsGroup(t, db, cluster.ID, "metrics")

	nodeA := models.Node{
		NodeUID:            "metrics-a",
		Hostname:           "metrics-a",
		AggregationGroupID: &group.ID,
		ClusterID:          &cluster.ID,
		Status:             "online",
		AgentVersion:       "1.0.0",
		FluentType:         "fluentbit",
		FluentVersion:      "3.0.0",
	}
	nodeB := models.Node{
		NodeUID:            "metrics-b",
		Hostname:           "metrics-b",
		AggregationGroupID: &group.ID,
		ClusterID:          &cluster.ID,
		Status:             "offline",
		AgentVersion:       "1.0.0",
		FluentType:         "fluentbit",
		FluentVersion:      "3.0.0",
	}
	if err := db.Create(&nodeA).Error; err != nil {
		t.Fatalf("create node a: %v", err)
	}
	if err := db.Create(&nodeB).Error; err != nil {
		t.Fatalf("create node b: %v", err)
	}
	if err := db.Create(&models.NodeFluentProfile{NodeID: nodeA.ID, SupportsForwardTLS: true}).Error; err != nil {
		t.Fatalf("create profile a: %v", err)
	}
	if err := db.Create(&models.NodeFluentProfile{NodeID: nodeB.ID, SupportsForwardTLS: false}).Error; err != nil {
		t.Fatalf("create profile b: %v", err)
	}
	if err := db.Create(&models.NodeMetrics{NodeID: nodeA.ID, CPUUsagePercent: 70, MemUsagePercent: 60}).Error; err != nil {
		t.Fatalf("create metrics a: %v", err)
	}
	if err := db.Create(&models.NodeMetrics{NodeID: nodeB.ID, CPUUsagePercent: 50, MemUsagePercent: 40}).Error; err != nil {
		t.Fatalf("create metrics b: %v", err)
	}

	metric, err := svc.AggregationGroupMetrics(group.ID, []uint{cluster.ID})
	if err != nil {
		t.Fatalf("aggregation group metrics: %v", err)
	}
	if metric.AssignedNodes != 2 {
		t.Fatalf("expected 2 assigned nodes, got %d", metric.AssignedNodes)
	}
	if metric.OnlineNodes != 1 {
		t.Fatalf("expected 1 online node, got %d", metric.OnlineNodes)
	}
	if metric.AvgCPU != 60 {
		t.Fatalf("expected avg CPU 60, got %.1f", metric.AvgCPU)
	}
	if metric.AvgMem != 50 {
		t.Fatalf("expected avg mem 50, got %.1f", metric.AvgMem)
	}
	if metric.TLSCoverageRate != 50 {
		t.Fatalf("expected TLS coverage 50, got %.1f", metric.TLSCoverageRate)
	}
}

func TestPipelineAllowsNameReuseAfterSoftDelete(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	cluster := seedOpsCluster(t, db, "reuse")
	outputTarget := seedOpsOutputTarget(t, db, "reuse")

	pipeline, err := svc.CreatePipeline(&LogPipelineInput{
		Name:                      "pipeline-reusable",
		FluentType:                "fluentbit",
		Protocol:                  "http",
		SourceClusterID:           &cluster.ID,
		DestinationOutputTargetID: &outputTarget.ID,
		Enabled:                   true,
	}, 1, nil)
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}
	if err := svc.DeletePipeline(pipeline.ID, nil); err != nil {
		t.Fatalf("delete pipeline: %v", err)
	}
	recreated, err := svc.CreatePipeline(&LogPipelineInput{
		Name:                      "pipeline-reusable",
		FluentType:                "fluentbit",
		Protocol:                  "http",
		SourceClusterID:           &cluster.ID,
		DestinationOutputTargetID: &outputTarget.ID,
		Enabled:                   true,
	}, 1, nil)
	if err != nil {
		t.Fatalf("recreate pipeline with soft-deleted name: %v", err)
	}
	if recreated.ID == pipeline.ID {
		t.Fatalf("expected recreated pipeline to have a new id, got %d", recreated.ID)
	}
}

func TestListPipelinesRespectsGlobalAggregationGroupsForScopedUsers(t *testing.T) {
	db, svc := setupFluentOpsTest(t)
	cluster := seedOpsCluster(t, db, "scoped-global")
	globalGroup := models.AggregationGroup{Name: "ops-group-global"}
	if err := db.Create(&globalGroup).Error; err != nil {
		t.Fatalf("create global group: %v", err)
	}
	inScopeGroup := seedOpsGroup(t, db, cluster.ID, "scoped")

	if err := db.Create(&models.LogPipeline{
		Name:                          "pipeline-global-dest",
		FluentType:                    "fluentbit",
		Protocol:                      "forward",
		SourceClusterID:               &cluster.ID,
		DestinationAggregationGroupID: &globalGroup.ID,
		Enabled:                       true,
		CreatedBy:                     1,
	}).Error; err != nil {
		t.Fatalf("create global dest pipeline: %v", err)
	}
	if err := db.Create(&models.LogPipeline{
		Name:                          "pipeline-global-source",
		FluentType:                    "fluentbit",
		Protocol:                      "forward",
		SourceAggregationGroupID:      &globalGroup.ID,
		DestinationAggregationGroupID: &inScopeGroup.ID,
		Enabled:                       true,
		CreatedBy:                     1,
	}).Error; err != nil {
		t.Fatalf("create global source pipeline: %v", err)
	}

	pipelines, err := svc.ListPipelines([]uint{cluster.ID})
	if err != nil {
		t.Fatalf("list pipelines: %v", err)
	}
	if len(pipelines) != 2 {
		t.Fatalf("expected 2 scoped pipelines via global aggregation groups, got %d", len(pipelines))
	}
}
