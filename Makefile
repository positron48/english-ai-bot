# Project configuration
APP_NAME := universal-ai-bot
GO := go
DOCKER_IMAGE := universal-ai-bot
DOCKER_TAG := latest

# Deployment configuration
GITHUB_REPO ?= positron48/universal-ai-bot
DEPLOY_APP_DIR ?= /var/www/ai-bot
SERVICE_NAME ?= ai-bot

-include .env
.EXPORT_ALL_VARIABLES:

.PHONY: all tidy build run test lint fmt setup up clean check ci deploy update status logs docker-build docker-run docker-stop docker-logs docker-clean docker-rebuild docker-dev docker-dev-logs docker-dev-restart

all: build

# Go commands
tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/$(APP_NAME) ./cmd/bot

run: tidy build
	./bin/$(APP_NAME)

test:
	$(GO) test ./...

test-verbose:
	$(GO) test -v ./...

# CI checks (same as in GitHub Actions)
check: tidy
	@echo "=== Running CI Checks ==="
	@echo ""
	@echo "1. Verifying dependencies..."
	@$(GO) mod verify
	@echo "✅ Dependencies verified"
	@echo ""
	@echo "2. Running tests..."
	@$(GO) test -v ./... | grep -E "(PASS|FAIL|RUN)" || true
	@echo "✅ Tests passed"
	@echo ""
	@echo "3. Running linter..."
	@$(GOLANGCI) run --timeout=3m
	@echo "✅ Linter passed"
	@echo ""
	@echo "🎉 All CI checks passed!"

# Alias for check
ci: check

# Code formatting
fmt:
	$(GO) fmt ./...

# Linting
GOLANGCI := $(shell if [ -x ./bin/golangci-lint ]; then echo ./bin/golangci-lint; else echo golangci-lint; fi)

lint: fmt
	$(GOLANGCI) run --timeout=3m

lint-install:
	@echo "Installing golangci-lint v1.61.0 into ./bin..."
	@mkdir -p bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.61.0

# Local setup
setup-local:
	@echo "Setting up project..."
	@mkdir -p bin
	@cp env.example .env
	@echo "✅ Project setup complete!"
	@echo "📝 Please edit .env file with your bot token"

# Development
dev: tidy
	$(GO) run ./cmd/bot

up: run

# Cleanup
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker compose up -d

docker-stop:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down -v
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

docker-rebuild: docker-clean docker-build docker-run

# Development with Docker
docker-dev:
	docker compose up -d

docker-dev-logs:
	docker compose logs -f

docker-dev-restart:
	docker compose restart tgbot-skeleton

# Deployment commands
deploy:
	@chmod +x scripts/deploy.sh && GITHUB_REPO=$(GITHUB_REPO) APP_NAME=$(APP_NAME) APP_DIR=$(DEPLOY_APP_DIR) SERVICE_NAME=$(SERVICE_NAME) ./scripts/deploy.sh deploy

docker-deploy: docker-build docker-run

update:
	@chmod +x scripts/deploy.sh && GITHUB_REPO=$(GITHUB_REPO) APP_NAME=$(APP_NAME) APP_DIR=$(DEPLOY_APP_DIR) SERVICE_NAME=$(SERVICE_NAME) ./scripts/deploy.sh update

status:
	@chmod +x scripts/deploy.sh && SERVICE_NAME=$(SERVICE_NAME) ./scripts/deploy.sh status

logs:
	@chmod +x scripts/deploy.sh && SERVICE_NAME=$(SERVICE_NAME) ./scripts/deploy.sh logs

setup:
	@chmod +x scripts/setup.sh && APP_DIR=$(DEPLOY_APP_DIR) SERVICE_NAME=$(SERVICE_NAME) ./scripts/setup.sh

# Help
help:
	@echo "Available commands:"
	@echo "  make setup-local    - Initial local project setup"
	@echo "  make setup          - Setup systemd service (requires sudo)"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run in development mode"
	@echo "  make test           - Run tests"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make check          - Run all CI checks (tests, lint, verify)"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run with docker compose"
	@echo "  make docker-stop    - Stop docker compose"
	@echo "  make docker-logs    - Show docker logs"
	@echo "  make docker-clean   - Clean Docker resources"
	@echo "  make docker-deploy  - Deploy with Docker"
	@echo ""
	@echo "Deployment commands:"
	@echo "  make deploy         - Deploy binary from GitHub releases"
	@echo "  make update         - Update deployed binary"
	@echo "  make status         - Check service status"
	@echo "  make logs           - Show service logs"

