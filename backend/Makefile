.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down lint

# Variables
APP_NAME=auth-gateway
DOCKER_IMAGE=auth-gateway:latest
DOCKER_COMPOSE=docker compose
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	$(GO) run ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -cover ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f auth-gateway

docker-rebuild: ## Rebuild and restart services
	@echo "Rebuilding and restarting..."
	$(DOCKER_COMPOSE) up -d --build

migrate-up: ## Run database migrations (requires migrate tool)
	@echo "Running migrations..."
	@which migrate > /dev/null || (echo "migrate tool not installed. Install: https://github.com/golang-migrate/migrate" && exit 1)
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/auth_gateway?sslmode=disable" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@which migrate > /dev/null || (echo "migrate tool not installed" && exit 1)
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/auth_gateway?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	@which migrate > /dev/null || (echo "migrate tool not installed" && exit 1)
	migrate create -ext sql -dir migrations -seq $(NAME)

dev: ## Start development environment
	@echo "Starting development environment..."
	$(DOCKER_COMPOSE) up postgres redis -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services ready! You can now run: make run"

proto-gen: ## Generate Go code from proto files
	@echo "Generating proto files..."
	@chmod +x scripts/generate-proto.sh
	@./scripts/generate-proto.sh

grpc-example: ## Run gRPC client example
	@echo "Running gRPC client example..."
	cd examples/grpc-client && go run main.go

swagger-gen: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || (echo "swag not installed. Install: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init --pd --st -g cmd/server/main.go -o docs --parseInternal

all: clean deps build test ## Clean, download deps, build and test

.DEFAULT_GOAL := help
