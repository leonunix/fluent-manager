# Fluent Control Plane Roadmap

## Goal

Build Fluent Manager into a control plane for mixed `Fluent Bit` and `Fluentd`
deployments, including hierarchical forwarding such as:

- `edge Fluent Bit -> regional Fluentd aggregator -> downstream output`
- `Kubernetes DaemonSet Fluent Bit -> shared Fluentd aggregation group`
- `standalone Fluentd -> direct output`

The roadmap below is intentionally staged. Each phase should produce usable
product value, not only data-model scaffolding.

## Guiding Principles

- Treat `Fluent Bit` and `Fluentd` as different runtimes with overlapping but
  non-identical capabilities.
- Model log forwarding topology explicitly instead of inferring it from freeform
  text configuration.
- Make configuration generation deterministic and analyzable.
- Prefer modular config composition over monolithic templates.
- Every rollout should improve either correctness, observability, or operator
  confidence.

## Phase Overview

### P0: Domain Foundation

Objective: introduce the minimum domain model needed to reason about Fluent
roles, runtime capabilities, and aggregation topology.

Deliverables:

1. Node role model
   Supported roles:
   - `edge_collector`
   - `aggregator`
   - `gateway`
   - `standalone`

2. Node Fluent profile
   Per-node capability snapshot:
   - loaded plugins
   - hot reload support
   - multiline support
   - storage layer support
   - forward TLS support
   - metrics API support
   - last reported timestamp
   - arbitrary JSON metadata for future runtime facts

3. Aggregation group model
   Represents a logical group of Fluent aggregators, usually Fluentd:
   - group name / alias / description
   - runtime type (`fluentd` or `fluentbit`)
   - mode (`forward`, `http`, `custom`)
   - endpoint host / port
   - TLS flags
   - optional link to topology cluster

4. Basic management APIs
   - CRUD aggregation groups
   - read/update node Fluent profile

5. Read-path integration
   - node detail/list should expose role, aggregation group, and fluent profile

Suggested schema:

- `nodes.node_role`
- `nodes.aggregation_group_id`
- `node_fluent_profiles`
- `aggregation_groups`

### P1: Fluent-Aware Configuration Model

Objective: stop treating configuration as an opaque blob.

Deliverables:

1. Config module model
   Module types:
   - `service`
   - `input`
   - `parser`
   - `filter`
   - `route`
   - `output`

2. Runtime-specific template families
   - Fluent Bit templates
   - Fluentd templates
   - shared logical modules mapped to runtime-specific renderers

3. Rendered config snapshots
   - rendered content
   - runtime type and version target
   - source module graph
   - render variables
   - hash

4. Config composition rules
   - inheritance by topology
   - inheritance by node role
   - aggregation-group-level defaults
   - environment overrides

Suggested schema:

- `config_modules`
- `config_module_versions`
- `rendered_configs`
- `config_bindings`

### P2: Topology and Flow Awareness

Objective: represent log pipelines as first-class objects.

Deliverables:

1. Log pipeline model
   - source selector (clusters / groups / labels)
   - upstream role expectations
   - destination aggregation group or terminal output
   - tag strategy
   - protocol (`forward`, `http`, `kafka`, `loki`, etc.)

2. Aggregation relationships
   - edge nodes assigned to aggregation groups
   - groups forwarding to groups
   - groups forwarding to outputs

3. Pipeline visualization API
   - topology graph nodes and edges
   - runtime type and health summary per hop

4. Auto-assignment logic extension
   - assign node role
   - assign aggregation group
   - assign config baseline

Suggested schema:

- `log_pipelines`
- `pipeline_edges`
- `pipeline_outputs`

### P3: Fluent Intelligence

Objective: make configuration safer and easier to optimize.

Deliverables:

1. Semantic lint engine
   Rules split by runtime:
   - undefined parser references
   - unreachable match / tag combinations
   - dangerous buffer defaults
   - missing retry or storage settings
   - version-incompatible directives

2. Sample replay engine
   - paste sample log
   - choose runtime
   - show parser result
   - show filter transformations
   - show final route / output

3. Semantic diff
   Compare versions by meaning:
   - inputs added/removed
   - parser changes
   - route changes
   - output changes

4. Compatibility checks
   - plugin availability
   - runtime version support
   - hot reload feasibility

Suggested schema:

- `config_analysis_results`
- `config_analysis_findings`
- `compatibility_checks`

### P4: Runtime Observability and Drift

Objective: operate large Fluent estates safely.

Deliverables:

1. Runtime pipeline health
   - upstream/downstream connectivity
   - retry state
   - queue / chunk buildup
   - flush latency

2. Drift detection
   - desired rendered config hash
   - on-disk config hash
   - runtime-reported effective hash

3. Aggregation capacity views
   - nodes per group
   - throughput
   - error rate
   - hot outputs

4. Optimization recommendations
   - excessive regex parser usage
   - missing storage layer on edge collectors
   - weak forward failover layout
   - aggregator overload risk

Suggested schema:

- `node_runtime_states`
- `config_drift_reports`
- `aggregation_group_metrics`

## Suggested API Surface

### P0

- `GET /api/v1/aggregation-groups`
- `GET /api/v1/aggregation-groups/:id`
- `POST /api/v1/aggregation-groups`
- `PUT /api/v1/aggregation-groups/:id`
- `DELETE /api/v1/aggregation-groups/:id`
- `GET /api/v1/nodes/:id/fluent-profile`
- `PUT /api/v1/nodes/:id/fluent-profile`

### P1

- `GET /api/v1/config-modules`
- `POST /api/v1/config-modules`
- `POST /api/v1/rendered-configs/preview`
- `GET /api/v1/rendered-configs/:id`

### P2

- `GET /api/v1/log-pipelines`
- `POST /api/v1/log-pipelines`
- `GET /api/v1/log-pipelines/graph`

### P3

- `POST /api/v1/config-analysis/lint`
- `POST /api/v1/config-analysis/replay`
- `POST /api/v1/config-analysis/diff`
- `POST /api/v1/config-analysis/compatibility`
- `GET /api/v1/config-analysis/:id`

### P4

- `GET /api/v1/runtime/drift`
- `GET /api/v1/runtime/recommendations`
- `GET /api/v1/aggregation-groups/:id/metrics`
- `GET /api/v1/runtime/health/graph`

## Recommended Delivery Order

1. P0 domain model and APIs
2. P1 modular config + rendered config previews
3. P2 log pipeline graph and aggregation assignment
4. P3 semantic lint + replay
5. P4 runtime observability + drift

## Current Implementation Status

The roadmap baseline is now implemented across backend and the existing Vue
console, with some advanced items still intentionally heuristic rather than
deep parser-grade engines.

### Delivered P0

- aggregation group model with encrypted `shared_key`
- node role on the node record
- node fluent profile model
- aggregation group CRUD, deleted-list, restore API
- node fluent profile read/update API
- node list/detail preload support for fluent role, group and profile
- scope filtering and cluster reassignment validation for aggregation groups

### Delivered P1

- `config_modules`
- `config_module_versions`
- `rendered_configs`
- module CRUD and versioning APIs
- rendered config preview API with runtime-aware module selection
- frontend config workspace for templates, modules, preview and lint

### Delivered P2

- `log_pipelines`
- pipeline CRUD and scope enforcement
- pipeline graph API
- aggregation group metrics API
- frontend aggregation group management page
- frontend pipeline graph page

### Delivered P3

- persisted semantic lint results:
  - unresolved template variables
  - missing storage hints
  - weak forward security
  - missing retry hints
  - missing `<match>` routes for Fluentd sources
- sample replay API:
  - sample log parse
  - filter match trace
  - output route match trace
- semantic diff API:
  - inputs
  - parsers
  - filters
  - routes
  - outputs
  - plugin set changes
- compatibility check API:
  - node runtime family match
  - plugin availability
  - hot reload feasibility
  - multiline / storage / forward TLS capability checks

### Delivered P4

- `node_runtime_states`
- drift API based on desired vs effective config hash
- runtime health graph API
- aggregation group runtime metrics
- runtime recommendation API for:
  - missing aggregation target
  - missing persistent buffering capability
  - backpressure buildup
  - recent apply failures
  - config drift
  - insecure forward links
- frontend runtime page for drift, health graph and recommendations

## Remaining Advanced Follow-Ups

The baseline is complete, but the following would further raise the platform to
full enterprise depth:

- parser-grade replay with real directive execution instead of heuristic trace
- version-aware compatibility rules sourced from runtime capability matrices
- semantic diff for module graphs, not only rendered content
- recommendation history / suppression / acknowledgment workflow
- rollout guardrails that block deploy when replay or compatibility fail
