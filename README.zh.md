# Fluent Manager

**[English](README.md)** | **[日本語](README.ja.md)**

[Fluent Bit](https://fluentbit.io/) 和 [Fluentd](https://www.fluentd.org/) 的集中管理平台。在一个控制台上管理整个日志基础设施的拓扑、配置、部署和监控。

## 功能特性

- **基础设施拓扑** — 按数据中心、区域、集群组织节点，支持自动分配规则
- **配置管理** — 基于模板的配置，支持版本管理、模块化、渲染预览和 Lint 检查
- **远程部署** — 按节点、集群或数据中心维度推送配置
- **实时监控** — 心跳检测、指标采集、运行时漂移检测、健康视图
- **日志管道可视化** — 以图形展示从源端到聚合端的转发拓扑
- **AI 辅助分析** — 基于可配置 LLM 的日志样本分析与配置生成
- **RBAC 与作用域** — 基于角色的权限控制 + 拓扑级别的作用域限制（数据中心/区域/集群）
- **多认证方式** — 本地、LDAP、SAML 认证
- **Agent 策略** — 分层覆盖策略（全局 → 环境 → 集群 → 标签选择器）控制 Agent 运行时配置
- **多语言** — 英文、中文、日文界面

## 部署方式

### 方式一：合体部署（All-in-One）

前端内嵌到 Go 二进制中，单文件部署，最简单。

```bash
make build-all-in-one            # 本机平台
make build-all-in-one-linux      # Linux amd64 + arm64
```

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml（数据库、认证等）
./fluent-manager-server
```

API 和 Web 界面在同一端口提供服务（默认 `:8080`）。

### 方式二：前后端分离

后端只提供 API，前端作为静态文件部署到 Nginx 等 Web 服务器。

```bash
# 构建后端
make build-server-linux

# 构建并打包前端
make frontend-package            # 生成 bin/fluent-manager-frontend.tar.gz
```

**后端** — 配置好 `config.yaml` 后运行二进制。

**前端** — 解压到 Web 服务器根目录：

```bash
tar -xzf fluent-manager-frontend.tar.gz -C /usr/share/nginx/html
```

Nginx 配置示例：

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

### 方式三：Docker

```bash
# 服务端（含前端）
docker build -t fluent-manager .
docker run -p 8080:8080 -v ./config.yaml:/app/config.yaml fluent-manager

# 仅 Agent
docker build --target runtime-agent -t fluent-manager-agent .
```

## Agent

Agent 是部署在每个受管节点上的轻量 Go 程序，负责心跳上报、指标采集、配置同步和远程命令执行。

```bash
make agent                       # 构建本机平台 Agent，并复制到 scripts/ansible/files/fluent-manager-agent
make build-agent                 # 本机平台
make build-agent-linux           # Linux amd64 + arm64
```

通过 `agent.yaml`（参考 `agent.yaml.example`）配置。必填项仅 `server_url` 和 `api_key`，其余均可通过服务端的 Agent 策略下发。

使用 Ansible 角色批量部署时，`fm_agent_log_groups` 变量会根据目标 OS 自动选择日志读取所需的附加组（Debian/Ubuntu → `adm`，RHEL 系 → `systemd-journal`），无需手动配置即可读取 `/var/log` 下的系统日志。详见 [scripts/ansible/README.md](scripts/ansible/README.md)。

## 快速开始

1. 启动服务端（以上任一方式）
2. 打开 Web 界面 — 首次启动会进入配置向导，引导完成初始设置（数据库、管理员账号、认证等）
3. 构建拓扑（数据中心 → 区域 → 集群）
4. 在节点上部署 Agent
5. 创建配置模板并推送下发

## 许可证

MIT
