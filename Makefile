.PHONY: build build-server build-agent run clean frontend

# Build server only (agent is a separate module in agent-client/)
build: build-server

build-server:
	go build -o bin/fluent-manager-server ./cmd/server

build-agent:
	cd agent-client && go build -o ../bin/fluent-manager-agent .

run: build-server
	./bin/fluent-manager-server

# Frontend
frontend:
	cd web/frontend && npm install && npm run build

# Development
dev:
	go run ./cmd/server

# Clean
clean:
	rm -rf bin/
	rm -rf web/frontend/dist
	rm -rf web/frontend/node_modules

# Build all (server + agent)
all: build-server build-agent

# Docker
docker-build:
	docker build -t fluent-manager .
