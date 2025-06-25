# DNS Sync Makefile

# Build variables
BINARY_NAME=dns-sync
MAIN_PACKAGE=cmd/dns-sync/main.go
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Default target
.DEFAULT_GOAL := build

# Help target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
.PHONY: build
build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

.PHONY: build-linux
build-linux: ## Build for Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

.PHONY: build-windows
build-windows: ## Build for Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

.PHONY: build-darwin
build-darwin: ## Build for macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

.PHONY: build-all
build-all: build-linux build-windows build-darwin ## Build for all platforms

# Development targets
.PHONY: run
run: build ## Build and run the application
	./$(BUILD_DIR)/$(BINARY_NAME) -config config.yaml

.PHONY: run-dry
run-dry: build ## Build and run in dry-run mode
	./$(BUILD_DIR)/$(BINARY_NAME) -config config.yaml -dry-run

.PHONY: dev
dev: ## Run in development mode with debug logging
	$(GOCMD) run $(MAIN_PACKAGE) -config config.yaml -log-level debug

# Testing targets
.PHONY: test
test: ## Run tests
	$(GOTEST) -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race: ## Run tests with race detection
	$(GOTEST) -v -race ./...

# Dependency management
.PHONY: deps
deps: ## Download dependencies
	$(GOMOD) download

.PHONY: deps-update
deps-update: ## Update dependencies
	$(GOMOD) tidy
	$(GOGET) -u ./...

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	$(GOMOD) verify

# Code quality targets
.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	$(GOCMD) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GOCMD) vet ./...

# Cleanup targets
.PHONY: clean
clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html


# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t dns-sync:$(VERSION) -t dns-sync:latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run --rm -v $(PWD)/config.yaml:/app/config.yaml dns-sync:latest

# Release targets
.PHONY: release
release: clean build-all ## Create release build
	@echo "Release $(VERSION) built successfully"

# Installation targets
.PHONY: install
install: build ## Install binary to GOPATH/bin
	$(GOCMD) install $(LDFLAGS) $(MAIN_PACKAGE)


# Ensure build directory exists
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

