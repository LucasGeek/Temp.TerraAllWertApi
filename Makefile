.PHONY: help run dev test build clean docker-up docker-down migrate seed lint fmt

# Variables
APP_NAME := terra-allwert-api
MAIN_PATH := src/main.go
TEST_PATH := ./test/...
BUILD_DIR := bin
DOCKER_COMPOSE := docker-compose -f docker/docker-compose.local.yml
DOCKER_COMPOSE_INFRA := docker-compose -f docker/docker-compose.infra.yml

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo '${GREEN}Usage:${NC}'
	@echo '  ${YELLOW}make${NC} ${GREEN}<target>${NC}'
	@echo ''
	@echo '${GREEN}Targets:${NC}'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-15s${NC} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally
	@echo "${GREEN}Starting application...${NC}"
	cd src && go run main.go

dev: ## Run with hot reload using air
	@echo "${GREEN}Starting development server with hot reload...${NC}"
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	cd src && air

test: ## Run all tests
	@echo "${GREEN}Running tests...${NC}"
	cd test && go test -v -race -cover ./...

test-unit: ## Run unit tests only
	@echo "${GREEN}Running unit tests...${NC}"
	cd test && go test -v -race -cover ./unit/...

test-integration: ## Run integration tests
	@echo "${GREEN}Running integration tests...${NC}"
	cd test && go test -v -race -cover ./integration/...

test-e2e: ## Run end-to-end tests
	@echo "${GREEN}Running e2e tests...${NC}"
	cd test && go test -v -race -cover ./e2e/...

test-coverage: ## Run tests with coverage report
	@echo "${GREEN}Running tests with coverage...${NC}"
	cd test && go test -v -race -coverprofile=coverage.out ./...
	cd test && go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: test/coverage.html${NC}"

build: ## Build the application binary
	@echo "${GREEN}Building application...${NC}"
	@mkdir -p $(BUILD_DIR)
	cd src && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../$(BUILD_DIR)/$(APP_NAME) main.go
	@echo "${GREEN}Binary created at: $(BUILD_DIR)/$(APP_NAME)${NC}"

build-docker: ## Build Docker image
	@echo "${GREEN}Building Docker image...${NC}"
	docker build -f docker/Dockerfile -t $(APP_NAME):latest .

clean: ## Clean build artifacts
	@echo "${YELLOW}Cleaning build artifacts...${NC}"
	rm -rf $(BUILD_DIR)
	rm -rf src/tmp
	rm -rf test/coverage.*
	@echo "${GREEN}Clean complete${NC}"

# Docker commands
docker-up: ## Start all containers (app + infrastructure)
	@echo "${GREEN}Starting Docker containers...${NC}"
	$(DOCKER_COMPOSE) up -d
	@echo "${GREEN}Containers started. API available at http://localhost:3000${NC}"

docker-down: ## Stop all containers
	@echo "${YELLOW}Stopping Docker containers...${NC}"
	$(DOCKER_COMPOSE) down

docker-logs: ## Show container logs
	$(DOCKER_COMPOSE) logs -f

docker-infra-up: ## Start only infrastructure containers (DB + Redis)
	@echo "${GREEN}Starting infrastructure containers...${NC}"
	$(DOCKER_COMPOSE_INFRA) up -d
	@echo "${GREEN}Infrastructure ready${NC}"

docker-infra-down: ## Stop infrastructure containers
	@echo "${YELLOW}Stopping infrastructure containers...${NC}"
	$(DOCKER_COMPOSE_INFRA) down

docker-clean: ## Remove containers and volumes
	@echo "${RED}Removing all containers and volumes...${NC}"
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE_INFRA) down -v

# Database commands
migrate: ## Run database migrations
	@echo "${GREEN}Running migrations...${NC}"
	cd src && go run main.go migrate up

migrate-down: ## Rollback last migration
	@echo "${YELLOW}Rolling back migration...${NC}"
	cd src && go run main.go migrate down

seed: ## Seed the database
	@echo "${GREEN}Seeding database...${NC}"
	cd src && go run main.go seed

db-reset: ## Reset database (drop, create, migrate, seed)
	@echo "${RED}Resetting database...${NC}"
	cd src && go run main.go db:reset

# Code quality
lint: ## Run linter
	@echo "${GREEN}Running linter...${NC}"
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	cd src && golangci-lint run ./...
	cd test && golangci-lint run ./...

fmt: ## Format code
	@echo "${GREEN}Formatting code...${NC}"
	cd src && go fmt ./...
	cd test && go fmt ./...
	@echo "${GREEN}Code formatted${NC}"

vet: ## Run go vet
	@echo "${GREEN}Running go vet...${NC}"
	cd src && go vet ./...
	cd test && go vet ./...

tidy: ## Tidy and download dependencies
	@echo "${GREEN}Tidying dependencies...${NC}"
	cd src && go mod tidy
	cd test && go mod tidy
	go work sync
	@echo "${GREEN}Dependencies updated${NC}"

# Installation and setup
install: ## Install dependencies
	@echo "${GREEN}Installing dependencies...${NC}"
	cd src && go mod download
	cd test && go mod download
	@echo "${GREEN}Installing development tools...${NC}"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "${GREEN}Installation complete${NC}"

setup: install docker-infra-up ## Complete project setup
	@echo "${GREEN}Setting up project...${NC}"
	@cp -n .env.example .env || true
	@echo "${GREEN}Project setup complete!${NC}"
	@echo "${YELLOW}Run 'make dev' to start development server${NC}"

# Monitoring and debugging
health: ## Check application health
	@echo "${GREEN}Checking application health...${NC}"
	@curl -f http://localhost:3000/health || echo "${RED}Application is not running${NC}"

logs: ## Show application logs
	@tail -f logs/app.log

# Git hooks
pre-commit: fmt lint test ## Run pre-commit checks
	@echo "${GREEN}Pre-commit checks passed!${NC}"

# Default target
.DEFAULT_GOAL := help