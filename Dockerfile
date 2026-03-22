# Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web/frontend
COPY web/frontend/package*.json ./
RUN npm ci
COPY web/frontend/ ./
RUN npm run build

# Build server
FROM golang:1.22-alpine AS server
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/frontend/dist ./web/frontend/dist
RUN CGO_ENABLED=1 go build -o fluent-manager-server ./cmd/server

# Build agent (separate module, minimal dependencies)
FROM golang:1.22-alpine AS agent
WORKDIR /app
COPY agent-client/ ./
RUN go mod download && CGO_ENABLED=0 go build -o fluent-manager-agent .

# Server runtime image
FROM alpine:3.20 AS runtime-server
RUN apk add --no-cache ca-certificates tzdata sqlite
WORKDIR /app
COPY --from=server /app/fluent-manager-server .
COPY --from=server /app/web/frontend/dist ./web/frontend/dist
COPY config.yaml.example config.yaml
EXPOSE 8080
CMD ["./fluent-manager-server"]

# Agent runtime image (use: docker build --target runtime-agent)
FROM alpine:3.20 AS runtime-agent
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=agent /app/fluent-manager-agent .
COPY agent.yaml.example agent.yaml
CMD ["./fluent-manager-agent", "-config", "agent.yaml"]
