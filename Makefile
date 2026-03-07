# Project configuration
APP_NAME := universal-ai-bot
GO := go
DOCKER_IMAGE := universal-ai-bot
DOCKER_TAG := latest
PG_TEST_PORT ?= 55433

# Deployment configuration
GITHUB_REPO ?= positron48/universal-ai-bot
DEPLOY_APP_DIR ?= /var/www/ai-bot
SERVICE_NAME ?= ai-bot

-include .env
.EXPORT_ALL_VARIABLES:

.PHONY: all tidy build run test lint fmt setup up clean check check-quick ci deploy update status logs docker-build docker-run docker-stop docker-logs docker-clean docker-rebuild docker-dev docker-dev-logs docker-dev-restart webapp-install webapp-dev webapp-build test-postgres test-integration test-integration-verbose

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

grammar-bundle:
	@echo "Updating grammar submodule..."
	@git submodule update --remote
	@echo "✅ Grammar submodule updated"
	@echo ""
	@echo "Generating grammar bundle..."
	@./scripts/generate-grammar-bundle.sh
	@echo "✅ Grammar bundle generated"

tag:
	@V=$$(git describe --tags --abbrev=0 2>/dev/null | sed -E 's/^v?//' || echo "0.0.0"); \
	MAJOR=$$(echo $$V | cut -d. -f1); \
	MINOR=$$(echo $$V | cut -d. -f2); \
	PATCH=$$(echo $$V | cut -d. -f3); \
	NEXT=$$MAJOR.$$MINOR.$$((PATCH+1)); \
	echo "Creating tag $$NEXT"; \
	git tag $$NEXT; \
	echo "Pushing HEAD and tags to origin"; \
	git push origin HEAD --tags

build: webapp-build grammar-bundle
	@echo "Building Go binary..."
	$(GO) build -o bin/$(APP_NAME) ./cmd/bot
	@echo "✅ Build complete: bin/$(APP_NAME)"

run: postgres-up tidy build
	@echo ""; echo "Web app: http://localhost:$${SERVER_PORT:-8184}/app"; echo ""
	./bin/$(APP_NAME)

test:
	$(GO) test ./...

test-verbose:
	$(GO) test -v ./...

test-postgres:
	@echo "Starting temporary Postgres container..."
	@docker rm -f english-test-postgres >/dev/null 2>&1 || true
	@docker run -d --name english-test-postgres \
		-e POSTGRES_DB=english_test \
		-e POSTGRES_USER=english \
		-e POSTGRES_PASSWORD=english \
		-p $(PG_TEST_PORT):5432 \
		postgres:16-alpine >/dev/null
	@echo "Waiting for Postgres to become ready..."
	@i=0; until docker exec english-test-postgres pg_isready -U english -d english_test >/dev/null 2>&1; do \
		i=$$((i+1)); \
		if [ $$i -gt 60 ]; then \
			echo "Postgres did not become ready in time"; \
			docker logs --tail=100 english-test-postgres || true; \
			docker rm -f english-test-postgres >/dev/null 2>&1 || true; \
			exit 1; \
		fi; \
		sleep 1; \
	done
	@echo "Running postgres smoke tests..."
	@set -e; \
	trap 'docker rm -f english-test-postgres >/dev/null 2>&1 || true' EXIT; \
	GOCACHE=/tmp/go-cache \
	DATABASE_DRIVER=postgres \
	DATABASE_URL='postgres://english:english@127.0.0.1:$(PG_TEST_PORT)/english_test?sslmode=disable' \
	$(GO) test -tags=postgres -v ./internal/integration/postgres/... -count=1
	@echo "Postgres smoke tests finished."

# Postgres integration tests (Testcontainers, require Docker)
# Excludes internal/integration/llm (requires AI_URL, AI_API_KEY)
# Override with: make test-integration INTEGRATION_TEST_TIMEOUT=900
INTEGRATION_TEST_TIMEOUT ?= 600
test-integration:
	@echo "Running Postgres integration tests (Testcontainers, timeout $(INTEGRATION_TEST_TIMEOUT)s)..."
	@echo "⚠️  Requires: Docker daemon"
	$(GO) test -tags=integration -v -count=1 ./internal/integration/testkit/... ./internal/integration/user_flows/... -timeout $(INTEGRATION_TEST_TIMEOUT)s

test-integration-verbose:
	@echo "Running Postgres integration tests (verbose)..."
	@echo "⚠️  Requires: Docker daemon"
	$(GO) test -tags=integration -v -count=3 ./internal/integration/testkit/... ./internal/integration/user_flows/... -timeout 360s

# LLM integration tests (require AI_URL, AI_API_KEY env vars)
llm-words:
	@echo "Running LLM word cards integration tests..."
	@echo "⚠️  Requires: AI_URL, AI_API_KEY environment variables"
	$(GO) test -tags=integration -v -run '^TestLLM_WordCards$$' -count=1 ./internal/integration/llm/...

llm-cards:
	@echo "Running LLM training cards integration tests..."
	@echo "⚠️  Requires: AI_URL, AI_API_KEY environment variables"
	$(GO) test -tags=integration -v -run '^TestLLM_TrainingCards$$' -count=1 ./internal/integration/llm/...

llm-all: llm-words llm-cards
	@echo "✅ All LLM integration tests completed"

# Packages included in coverage (exclude cmd, integration tests and test-only helpers)
COVER_PKGS := $(shell $(GO) list ./... | grep -v '/cmd/' | grep -v 'internal/integration/' | grep -v 'internal/testutil')

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
	@echo "5. Running Go tests for coverage (excluding cmd, integration and testutil packages)..."
	@/bin/bash -c 'GOMAXPROCS=2 $(GO) test -tags=test -count=1 -p 1 -parallel 1 -timeout 30m -coverprofile=coverage.out -covermode=atomic -v $(COVER_PKGS) 2>&1 | tee .go-test-output.txt | grep -v -E "Container (created|started|ready|stopped|terminated)|Creating container|Starting container|Terminating container|Waiting for container|Waiting for Reaper|Shell not found|Reaper obtained|🐳|✅ Container|🔔 Container|⏳ Waiting|🔥 Reaper|🚫 Container|testcontainers-go -|Resolved Docker|Server Version|API Version|Operating System|Total Memory|Testcontainers for Go|Test SessionID|Test ProcessID"; exit $${PIPESTATUS[0]}'; \
	TEST_EXIT_CODE=$$?; \
	if [ $$TEST_EXIT_CODE -ne 0 ]; then \
		echo ""; \
		echo "========== FAILED TESTS =========="; \
		grep -E "^--- FAIL:" .go-test-output.txt || true; \
		echo ""; \
		echo "========== FAILURE DETAILS =========="; \
		awk '/^--- FAIL:/{p=1} p{print} /^=== RUN |^--- PASS:|^ok  /{if(p) p=0}' .go-test-output.txt | head -500; \
		rm -f .go-test-output.txt; \
		exit $$TEST_EXIT_CODE; \
	fi; \
	echo ""; \
	echo "========== TEST RESULTS =========="; \
	grep -E "^(=== RUN|--- PASS:|--- FAIL:|ok  |FAIL$$)" .go-test-output.txt || true; \
	rm -f .go-test-output.txt; \
	echo "✅ Go tests passed"
	@echo ""
	@if [ "$${CHECK_SKIP_INTEGRATION:-0}" = "1" ]; then \
		echo "5b. Skipping integration tests (quick run)"; \
	else \
		echo "5b. Running integration tests (Testcontainers)..."; \
		$(MAKE) test-integration; \
	fi
	@echo ""
	@echo "6. Running Go linter..."
	@$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1 run --timeout=3m
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

build-backfill-mastering:
	@mkdir -p bin
	$(GO) build -o bin/backfill_mastering ./cmd/backfill_mastering
	@echo "✅ Backfill tool built: bin/backfill_mastering"

backfill-mastering: build-backfill-mastering
	@echo "One-time backfill of user_word_mastering from review_events."
	@echo "Run after deploying the user_word_mastering table. DATABASE_URL must be set."
	@echo ""
	./bin/backfill_mastering

# Quick check: same as check but skips integration tests (step 5b). Use for fast feedback.
check-quick: export CHECK_SKIP_INTEGRATION := 1
check-quick: check

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
	@echo "Building webapp..."
	@cd webapp && npm run build
	@echo "✅ Webapp built. Use 'make run' to start the server."

webapp-build:
	@echo "Building webapp..."
	@cd webapp && npm run build

# Development - builds and runs both backend and frontend
dev: tidy webapp-install webapp-build grammar-bundle
	@echo "Building Go binary..."
	@$(GO) build -o bin/$(APP_NAME) ./cmd/bot
	@echo "✅ Build complete: bin/$(APP_NAME)"
	@echo ""
	@echo "Starting development environment..."
	@echo "Backend: http://localhost:8184"
	@echo "Frontend: http://localhost:8184/app"
	@echo ""
	./bin/$(APP_NAME) 

up: run

# Local development: start PostgreSQL (port 5433).
# Prefer existing container english-postgres-local (e.g. with your data); else start from docker-compose.dev.yml.
# In .env set: DATABASE_URL=postgres://english:english@127.0.0.1:5433/english?sslmode=disable
postgres-up:
	@if docker inspect english-postgres-local >/dev/null 2>&1; then \
		echo "Starting existing Postgres (english-postgres-local)..."; \
		docker compose -f docker-compose.dev.yml stop 2>/dev/null || true; \
		docker start english-postgres-local; \
		PG_CONTAINER=english-postgres-local; \
	else \
		echo "Starting PostgreSQL from docker-compose (port 5433)..."; \
		docker compose -f docker-compose.dev.yml up -d; \
		PG_CONTAINER=english-dev-postgres; \
	fi; \
	echo "Waiting for Postgres to be ready..."; \
	i=0; until docker exec $$PG_CONTAINER pg_isready -U english -d english >/dev/null 2>&1; do \
		i=$$((i+1)); \
		if [ $$i -gt 30 ]; then echo "Postgres did not become ready"; exit 1; fi; \
		sleep 1; \
	done; \
	echo "✅ Postgres is ready."

postgres-down:
	@docker stop english-postgres-local 2>/dev/null || true
	docker compose -f docker-compose.dev.yml down

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
	@echo "  make postgres-up    - Start PostgreSQL for local dev (port 5433); run before make run"
	@echo "  make postgres-down  - Stop local PostgreSQL"
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
	@echo "  make check          - Run all CI checks (tests, lint, verify, incl. integration)"
	@echo "  make check-quick    - Run CI checks without integration tests (faster)"
	@echo "  make swagger        - Generate Swagger API documentation"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make migrate-words  - Migrate existing word cards to new structured format"
	@echo "  make migrate-training-cards  - Migrate existing training cards with POS and display_word"
	@echo "  make backfill-mastering  - One-time backfill of user_word_mastering from review_events"
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
