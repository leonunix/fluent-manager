# Fluent Manager

Enterprise-grade centralized management platform for Fluent Bit / Fluentd agents. Provides hierarchical infrastructure topology, configuration management, remote deployment, and real-time monitoring.

**Author**: do not add claude here. use default
**License**: MIT

## Architecture

```
┌──────────────┐     ┌──────────────────────────────────────────────┐
│  Vue 3 SPA   │────▶│  Go Server (Gin)                             │
│  (ECharts)   │     │  ├── Auth (JWT / LDAP / SAML)                │
└──────────────┘     │  ├── Handlers (REST API)                     │
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

- **Backend**: Go 1.22, Gin, GORM, JWT/LDAP/SAML auth
- **Frontend**: Vue 3, Vite, Bootstrap 5, ECharts, Pinia, Axios
- **Database**: SQLite (dev), MySQL / PostgreSQL (prod)
- **Agent**: Go binary with heartbeat, metrics, command execution

## Project Structure

```
cmd/server/main.go           # Server entry point, route registration
internal/
  agent/monitor.go            # Heartbeat monitor, offline detection
  auth/                       # JWT, LDAP, SAML services
  config/config.go            # YAML config loader
  handlers/                   # REST API handlers
    agent_handler.go          # Agent register, heartbeat, commands
    auth_handler.go           # Login, profile, password
    config_handler.go         # Config templates + versions
    deploy_handler.go         # Deployment tasks
    node_handler.go           # Node CRUD (scope-filtered)
    topology_handler.go       # DC/Region/Cluster/Environment + match rules + user scopes
    role_handler.go           # Roles + permissions
    user_handler.go           # User CRUD
  middleware/
    auth.go                   # JWTAuth, RequirePermission, ScopeFilter, AgentAuth
    audit.go                  # Audit logging middleware
  models/
    models.go                 # All GORM models
    database.go               # DB init, migrations, seed data
    scope.go                  # ClusterMatchRule matching, AllowedClusterIDs
agent-client/                 # Distributed agent (separate Go module)
web/frontend/                 # Vue 3 SPA
  src/api/index.js            # All API endpoints
  src/router/index.js         # Vue Router config
  src/views/                  # Page components
  src/components/             # Reusable components (TopologyGraph)
  src/store/auth.js           # Pinia auth store
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

Config inheritance: Node config > Cluster config (EffectiveConfigID)

## RBAC Model

Two levels of access control:

1. **Permission-based** (action level): `resource:action` pairs on roles
   - Resources: `nodes`, `topology`, `configs`, `users`, `roles`, `audit`
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

Key server config sections: `server`, `database`, `auth` (jwt/ldap/saml), `agent` (heartbeat/api_key), `log`

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
| Configs | CRUD `/configs/templates`, `/configs/templates/:id/versions` | JWT + configs perm |
| Deploys | `GET\|POST /deploys` (supports scope: node/cluster/region/datacenter) | JWT + configs perm |
| Users/Roles | CRUD `/users`, `/roles`, `GET /permissions` | JWT + users/roles perm |
| Audit | `GET /audit-logs` | JWT + audit perm |

## Frontend Pages

| Page | Route | Description |
|------|-------|-------------|
| Dashboard | `/` | KPI cards, ECharts (status pie, DC bar chart), topology overview, recent deploys/audit |
| Topology | `/topology` | ECharts tree graph view + management tree view, match rules editor |
| Environments | `/environments` | Environment CRUD with color picker, cluster association counts |
| Nodes | `/nodes` | Filterable node list (status, cluster, env, DC, search), pagination |
| Node Detail | `/nodes/:id` | Metrics, commands, logs |
| Configs | `/configs` | Template CRUD |
| Config Detail | `/configs/:id` | Versions, topology-scoped deployment |
| Deploys | `/deploys` | Deployment task list and detail |
| Users | `/users` | User CRUD with role assignment and scope editor |
| Roles | `/roles` | Role + permission management |
| Audit Logs | `/audit` | Paginated audit trail |

## Development Notes

- Database auto-migrates on startup. Delete `fluent_manager.db` to reset SQLite.
- Default admin: `admin` / `admin123`
- Agent API key is in config.yaml `agent.api_key`
- Frontend dev server proxies `/api` to `http://localhost:8080`
- ECharts is lazy-loaded per component (tree in Topology, pie+bar in Dashboard)
- Match rules are evaluated by priority (lower number = higher priority) on agent registration
- Scope filtering is applied via `ScopeFilter` middleware; handlers use `middleware.GetAllowedClusters(c)`
