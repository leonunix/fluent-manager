.PHONY: build build-server build-server-linux build-agent build-agent-linux agent build-linux \
       build-all-in-one build-all-in-one-linux \
       run clean frontend frontend-package copy-frontend-dist clean-frontend-dist

# ── API-only (separated mode) ───────────────────────────────────
build: build-server

build-server:
	go build -o bin/fluent-manager-server ./cmd/server

build-server-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/fluent-manager-server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/fluent-manager-server-linux-arm64 ./cmd/server

# ── All-in-one (embedded frontend) ──────────────────────────────
# Copies dist into cmd/server/ so //go:embed can pick it up, then
# builds with the embed_frontend tag.

copy-frontend-dist:
	rm -rf cmd/server/frontend_dist
	cp -r web/frontend/dist cmd/server/frontend_dist

clean-frontend-dist:
	rm -rf cmd/server/frontend_dist

build-all-in-one: frontend copy-frontend-dist
	go build -tags embed_frontend -o bin/fluent-manager-server ./cmd/server
	$(MAKE) clean-frontend-dist

build-all-in-one-linux: frontend copy-frontend-dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embed_frontend -o bin/fluent-manager-server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags embed_frontend -o bin/fluent-manager-server-linux-arm64 ./cmd/server
	$(MAKE) clean-frontend-dist

# ── Agent ────────────────────────────────────────────────────────
build-agent:
	mkdir -p bin
	cd agent-client && go build -o ../bin/fluent-manager-agent .

build-agent-linux:
	mkdir -p bin
	cd agent-client && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../bin/fluent-manager-agent-linux-amd64 .
	cd agent-client && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ../bin/fluent-manager-agent-linux-arm64 .

agent: build-agent
	mkdir -p scripts/ansible/files
	cp bin/fluent-manager-agent scripts/ansible/files/fluent-manager-agent

# ── Combos ───────────────────────────────────────────────────────
build-linux: build-server-linux build-agent-linux

run: build-server
	./bin/fluent-manager-server

# Frontend
frontend:
	cd web/frontend && npm install && npm run build

# Package frontend dist as tarball for separated deployment (nginx etc.)
frontend-package: frontend
	@mkdir -p bin
	tar -czf bin/fluent-manager-frontend.tar.gz -C web/frontend/dist .

# Development
dev:
	go run ./cmd/server

# Clean
clean:
	rm -rf bin/
	rm -rf web/frontend/dist
	rm -rf web/frontend/node_modules
	rm -rf cmd/server/frontend_dist

# Build all for Linux deployment
all: build-linux

# Docker
docker-build:
	docker build -t fluent-manager .
