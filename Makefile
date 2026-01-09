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

.PHONY: all tidy build run test lint fmt setup up clean check ci deploy update status logs docker-build docker-run docker-stop docker-logs docker-clean docker-rebuild docker-dev docker-dev-logs docker-dev-restart webapp-install webapp-dev webapp-build

all: build

# Go commands
tidy:
	$(GO) mod tidy

swagger:
	@echo "Generating Swagger documentation..."
	@go run github.com/swaggo/swag/cmd/swag@latest init \
		-g internal/web/router.go \
		-o docs/swagger \
		--parseDependency \
		--parseInternal \
		2>&1 | grep -vE "(warning: failed to get package name|warning: failed to evaluate const)" || true
	@echo "✅ Swagger documentation generated in docs/swagger/"

build: webapp-build
	@echo "Building Go binary..."
	$(GO) build -o bin/$(APP_NAME) ./cmd/bot
	@echo "✅ Build complete: bin/$(APP_NAME)"

run: tidy build
	./bin/$(APP_NAME)

test:
	$(GO) test ./...

test-verbose:
	$(GO) test -v ./...

# LLM integration tests (require AI_URL, AI_API_KEY env vars)
llm-words:
	@echo "Running LLM word cards integration tests..."
	@echo "⚠️  Requires: AI_URL, AI_API_KEY environment variables"
	$(GO) test -tags=integration -v -run '^TestLLM_WordCards$$' -count=1 ./internal/integration/llm/...

llm-training:
	@echo "Running LLM training cards integration tests..."
	@echo "⚠️  Requires: AI_URL, AI_API_KEY environment variables"
	$(GO) test -tags=integration -v -run '^TestLLM_TrainingCards$$' -count=1 ./internal/integration/llm/...

llm-all: llm-words llm-training
	@echo "✅ All LLM integration tests completed"

# CI checks (same as in GitHub Actions)
check: tidy
	@echo "=== Running CI Checks ==="
	@echo ""
	@echo "1. Checking webapp dependencies..."
	@cd webapp && npm install --prefer-offline --no-audit --no-fund > /dev/null 2>&1 || npm install --no-audit --no-fund
	@echo "✅ Webapp dependencies installed"
	@echo ""
	@echo "2. Running webapp type check..."
	@cd webapp && npm run type-check
	@echo "✅ Webapp type check passed"
	@echo ""
	@echo "3. Building webapp..."
	@cd webapp && npm run build
	@echo "✅ Webapp build passed"
	@echo ""
	@echo "4. Verifying Go dependencies..."
	@$(GO) mod verify
	@echo "✅ Go dependencies verified"
	@echo ""
	@echo "5. Running Go tests..."
	@$(GO) test -tags=test -coverprofile=coverage.out -covermode=atomic -v ./... > .go-test-output.txt 2>&1
	@grep -E "(PASS|FAIL|RUN)" .go-test-output.txt || true
	@rm -f .go-test-output.txt
	@echo "✅ Go tests passed"
	@echo ""
	@echo "6. Running Go linter..."
	@$(GOLANGCI) run --timeout=3m
	@echo "✅ Go linter passed"
	@echo ""
	@echo "7. Checking test coverage..."
	@COVERAGE=$$($(GO) tool cover -func=coverage.out | awk '/^total:/ {print $$3}' | sed 's/%//'); \
	if [ -z "$$COVERAGE" ]; then \
		echo "❌ Failed to get coverage"; \
		exit 1; \
	fi; \
	COVERAGE_INT=$$(echo "$$COVERAGE" | cut -d. -f1); \
	if [ "$$COVERAGE_INT" -lt 50 ]; then \
		echo "❌ Test coverage is $$COVERAGE% (minimum required: 50%)"; \
		exit 1; \
	fi; \
	echo "✅ Test coverage: $$COVERAGE% (minimum: 50%)"
	@echo ""
	@echo "🎉 All CI checks passed!"
	@COVERAGE=$$($(GO) tool cover -func=coverage.out | awk '/^total:/ {print $$3}'); echo "📊 Total test coverage: $$COVERAGE"

# Migration command for existing word cards
build-migrate:
	@echo "Building migration tool..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o bin/migrate_words ./cmd/migrate_words
	@echo "✅ Migration tool built: bin/migrate_words"

build-migrate-training:
	@echo "Building training cards migration tool..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o bin/migrate_training_cards ./cmd/migrate_training_cards
	@echo "✅ Training cards migration tool built: bin/migrate_training_cards"

migrate-words: build-migrate
	@echo "⚠️  Word Cards Migration Tool"
	@echo "⚠️  This will process all existing word cards and regenerate them with structured data"
	@echo "⚠️  Make sure you have a backup of your database!"
	@echo ""
	@echo "To run migration, execute: ./bin/migrate_words"
	@echo "Or run directly without confirmation: ./bin/migrate_words"

migrate-training-cards: build-migrate-training
	@echo "⚠️  Training Cards Migration Tool"
	@echo "⚠️  This will update all existing training cards with POS and display_word"
	@echo "⚠️  Make sure you have a backup of your database!"
	@echo ""
	@echo "To run migration, execute: ./bin/migrate_training_cards"

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
	@echo "Installing golangci-lint v2.1.0 into ./bin..."
	@mkdir -p bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v2.1.0

# Local setup
setup-local:
	@echo "Setting up project..."
	@mkdir -p bin
	@cp env.example .env
	@echo "✅ Project setup complete!"
	@echo "📝 Please edit .env file with your bot token"

# Webapp commands
webapp-install:
	@echo "Installing webapp dependencies..."
	@cd webapp && npm install

webapp-dev:
	@echo "Starting webapp dev server..."
	@cd webapp && npm run dev

webapp-build:
	@echo "Building webapp..."
	@cd webapp && npm run build

# Development - runs both backend and frontend
dev: tidy webapp-install
	@echo "Starting development environment..."
	@echo "Backend: http://localhost:8184"
	@echo "Frontend: http://localhost:5173/app"
	@echo "Note: Using test build tag - webapp files served by Vite dev server"
	@echo ""
	@echo "=== Backend logs ==="
	@trap 'kill 0' EXIT; \
	$(GO) run -tags=test ./cmd/bot 2>&1 | sed 's/^/[BACKEND] /' & \
	echo "=== Frontend logs ==="; \
	cd webapp && npm run dev 2>&1 | sed 's/^/[FRONTEND] /' & \
	wait 

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
	@echo "  make dev            - Run backend + frontend in development mode"
	@echo "  make webapp-install - Install webapp dependencies"
	@echo "  make webapp-dev     - Run webapp dev server only"
	@echo "  make webapp-build   - Build webapp for production"
	@echo "  make test           - Run tests"
	@echo "  make llm-words       - Run LLM word cards integration tests (requires AI_URL, AI_API_KEY)"
	@echo "  make llm-training   - Run LLM training cards integration tests (requires AI_URL, AI_API_KEY)"
	@echo "  make llm-all        - Run all LLM integration tests"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make check          - Run all CI checks (tests, lint, verify)"
	@echo "  make swagger        - Generate Swagger API documentation"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make migrate-words  - Migrate existing word cards to new structured format"
	@echo "  make migrate-training-cards  - Migrate existing training cards with POS and display_word"
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

