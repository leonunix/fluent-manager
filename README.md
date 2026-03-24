# Fluent Manager

**[日本語](README.ja.md)** | **[中文](README.zh.md)**

Centralized management platform for [Fluent Bit](https://fluentbit.io/) and [Fluentd](https://www.fluentd.org/) agents. Manage your entire logging infrastructure — topology, configuration, deployment, and monitoring — from a single dashboard.

## Features

- **Infrastructure Topology** — Organize nodes into data centers, regions, and clusters with auto-assignment rules
- **Configuration Management** — Template-based configs with versioning, modules, rendering preview, and lint
- **Remote Deployment** — Push configurations to agents across nodes, clusters, or entire data centers
- **Real-time Monitoring** — Agent heartbeat, metrics collection, runtime drift detection, and health views
- **Log Pipeline Visualization** — Graph-based view of forwarding topology from sources to aggregators
- **AI-assisted Analysis** — Log sample analysis and config generation powered by configurable LLM providers
- **RBAC & Scoping** — Role-based permissions with topology-level scope control (DC / region / cluster)
- **Multi-auth** — Local, LDAP, and SAML authentication
- **Agent Policies** — Layered override policies (global → environment → cluster → label selector) for agent runtime settings
- **i18n** — English, Chinese, and Japanese UI

## Deployment Options

Fluent Manager supports three deployment modes to fit different environments.

### Option 1: All-in-One Binary

A single binary with the frontend embedded. Simplest to deploy — just copy and run.

```bash
make build-all-in-one            # local platform
make build-all-in-one-linux      # Linux amd64 + arm64
```

```bash
cp config.yaml.example config.yaml
# edit config.yaml (database, auth, etc.)
./fluent-manager-server
```

The server serves both the API and the web UI on the same port (default `:8080`).

### Option 2: Separated (API + Frontend)

Backend serves API only. Frontend is a static build you deploy to Nginx, Caddy, or any web server.

```bash
# Build backend
make build-server-linux

# Build & package frontend
make frontend-package            # produces bin/fluent-manager-frontend.tar.gz
```

**Backend** — run the server binary with your `config.yaml`.

**Frontend** — extract the tarball to your web server root:

```bash
tar -xzf fluent-manager-frontend.tar.gz -C /usr/share/nginx/html
```

Nginx example:

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location /api/ {
        proxy_pass http://backend:8080;
    }
    location /saml/ {
        proxy_pass http://backend:8080;
    }
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### Option 3: Docker

```bash
# Server (includes frontend)
docker build -t fluent-manager .
docker run -p 8080:8080 -v ./config.yaml:/app/config.yaml fluent-manager

# Agent only
docker build --target runtime-agent -t fluent-manager-agent .
```

## Agent

The agent is a lightweight Go binary deployed on each managed node. It handles heartbeat, metrics collection, config sync, and remote command execution.

```bash
make build-agent                 # local platform
make build-agent-linux           # Linux amd64 + arm64
```

Configure with `agent.yaml` (see `agent.yaml.example`). Only `server_url` and `api_key` are required — everything else can be delivered by the server via Agent Policies.

## Quick Start

1. Start the server (any deployment option above)
2. Open the web UI — the setup wizard will guide you through initial configuration (database, admin account, auth, etc.)
3. Build the topology (data centers → regions → clusters)
4. Deploy agents to your nodes
5. Create config templates and push them out

## License

MIT
