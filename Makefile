.PHONY: build clean test test-unit test-integration lint fmt check deps help

# Build settings
BINARY_NAME=tailscale-mcp-server
BUILD_DIR=dist
CMD_DIR=cmd/tailscale-mcp-server
VERSION?=dev
LDFLAGS=-X main.version=$(VERSION)

# Go settings
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Install dependencies
	go mod download
	go mod tidy

build: deps ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

build-all: deps ## Build binaries for all platforms
	@mkdir -p $(BUILD_DIR)
 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

run: build ## Build and run the server
	./$(BUILD_DIR)/$(BINARY_NAME)

run-dev: ## Run with go run for development
	go run ./$(CMD_DIR)

##@ Testing

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests only
	go test -v -race -coverprofile=coverage.out ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -race -tags=integration ./test/integration/...

test-all: ## Run all tests including integration
	go test -v -race -coverprofile=coverage.out ./internal/... ./pkg/... ./cmd/...
	go test -v -race -tags=integration ./test/integration/...

test-coverage: test-unit ## Generate test coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

##@ Quality

lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

fmt: ## Format code
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install it with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

check: fmt lint test ## Run all quality checks

##@ Utilities

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install: build ## Install binary to GOPATH/bin
	go install $(LDFLAGS) ./$(CMD_DIR)

dev-setup: ## Set up development environment
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

##@ Docker

docker-build: ## Build Docker image
	docker build -t hexsleeves/tailscale-mcp-server:$(VERSION) .

docker-run: docker-build ## Build and run Docker container
	docker run --rm -it \
		-e TAILSCALE_API_KEY \
		-e TAILSCALE_TAILNET \
		hexsleeves/tailscale-mcp-server:$(VERSION)

##@ Information

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"

go-version: ## Show Go version
	go version
