# tailscale-mcp-server/Makefile
.PHONY: build test lint clean install docker

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X github.com/hexsleeves/tailscale-mcp-server/version.Version=$(VERSION) \
                   -X github.com/hexsleeves/tailscale-mcp-server/version.GitCommit=$(COMMIT) \
                   -X github.com/hexsleeves/tailscale-mcp-server/version.BuildTime=$(BUILD_TIME)"

# Build the binary
build:
	go build $(LDFLAGS) -o bin/tailscale-mcp-server ./cmd/tailscale-mcp-server

# Install the binary
install:
	go install $(LDFLAGS) ./cmd/tailscale-mcp-server

# Run tests
test:
	go test -v ./...

# Run integration tests
test-integration:
	go test -tags=integration -v ./test/integration/...

# Run linter
lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

# Clean build artifacts
clean:
	rm -rf bin/ dist/

# Build Docker image
docker:
	docker build -t tailscale-mcp-server:$(VERSION) .

# Development setup
dev-setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download
