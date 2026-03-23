# Fluent Manager

Enterprise-grade centralized management platform for Fluent Bit / Fluentd agents. Provides hierarchical infrastructure topology, configuration management, remote deployment, and real-time monitoring.

**Author**: do not add claude here. use default
**License**: MIT

## Architecture

```
┌──────────────┐     ┌──────────────────────────────────────────────┐
│  Vue 3 SPA   │────▶│  Go Server (Gin)                             │
│  (TypeScript) │     │  ├── Auth (JWT / LDAP / SAML)                │
│  (ECharts)   │     │  ├── Handlers (REST API, HTTP only)           │
└──────────────┘     │  ├── Services (business logic, interfaces)    │
                     │  ├── Middleware (RBAC + Scope + Audit)        │
                     │  └── Models (GORM: SQLite / MySQL / Postgres) │
                     └──────────────────┬───────────────────────────┘
                                        │ Agent API (heartbeat/config)
                     ┌──────────────────▼───────────────────────────┐
                     │  Agent Client (distributed on each node)      │
                     │  ├── Heartbeat + Metrics collection           │
                     │  ├── Config sync + hot-reload                 │
                     │  ├── Remote command execution                  │
                     │  └── Log shipping                             │
                     └──────────────────────────────────────────────┘
```

## Tech Stack

- **Backend**: Go 1.22, Gin, GORM, JWT/LDAP/SAML auth, layered (handler → service → model)
- **Frontend**: Vue 3, Vite, TypeScript, Bootstrap 5, ECharts, Pinia, Axios
- **Database**: SQLite (dev), MySQL / PostgreSQL (prod)
- **Agent**: Go binary with heartbeat, metrics, command execution

## Project Structure

```
cmd/server/main.go           # Server entry point, route registration, service wiring
internal/
  agent/monitor.go            # Heartbeat monitor, offline detection
  auth/                       # JWT, LDAP, SAML services
  cache/redis.go              # Redis cache (optional, graceful fallback)
  config/config.go            # YAML config loader
  handlers/                   # REST API handlers (HTTP parsing + response only)
    agent_handler.go          # Agent register, heartbeat, commands
    agent_policy_handler.go   # Agent policy CRUD, resolve preview, audit snapshots
    auth_handler.go           # Login, profile, password
    config_handler.go         # Config templates + modules + rendered previews
    deploy_handler.go         # Deployment tasks
    fluent_handler.go         # Aggregation groups + node fluent profiles
    fluent_ops_handler.go     # Pipelines, analysis, runtime drift/recommendations
    metrics_handler.go        # Aggregated metrics endpoints
    node_handler.go           # Node CRUD (scope-filtered)
    topology_handler.go       # DC/Region/Cluster/Environment + match rules + user scopes
    role_handler.go           # Roles + permissions
    user_handler.go           # User CRUD
  services/                   # Business logic layer (interface + implementation)
    services.go               # Registry, NewRegistry(db), ErrHasChildren
    auth_service.go           # AuthService: login, LDAP user, password
    user_service.go           # UserService: CRUD + bcrypt
    role_service.go           # RoleService: CRUD + permissions
    topology_service.go       # TopologyService: DC/Region/Cluster/Env/MatchRule/Scope/Tree
    node_service.go           # NodeService: CRUD + scope filter + stats
    agent_policy_service.go   # AgentPolicyService: scoped overrides + resolved delivery settings
    fluent_service.go         # FluentService: aggregation groups + node fluent profiles
    fluent_ops_service.go     # FluentOpsService: pipelines + analysis + runtime views
    config_service.go         # ConfigService: templates + modules + rendered previews
    deploy_service.go         # DeployService: deploy tasks + audit logs
    agent_service.go          # AgentService: register/heartbeat/command/log
    metrics_service.go        # MetricsService: aggregation + cache
  middleware/
    auth.go                   # JWTAuth, RequirePermission, ScopeFilter, AgentAuth
    audit.go                  # Audit logging middleware
  models/                     # GORM models (split by domain)
    auth.go                   # User, Role, Permission, UserScope
    topology.go               # DataCenter, Region, Cluster, Environment, ClusterMatchRule
    node.go                   # Node, NodeMetrics, RemoteCommand, NodeLog
    agent_policy.go           # AgentPolicy
    fluent.go                 # AggregationGroup, NodeFluentProfile, node roles
    fluent_ops.go             # LogPipeline, ConfigAnalysisResult, NodeRuntimeState
    config.go                 # ConfigTemplate, ConfigVersion, ConfigModule, RenderedConfig
    deploy.go                 # DeployTask, DeployRecord
    audit.go                  # AuditLog
    database.go               # DB init, migrations, seed data
    scope.go                  # ClusterMatchRule matching, AllowedClusterIDs
agent-client/                 # Distributed agent (separate Go module)
web/frontend/                 # Vue 3 SPA (TypeScript)
  tsconfig.json               # TypeScript configuration
  src/types/                  # Type definitions (auth, topology, node, config, deploy, audit, common)
  src/api/                    # API modules (split by domain, typed)
    client.ts                 # Axios instance + interceptors
    auth.ts                   # Auth API
    users.ts, roles.ts        # User/Role API
    topology.ts               # DC/Region/Cluster/Env/MatchRule/Scope API
    nodes.ts                  # Node API
    fluent.ts                 # Aggregation groups + node fluent profiles API
    fluent_ops.ts             # Pipelines + analysis + runtime API
    configs.ts                # Config template/version API
    deploys.ts, metrics.ts    # Deploy/Metrics API
    audit.ts                  # Audit log API
    index.ts                  # Barrel re-export (backward compatible)
  src/router/index.ts         # Vue Router config
  src/views/                  # Page components (.vue, not yet lang="ts")
  src/components/             # Reusable components (TopologyGraph, FluentFlowGraph)
  src/store/auth.ts           # Pinia auth store
  src/i18n/                   # i18n (index.ts + zh/en/ja.js)
```

## Data Model (Topology Hierarchy)

```
DataCenter  →  Region  →  Cluster  →  Node
                            ├── Environment (prod/staging/dev/test)
                            ├── MatchRules (auto-assign new nodes)
                            └── Config (inherited by nodes)
```

- **DataCenter**: Physical DC or cloud provider (aws, aliyun, azure, idc)
- **Region**: Logical zone (cn-east-1, us-west-2)
- **Cluster**: HA group with optional environment and inherited config
- **Node**: Fluent Bit/Fluentd agent, belongs to a cluster
- **ClusterMatchRule**: Auto-assigns new nodes by hostname glob, IP CIDR, labels, fluent_type, OS
- **Cluster.IsDefault**: Fallback for unmatched nodes
- **AggregationGroup**: Logical Fluentd / Fluent Bit fan-in target, optionally bound to a cluster
- **LogPipeline**: Explicit forwarding link from cluster/group/selector to aggregator or terminal output
- **NodeFluentProfile**: Runtime capability snapshot reported or edited per node
- **NodeRuntimeState**: Desired/effective hash, queue/retry/flush state for drift and health views
- **AgentPolicy**: Global / environment / cluster / label-selector override policy that is merged into the final server-delivered agent settings

Config inheritance: Node config > Cluster config (EffectiveConfigID)
Agent runtime inheritance: server bootstrap defaults > matching agent policies (ordered by priority) > node runtime detection/local persisted UID

## RBAC Model

Two levels of access control:

1. **Permission-based** (action level): `resource:action` pairs on roles
   - Resources: `nodes`, `topology`, `configs`, `users`, `roles`, `audit`, `agent_policies`
   - Actions: `create`, `read`, `update`, `delete`, `deploy`
   - Default roles: `admin` (all), `operator` (nodes+configs+topology), `viewer` (read-only)

2. **Scope-based** (resource level): `UserScope` binds users to specific topology
   - Scope types: `datacenter`, `region`, `cluster`
   - No scopes = global access (admin)
   - Scoped users only see their allowed DCs/regions/clusters/nodes

## Build & Run

```bash
# Server
make build-server              # or: go build -o bin/fluent-manager-server ./cmd/server
make dev                       # go run ./cmd/server

# Agent
make build-agent               # cd agent-client && go build

# Frontend
cd web/frontend
npm install
npm run dev                    # dev server on :3000, proxies API to :8080
npm run build                  # production build to dist/

# All
make all                       # server + agent
make frontend                  # npm install + build
```

## Configuration

- Server: `config.yaml` (see `config.yaml.example`)
- Agent: `agent.yaml` (see `agent.yaml.example`)

Key server config sections: `server`, `database`, `auth` (jwt/ldap/saml), `agent` (agent bootstrap defaults + API key + runtime delivery defaults), `fluent` (shared key encryption), `cache` (redis, optional), `log`

Agent bootstrap is intentionally minimal now:
- required: `server_url`, `api_key`
- optional: `node_uid` (auto-generated and persisted if omitted)
- runtime-specific fluent paths/commands/intervals can be delivered by the server and overridden by Agent Policies, instead of requiring local hand-written agent config on every node

## API Routes

All under `/api/v1`:

| Group | Endpoints | Auth |
|-------|-----------|------|
| Auth | `POST /auth/login`, `GET /auth/profile`, `PUT /auth/password` | JWT |
| Agent | `POST /agent/register\|heartbeat\|report\|command-result\|logs` | API Key |
| Topology | `GET /topology/tree`, CRUD `/datacenters`, `/regions`, `/clusters`, `/environments` | JWT + topology perm |
| Match Rules | CRUD `/clusters/:id/rules` | JWT + topology perm |
| User Scopes | `GET\|PUT /users/:id/scopes` | JWT + users perm |
| Nodes | `GET /nodes`, `GET /nodes/stats`, CRUD `/nodes/:id`, `POST /nodes/batch-move` | JWT + nodes perm + scope |
| Fluent | CRUD `/aggregation-groups`, `GET /aggregation-groups/deleted`, `POST /aggregation-groups/:id/restore`, `GET /aggregation-groups/:id/metrics`, `GET/PUT /nodes/:id/fluent-profile` | JWT + topology/nodes perm |
| Pipelines | CRUD `/log-pipelines`, `GET /log-pipelines/graph` | JWT + topology perm |
| Configs | CRUD `/configs/templates`, `/configs/templates/:id/versions`, `/configs/modules`, `/configs/modules/:id/versions`, `POST /configs/rendered-configs/preview`, `GET /configs/rendered-configs/:id` | JWT + configs perm |
| Config Analysis | `POST /config-analysis/lint`, `POST /config-analysis/replay`, `POST /config-analysis/diff`, `POST /config-analysis/compatibility`, `GET /config-analysis/:id` | JWT + configs perm |
| Deploys | `GET\|POST /deploys` (supports scope: node/cluster/region/datacenter) | JWT + configs perm |
| Runtime | `GET /runtime/drift`, `GET /runtime/health/graph`, `GET /runtime/recommendations` | JWT + nodes perm |
| Agent Policies | CRUD `/agent-policies`, `GET /agent-policies/resolve/:node_id` | JWT + agent_policies perm (resolve uses nodes read) |
| Users/Roles | CRUD `/users`, `/roles`, `GET /permissions` | JWT + users/roles perm |
| Audit | `GET /audit-logs` | JWT + audit perm |

## Frontend Pages

| Page | Route | Description |
|------|-------|-------------|
| Dashboard | `/` | KPI cards, ECharts (status pie, DC bar chart), topology overview, recent deploys/audit |
| Topology | `/topology` | ECharts tree graph view + management tree view, match rules editor |
| Environments | `/environments` | Environment CRUD with color picker, cluster association counts |
| Nodes | `/nodes` | Filterable node list (status, cluster, env, DC, search), pagination |
| Node Detail | `/nodes/:id` | Metrics, commands, logs, fluent role/profile |
| Aggregation Groups | `/aggregation-groups` | Fluent fan-in group CRUD, deleted list restore, runtime metrics |
| Pipelines | `/pipelines` | Explicit forwarding topology graph and pipeline CRUD |
| Configs | `/configs` | Template CRUD, module workspace, render preview, lint/replay/compatibility |
| Config Detail | `/configs/:id` | Versions, topology-scoped deployment |
| Deploys | `/deploys` | Deployment task list and detail |
| Runtime | `/runtime` | Drift table, runtime health graph, optimization recommendations |
| Agent Policies | `/agent-policies` | Scoped policy CRUD, current user scope badges, resolved preview, node search, and policy targeting UX |
| Users | `/users` | User CRUD with role assignment and scope editor |
| Roles | `/roles` | Role + permission management |
| Audit Logs | `/audit` | Paginated audit trail with expandable Agent Policy field-level diffs |

## Layering Convention

```text
Handler (HTTP) → Service (business logic, interface) → Model (GORM/DB)
```

- **Handlers** only parse requests, extract auth context, call service, return HTTP response
- **Services** contain all business logic; each has an interface for testability
- **Models** are pure GORM structs split by domain; `database.go` handles init/seed, `scope.go` handles RBAC scope resolution
- Services are initialized via `services.NewRegistry(db, fluentSharedKeySecret)` in `main.go` and injected into handlers
- Fluent-specific logic is split into:
  - `FluentService` for aggregation groups and per-node runtime capability metadata
  - `FluentOpsService` for pipelines, analysis, drift, metrics, and recommendations
- Agent runtime delivery is split into:
  - `AgentService` for registration, heartbeat, reporting, commands, and log upload
  - `AgentPolicyService` for final delivered agent settings, scoped overrides, and node resolve previews
- `scope.go` functions (`AllowedClusterIDs`, `AutoAssignCluster`) still use global `models.DB` (to be refactored later)

## Development Notes

- Database auto-migrates on startup. Delete `fluent_manager.db` to reset SQLite.
- Default admin: `admin` / `admin123`
- Agent API key is in config.yaml `agent.api_key`
- `/auth/profile` includes `roles`, `permissions`, and `scopes`; frontend uses it for UX-side scope hints and secondary cluster filtering
- Frontend dev server proxies `/api` to `http://localhost:8080`
- Frontend uses TypeScript (`strict: false` for incremental adoption); Vue components remain JS (`<script setup>` without `lang="ts"`)
- ECharts is lazy-loaded per component (tree in Topology, pie+bar in Dashboard)
- Match rules are evaluated by priority (lower number = higher priority) on agent registration
- Scope filtering is applied via `ScopeFilter` middleware; handlers use `middleware.GetAllowedClusters(c)`
- Agent Policy reads are scope-filtered; scoped users can only create/manage cluster-scoped policies inside their allowed clusters
- Redis cache is optional; metrics service falls back to direct DB queries if cache is disabled
- Aggregation group shared keys are encrypted at rest via `fluent.shared_key_secret` (falls back to JWT secret if unset)
- Agent Policy create/update/delete writes structured audit detail; the audit page can expand and render field-level before/after changes
- Semantic replay / diff / compatibility are baseline heuristics today, designed to be upgraded into deeper parser-grade engines later
