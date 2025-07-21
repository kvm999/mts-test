# MTS Test Assignment Makefile

# Variables
BACKEND_DIR := backend
SHARED_DIR := shared
CMD_DIR := $(BACKEND_DIR)/cmd
MIGRATION_DIR := $(BACKEND_DIR)/migration/postgres
CONFIG_FILE := $(BACKEND_DIR)/config.yaml
LOCAL_CONFIG_FILE := $(BACKEND_DIR)/config.local.yaml

# Binary names
BINARY_NAME := mts
BINARY_PATH := ./bin/$(BINARY_NAME)

# Database configuration (can be overridden)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= mts
DB_PASSWORD ?= mts_password_2024
DB_NAME ?= mts
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Go commands
GO_CMD := go
GO_BUILD := $(GO_CMD) build
GO_TEST := $(GO_CMD) test
GO_CLEAN := $(GO_CMD) clean
GO_MOD := $(GO_CMD) mod
GO_FMT := $(GO_CMD) fmt
GO_VET := $(GO_CMD) vet
GO_WORK := $(GO_CMD)

# Default target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
.PHONY: build
build: ## Build the application
	@echo "Building MTS application..."
	@mkdir -p bin
	cd $(BACKEND_DIR) && $(GO_BUILD) -o ../$(BINARY_PATH) ./cmd/

.PHONY: build-race
build-race: ## Build with race detector
	@echo "Building MTS application with race detector..."
	@mkdir -p bin
	cd $(BACKEND_DIR) && $(GO_BUILD) -race -o ../$(BINARY_PATH) ./cmd/

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "Building MTS application for Linux..."
	@mkdir -p bin
	cd $(BACKEND_DIR) && GOOS=linux GOARCH=amd64 $(GO_BUILD) -o ../$(BINARY_PATH)-linux ./cmd/

# Run targets
.PHONY: run
run: ## Run the application
	cd $(BACKEND_DIR) && $(GO_CMD) run ./cmd/

.PHONY: run-local
run-local: ## Run with local config
	cd $(BACKEND_DIR) && CONFIG_FILE=config.local.yaml $(GO_CMD) run ./cmd/

.PHONY: start
start: build ## Build and run the application
	$(BINARY_PATH)

# Development targets
.PHONY: dev
dev: ## Run in development mode with live reload (requires air)
	@command -v air >/dev/null 2>&1 || { echo "air is not installed. Run: go install github.com/cosmtrek/air@latest"; exit 1; }
	cd $(BACKEND_DIR) && air

# Test targets
.PHONY: test
test: ## Run tests
	cd backend && go test -v ./internal/...
	cd shared && go test -v ./...

# Documentation targets
.PHONY: docs
docs: ## Generate Swagger documentation
	@command -v swag >/dev/null 2>&1 || { echo "swag is not installed. Run: make install-dev-tools"; exit 1; }
	cd $(BACKEND_DIR) && swag init -g internal/transport/rest/app.go -o internal/transport/rest/docs

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t mts:latest -f $(BACKEND_DIR)/Dockerfile .


.PHONY: docker-up
docker-up: ## Start all Docker services
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all Docker services
	docker-compose down

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	docker-compose logs -f


.PHONY: info
info: ## Show project information
	@echo "MTS Test Assignment"
	@echo "=================="
	@echo "Go version: $(shell $(GO_CMD) version)"
	@echo "Backend module: $(shell cd $(BACKEND_DIR) && $(GO_CMD) list -m)"
	@echo "Shared module: $(shell cd $(SHARED_DIR) && $(GO_CMD) list -m)"
	@echo "Database URL: $(DB_URL)"
	@echo "Binary path: $(BINARY_PATH)"

.DEFAULT_GOAL := help