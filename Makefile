.PHONY: build build-server build-server-linux build-agent build-agent-linux build-linux run clean frontend

# Build for current platform (local dev)
build: build-server

build-server:
	go build -o bin/fluent-manager-server ./cmd/server

build-agent:
	cd agent-client && go build -o ../bin/fluent-manager-agent .

# Cross-compile for Linux (amd64 + arm64)
build-server-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/fluent-manager-server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/fluent-manager-server-linux-arm64 ./cmd/server

build-agent-linux:
	cd agent-client && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../bin/fluent-manager-agent-linux-amd64 .
	cd agent-client && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ../bin/fluent-manager-agent-linux-arm64 .

build-linux: build-server-linux build-agent-linux

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

# Build all for Linux deployment
all: build-linux

# Docker
docker-build:
	docker build -t fluent-manager .
