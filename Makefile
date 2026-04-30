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

.PHONY: all tidy build run test lint fmt setup up up-en up-es complaints-dry-en complaints-apply-en complaints-dry-es complaints-apply-es complaints-dry-both complaints-apply-both complaints-improve-both complaints-plan-both complaints-prompt-autofix-en complaints-prompt-autofix-es complaints-prompt-autofix-both complaints-prompt-regression complaints-prompt-integration-es complaints-smoke-en complaints-smoke-es complaints-smoke-both complaints-quality-both complaints-quality-baseline-both complaints-regenerate-affected complaints-improve-loop-both complaints-both complaints-cycle-both complaints-loop-tests clean check check-quick ci deploy update status logs docker-build docker-run docker-stop docker-logs docker-clean docker-rebuild docker-dev docker-dev-logs docker-dev-restart webapp-install webapp-dev webapp-build test-postgres test-integration test-integration-verbose grammar-bundle grammar-bundle-list postgres-dev-init-dbs clean-spanish-csv sync-spanish-word-sets requeue-invalid-cards-es-dry requeue-invalid-cards-es requeue-invalid-cards-es-no-tts-dry requeue-invalid-cards-es-no-tts import-spanish-verbs import-spanish-verbs-jehle-bundled backfill-word-verb-links build-verb-form-examples backfill-verb-lemma-ru-glosses backfill-verb-template-links preview-verb-templates

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

# Regenerate internal/grammarbundle/* and internal/grammartrainingpack/* from courses/* (bundle + training pack per bundle.target)
grammar-bundle:
	@echo "Updating grammar course repos from git (without switching pinned submodule commits)..."
	@for d in courses/*; do \
		if [ ! -d "$$d" ]; then continue; fi; \
		if git -C "$$d" rev-parse --is-inside-work-tree >/dev/null 2>&1; then \
			branch=$$(git -C "$$d" symbolic-ref --short HEAD 2>/dev/null || true); \
			echo "-> $$d"; \
			git -C "$$d" fetch --all --prune --quiet || true; \
			if [ -n "$$branch" ]; then \
				git -C "$$d" pull --ff-only origin "$$branch" --quiet || true; \
			else \
				echo "   (detached HEAD, fetch-only)"; \
			fi; \
		fi; \
	done
	@echo ""
	@echo "Generating embedded grammar bundles (see: ./scripts/generate-grammar-bundle.sh list)..."
	@./scripts/generate-grammar-bundle.sh
	@echo "✅ Grammar bundles generated"
	@echo ""
	@echo "Copying embedded grammar training packs (courses/*/training_pack -> internal/grammartrainingpack)..."
	@./scripts/generate-grammar-training-pack.sh
	@echo "✅ Grammar training packs generated"

grammar-bundle-list:
	@./scripts/generate-grammar-bundle.sh list

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
	/bin/bash -c 'set -o pipefail; $(GO) test -tags=postgres -json ./internal/integration/postgres/... -count=1 2>&1 | python3 scripts/go-test-compact.py'
	@echo "Postgres smoke tests finished."

# Postgres integration tests (Testcontainers, require Docker)
# Excludes internal/integration/llm (requires AI_URL, AI_API_KEY)
# Override with: make test-integration INTEGRATION_TEST_TIMEOUT=900
INTEGRATION_TEST_TIMEOUT ?= 600
test-integration:
	@echo "Running Postgres integration tests (Testcontainers, timeout $(INTEGRATION_TEST_TIMEOUT)s)..."
	@echo "⚠️  Requires: Docker daemon"
	@/bin/bash -c 'set -o pipefail; $(GO) test -tags=integration -json -count=1 ./internal/integration/testkit/... ./internal/integration/user_flows/... -timeout $(INTEGRATION_TEST_TIMEOUT)s 2>&1 | python3 scripts/go-test-compact.py'

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

llm-words-es:
	@echo "Running Spanish LLM word cards integration tests..."
	@echo "⚠️  Requires: .env.es (and AI_URL, AI_API_KEY)"
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) test -tags=integration -v -run '^TestLLM_WordCards_ES$$' -count=1 ./internal/integration/llm/...

llm-cards-es:
	@echo "Running Spanish LLM training cards integration tests..."
	@echo "⚠️  Requires: .env.es (and AI_URL, AI_API_KEY)"
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) test -tags=integration -v -run '^TestLLM_TrainingCards_ES$$' -count=1 ./internal/integration/llm/...

llm-es: llm-words-es llm-cards-es
	@echo "✅ Spanish LLM prompt regression tests completed"

# Packages included in coverage (exclude cmd, integration tests and test-only helpers)
COVER_PKGS := $(shell $(GO) list ./... | grep -v '/cmd/' | grep -v 'internal/integration/' | grep -v 'internal/testutil')

# CI checks (same as in GitHub Actions)
check:
	@echo "=== Running CI Checks ==="
	@echo ""
	@echo "1. Checking webapp dependencies..."
	@cd webapp && npm ci --prefer-offline --no-audit --no-fund > /dev/null 2>&1 || npm install --no-audit --no-fund
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
	@echo "5. Running Go tests for coverage (compact: . ok | E fail | F panic/build | S skip)..."
	@rm -f coverage.out .go-test-output.jsonl
	@/bin/bash -c 'set -o pipefail; $(GO) test -tags=test -count=1 -p 3 -timeout 30m -coverprofile=coverage.out -covermode=atomic -json $(COVER_PKGS) 2>&1 | tee .go-test-output.jsonl | python3 scripts/go-test-compact.py'; \
	TEST_EXIT_CODE=$$?; \
	if [ $$TEST_EXIT_CODE -ne 0 ]; then \
		echo ""; \
		echo "ℹ️  Raw JSON stream: .go-test-output.jsonl"; \
		exit $$TEST_EXIT_CODE; \
	fi; \
	rm -f .go-test-output.jsonl; \
	echo ""; \
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

build-backfill-noun-gender:
	@mkdir -p bin
	$(GO) build -o bin/backfill_noun_gender ./cmd/backfill_noun_gender
	@echo "✅ Backfill tool built: bin/backfill_noun_gender"

build-normalize-word-pos:
	@mkdir -p bin
	$(GO) build -o bin/normalize_word_pos ./cmd/normalize_word_pos
	@echo "✅ POS normalization tool built: bin/normalize_word_pos"

# Local Spanish profile backfill for noun_gender.
# Imports vars from optional .env, then required .env.es.
# Tunables: BATCH (default 100), LIMIT (default 0 = all).
backfill-noun-gender-es-dry: build-backfill-noun-gender
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	BATCH="$${BATCH:-100}"; \
	LIMIT="$${LIMIT:-0}"; \
	echo "Running DRY backfill_noun_gender with BATCH=$$BATCH LIMIT=$$LIMIT"; \
	./bin/backfill_noun_gender -dry-run=true -batch "$$BATCH" -limit "$$LIMIT"

backfill-noun-gender-es: build-backfill-noun-gender
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	BATCH="$${BATCH:-100}"; \
	LIMIT="$${LIMIT:-0}"; \
	echo "Running WRITE backfill_noun_gender with BATCH=$$BATCH LIMIT=$$LIMIT"; \
	./bin/backfill_noun_gender -dry-run=false -batch "$$BATCH" -limit "$$LIMIT"

normalize-word-pos-es-dry: build-normalize-word-pos
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	BATCH="$${BATCH:-200}"; \
	LIMIT="$${LIMIT:-0}"; \
	echo "Running DRY normalize_word_pos with BATCH=$$BATCH LIMIT=$$LIMIT"; \
	./bin/normalize_word_pos -dry-run=true -batch "$$BATCH" -limit "$$LIMIT"

normalize-word-pos-es: build-normalize-word-pos
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	BATCH="$${BATCH:-200}"; \
	LIMIT="$${LIMIT:-0}"; \
	echo "Running WRITE normalize_word_pos with BATCH=$$BATCH LIMIT=$$LIMIT"; \
	./bin/normalize_word_pos -dry-run=false -batch "$$BATCH" -limit "$$LIMIT"

import-spanish-verbs:
	@if [ -z "$(INPUT)" ]; then \
		echo "Usage: make import-spanish-verbs INPUT=/path/to/file [FORMAT=json|jehle-csv] [SOURCE=open-data] [VERSION=v1]"; \
		echo "Bundled Jehle CSV: make import-spanish-verbs-jehle-bundled"; \
		echo "Uses optional .env then required .env.es (DATABASE_URL), like other Spanish maintenance targets."; \
		exit 1; \
	fi
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/import_spanish_verbs --input "$(INPUT)" --format "$${FORMAT:-json}" --source "$${SOURCE:-open-data}" --source-version "$${VERSION:-v1}"

import-spanish-verbs-jehle-bundled:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/import_spanish_verbs \
		--input resources/verbs/jehle_verb_database.csv \
		--format jehle-csv \
		--source fred-jehle-ghidinelli \
		--source-version jehle-csv-sha256-f77f01d536cd351584051d76902ff8051ab1b945a38e69c7ed02da78ab082ea8 && \
	$(GO) run ./cmd/import_spanish_verbs \
		--input resources/verbs/jehle_supplement_aux_haber.csv \
		--format jehle-csv \
		--source project-supplement \
		--source-version haber-paradigm-v1

backfill-word-verb-links:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/backfill_word_verb_links

build-verb-form-examples:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/build_verb_form_examples

# Batch LLM: fill verb_lemmas.metadata_json ru.gloss for Spanish lemmas (min requests: ~ceil(N/batch-size)). Needs AI_* from .env / .env.es.
backfill-verb-lemma-ru-glosses:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/backfill_verb_lemma_ru_glosses -- $(ARGS)

# Offline merge: verb_class + allowed_template_ids for curated lemmas (see cmd/backfill_verb_template_links).
backfill-verb-template-links:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/backfill_verb_template_links -- $(ARGS)

# Preview ES/RU example lines for every stored paradigm form (runtime templates + DB catalog). Needs .env.es + imported verb_forms_dict.
preview-verb-templates:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните DATABASE_URL"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	$(GO) run ./cmd/preview_verb_templates -- $(ARGS)

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
	@echo "Installing golangci-lint v2.10.1 into ./bin..."
	@mkdir -p bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v2.10.1

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

# Second DB on the same local Postgres (port 5433) for running Spanish instance alongside English.
postgres-dev-init-dbs:
	@set -e; \
	for c in english-dev-postgres english-postgres-local; do \
	  if docker inspect $$c >/dev/null 2>&1; then \
	    docker exec $$c psql -U english -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='spanish'" 2>/dev/null | grep -q 1 \
	      || docker exec $$c psql -U english -d postgres -c "CREATE DATABASE spanish;"; \
	    echo "✅ Database 'spanish' ready (container $$c)"; \
	    exit 0; \
	  fi; \
	done; \
	echo "⚠️  Local postgres container not found — start with: make postgres-up"; \
	exit 0

# Local RU→EN: copy env.example.en → .env.en (and optionally merge secrets from .env)
up-en: postgres-up build
	@test -f .env.en || (echo "Нет .env.en — скопируйте env.example.en в .env.en и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.en && set +a; \
	PORT=$${SERVER_PORT:-}; \
	ADDR=$${SERVER_ADDRESS:-}; \
	if [ -n "$$PORT" ]; then APP_URL="http://localhost:$$PORT"; \
	elif [ -n "$$ADDR" ]; then APP_URL="http://localhost$${ADDR}"; \
	else APP_URL="http://localhost:8184"; fi; \
	echo ""; \
	echo "========================================"; \
	echo "🚀 APP URL (EN): $$APP_URL"; \
	echo "========================================"; \
	echo ""; \
	exec ./bin/$(APP_NAME)

# Local RU→ES: отдельный порт HTTP и БД spanish (make postgres-dev-init-dbs после первого postgres-up)
up-es: postgres-up postgres-dev-init-dbs build
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните секреты"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	PORT=$${SERVER_PORT:-}; \
	ADDR=$${SERVER_ADDRESS:-}; \
	if [ -n "$$PORT" ]; then APP_URL="http://localhost:$$PORT"; \
	elif [ -n "$$ADDR" ]; then APP_URL="http://localhost$${ADDR}"; \
	else APP_URL="http://localhost:8284"; fi; \
	echo ""; \
	echo "========================================"; \
	echo "🚀 APP URL (ES): $$APP_URL"; \
	echo "========================================"; \
	echo ""; \
	exec ./bin/$(APP_NAME)

# Local complaints worker (English profile)
complaints-dry-en:
	@test -f .env.en || (echo "Нет .env.en — скопируйте env.example.en в .env.en и заполните COMPLAINTS_SERVICE_TOKEN"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.en && set +a; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	URL="$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:$${SERVER_PORT:-8184}}"; \
	TOKEN="$${COMPLAINTS_SERVICE_TOKEN:-}"; \
	[ -n "$$TOKEN" ] || (echo "COMPLAINTS_SERVICE_TOKEN пустой"; exit 1); \
	COMPLAINTS_SERVICE_URL="$$URL" COMPLAINTS_SERVICE_TOKEN="$$TOKEN" COURSE_SCOPE=english \
	python3 tools-local/complaints-worker/worker.py

complaints-apply-en:
	@test -f .env.en || (echo "Нет .env.en — скопируйте env.example.en в .env.en и заполните COMPLAINTS_SERVICE_TOKEN"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.en && set +a; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	URL="$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:$${SERVER_PORT:-8184}}"; \
	TOKEN="$${COMPLAINTS_SERVICE_TOKEN:-}"; \
	[ -n "$$TOKEN" ] || (echo "COMPLAINTS_SERVICE_TOKEN пустой"; exit 1); \
	COMPLAINTS_SERVICE_URL="$$URL" COMPLAINTS_SERVICE_TOKEN="$$TOKEN" COURSE_SCOPE=english \
	python3 tools-local/complaints-worker/worker.py --apply

# Local complaints worker (Spanish profile)
complaints-dry-es:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните COMPLAINTS_SERVICE_TOKEN"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	URL="$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:$${SERVER_PORT:-8284}}"; \
	TOKEN="$${COMPLAINTS_SERVICE_TOKEN:-}"; \
	[ -n "$$TOKEN" ] || (echo "COMPLAINTS_SERVICE_TOKEN пустой"; exit 1); \
	COMPLAINTS_SERVICE_URL="$$URL" COMPLAINTS_SERVICE_TOKEN="$$TOKEN" COURSE_SCOPE=spanish \
	python3 tools-local/complaints-worker/worker.py

complaints-apply-es:
	@test -f .env.es || (echo "Нет .env.es — скопируйте env.example.es в .env.es и заполните COMPLAINTS_SERVICE_TOKEN"; exit 1)
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	URL="$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:$${SERVER_PORT:-8284}}"; \
	TOKEN="$${COMPLAINTS_SERVICE_TOKEN:-}"; \
	[ -n "$$TOKEN" ] || (echo "COMPLAINTS_SERVICE_TOKEN пустой"; exit 1); \
	COMPLAINTS_SERVICE_URL="$$URL" COMPLAINTS_SERVICE_TOKEN="$$TOKEN" COURSE_SCOPE=spanish \
	python3 tools-local/complaints-worker/worker.py --apply

# Run both profiles sequentially (requires EN/ES urls; token can be shared)
complaints-dry-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	URL_EN="$${COMPLAINTS_SERVICE_URL_EN:-$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:8184}}"; \
	URL_ES="$${COMPLAINTS_SERVICE_URL_ES:-$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:8284}}"; \
	TOKEN_EN="$${COMPLAINTS_SERVICE_TOKEN_EN:-$${COMPLAINTS_SERVICE_TOKEN:-}}"; \
	TOKEN_ES="$${COMPLAINTS_SERVICE_TOKEN_ES:-$${COMPLAINTS_SERVICE_TOKEN:-}}"; \
	[ -n "$$TOKEN_EN" ] || (echo "COMPLAINTS_SERVICE_TOKEN_EN или COMPLAINTS_SERVICE_TOKEN обязателен"; exit 1); \
	[ -n "$$TOKEN_ES" ] || (echo "COMPLAINTS_SERVICE_TOKEN_ES или COMPLAINTS_SERVICE_TOKEN обязателен"; exit 1); \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	COMPLAINTS_SERVICE_URL="$$URL_EN" COMPLAINTS_SERVICE_TOKEN="$$TOKEN_EN" COURSE_SCOPE=english \
	python3 tools-local/complaints-worker/worker.py; \
	COMPLAINTS_SERVICE_URL="$$URL_ES" COMPLAINTS_SERVICE_TOKEN="$$TOKEN_ES" COURSE_SCOPE=spanish \
	python3 tools-local/complaints-worker/worker.py

complaints-apply-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	URL_EN="$${COMPLAINTS_SERVICE_URL_EN:-$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:8184}}"; \
	URL_ES="$${COMPLAINTS_SERVICE_URL_ES:-$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:8284}}"; \
	TOKEN_EN="$${COMPLAINTS_SERVICE_TOKEN_EN:-$${COMPLAINTS_SERVICE_TOKEN:-}}"; \
	TOKEN_ES="$${COMPLAINTS_SERVICE_TOKEN_ES:-$${COMPLAINTS_SERVICE_TOKEN:-}}"; \
	[ -n "$$TOKEN_EN" ] || (echo "COMPLAINTS_SERVICE_TOKEN_EN или COMPLAINTS_SERVICE_TOKEN обязателен"; exit 1); \
	[ -n "$$TOKEN_ES" ] || (echo "COMPLAINTS_SERVICE_TOKEN_ES или COMPLAINTS_SERVICE_TOKEN обязателен"; exit 1); \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	COMPLAINTS_SERVICE_URL="$$URL_EN" COMPLAINTS_SERVICE_TOKEN="$$TOKEN_EN" COURSE_SCOPE=english \
	python3 tools-local/complaints-worker/worker.py --apply; \
	COMPLAINTS_SERVICE_URL="$$URL_ES" COMPLAINTS_SERVICE_TOKEN="$$TOKEN_ES" COURSE_SCOPE=spanish \
	python3 tools-local/complaints-worker/worker.py --apply

complaints-improve-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/analyze-journal.py

complaints-plan-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/analyze-journal.py

complaints-prompt-autofix-en:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/apply-prompt-improvements.py --course english

complaints-prompt-autofix-es:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/apply-prompt-improvements.py --course spanish

complaints-prompt-autofix-both: complaints-prompt-autofix-en complaints-prompt-autofix-es
	@echo "✅ complaints-prompt-autofix-both done"

complaints-prompt-regression:
	@python3 tools-local/complaints-worker/prompt-validator-regression.py

complaints-smoke-en:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/prompt-llm-integration-smoke.py --course english

complaints-smoke-es:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/prompt-llm-integration-smoke.py --course spanish

complaints-smoke-both: complaints-smoke-en complaints-smoke-es
	@echo "✅ complaints-smoke-both done"

complaints-quality-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/quality-regression.py

complaints-quality-baseline-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/quality-regression.py --set-baseline

complaints-regenerate-affected:
	@python3 tools-local/complaints-worker/regenerate-affected-chapters.py

complaints-prompt-integration-es:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/prompt-llm-integration-smoke.py

complaints-loop-tests:
	@python3 tools-local/complaints-worker/tests/test_improve_prompt_loop.py

complaints-improve-loop-both:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	if [ -f .env.en ]; then set -a; . ./.env.en; set +a; fi; \
	if [ -f .env.es ]; then set -a; . ./.env.es; set +a; fi; \
	bash tools-local/complaints-worker/ensure-llama.sh; \
	python3 tools-local/complaints-worker/improve-prompt-loop.py

# Full automated complaints loop entrypoint:
# 1) apply complaints cleanup + resolve
# 2) LLM analysis of journal and actionable improvement plan
complaints-both: complaints-apply-both complaints-improve-both
	@echo "✅ complaints-both done"
	@echo "Next: run training-pack-fill in affected course(s), then make grammar-bundle, make check, commit/tag."

# End-to-end automated cycle for Spanish complaints:
# apply complaints -> improve journal -> auto-fix prompt -> regression test -> regenerate pack -> rebuild bundles
complaints-cycle-both: complaints-both complaints-prompt-autofix-es complaints-prompt-regression
	@$(MAKE) -C courses/spanish-grammar training-pack-fill
	@$(MAKE) grammar-bundle
	@echo "✅ complaints-cycle-both done"
	@echo "Next: make check, then manual commit and make tag."

clean-spanish-csv:
	@python3 scripts/clean_spanish_frequency_csv.py \
		resources/wordsets/spanish_word_freq_pos_ud_top6000.csv \
		courses/spanish-grammar/spanish_word_freq_pos_ud_top6000.csv

sync-spanish-word-sets:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.es && set +a; \
	LEARNING_PAIR=ru-es \
	LEARNING_NATIVE_LANG=ru \
	LEARNING_TARGET_LANG=es \
	LEARNING_APP_CODE=spanish \
	GRAMMAR_BUNDLE_ID=es \
	go run ./cmd/import_word_sets_from_csv \
		--csv "resources/wordsets/spanish_word_freq_pos_ud_top6000.csv" \
		--commit

requeue-invalid-cards-es-dry:
	@bash scripts/requeue_invalid_training_cards.sh --target-lang es

requeue-invalid-cards-es:
	@bash scripts/requeue_invalid_training_cards.sh --target-lang es --commit

requeue-invalid-cards-es-no-tts-dry:
	@bash scripts/requeue_invalid_training_cards.sh --target-lang es --no-tts

requeue-invalid-cards-es-no-tts:
	@bash scripts/requeue_invalid_training_cards.sh --target-lang es --no-tts --commit

build-spanish-gender-lexicon:
	@python3 scripts/build_spanish_gender_lexicon.py --download

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
	@echo "  make postgres-dev-init-dbs - CREATE DATABASE spanish on local Postgres (for make up-es)"
	@echo "  make postgres-down  - Stop local PostgreSQL"
	@echo "  make grammar-bundle - Regenerate internal/grammarbundle/* and internal/grammartrainingpack/* from courses/*/bundle.target"
	@echo "  make grammar-bundle-list - List course dirs -> bundle ids"
	@echo "  make up-en          - Run bot+web with .env.en (+ optional .env), http :8184, DB english"
	@echo "  make up-es          - Run second instance with .env.es (+ optional .env), http :8284, DB spanish"
	@echo "  make complaints-dry-en   - Dry-run complaints worker for English profile (.env + .env.en)"
	@echo "  make complaints-apply-en - Apply complaints worker for English profile (.env + .env.en)"
	@echo "  make complaints-dry-es   - Dry-run complaints worker for Spanish profile (.env + .env.es)"
	@echo "  make complaints-apply-es - Apply complaints worker for Spanish profile (.env + .env.es)"
	@echo "  make complaints-dry-both - Dry-run for EN and ES URLs sequentially (COMPLAINTS_SERVICE_URL_EN/ES)"
	@echo "  make complaints-apply-both - Apply for EN and ES URLs sequentially (COMPLAINTS_SERVICE_URL_EN/ES)"
	@echo "  make complaints-improve-both - Analyze latest complaints journal via LLM and build improvement plan"
	@echo "  make complaints-plan-both - Build/update improvement plan from latest complaints journal"
	@echo "  make complaints-both - Full loop: apply complaints + auto improvement analysis"
	@echo "  make complaints-prompt-autofix-es - Auto-update ES generator prompt from latest improvement plan"
	@echo "  make complaints-prompt-autofix-en - Auto-update EN generator prompt from latest improvement plan"
	@echo "  make complaints-prompt-autofix-both - Auto-update EN+ES prompts from latest improvement plan"
	@echo "  make complaints-prompt-regression - Regression checks for prompt + validator compatibility"
	@echo "  make complaints-prompt-integration-es - Real LLM smoke: generate 1 block and ensure validation passes"
	@echo "  make complaints-smoke-en - Real LLM smoke for English prompt and validator"
	@echo "  make complaints-smoke-es - Real LLM smoke for Spanish prompt and validator"
	@echo "  make complaints-smoke-both - Real LLM smoke for EN + ES"
	@echo "  make complaints-quality-both - Multi-scenario LLM quality regression with baseline comparison"
	@echo "  make complaints-quality-baseline-both - Save current quality run as baseline"
	@echo "  make complaints-regenerate-affected - Targeted fill-training-pack for chapters from latest changed-theory-blocks"
	@echo "  make complaints-loop-tests - Contract tests for iterative improve loop helpers"
	@echo "  make complaints-improve-loop-both - Autonomous strict iterative prompt-improve loop (max 3 iterations)"
	@echo "  make complaints-cycle-both - Full automated cycle through regen + grammar-bundle"
	@echo "  make build-spanish-gender-lexicon - Rebuild resources/wordsets/spanish_gender_lexicon.tsv from online source"
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
	@echo "  make backfill-noun-gender-es-dry - Dry-run noun_gender backfill with .env.es (+ optional .env)"
	@echo "  make backfill-noun-gender-es - Write noun_gender backfill with .env.es (+ optional .env)"
	@echo "  make normalize-word-pos-es-dry - Dry-run POS normalization with .env.es (+ optional .env)"
	@echo "  make normalize-word-pos-es - Write POS normalization with .env.es (+ optional .env)"
	@echo "  make import-spanish-verbs INPUT=... [FORMAT=json|jehle-csv] - Import Spanish verb forms"
	@echo "  make import-spanish-verbs-jehle-bundled - Import bundled Fred Jehle CSV (resources/verbs)"
	@echo "  make backfill-word-verb-links - Link existing Spanish verb word_cards with verb_lemmas"
	@echo "  make build-verb-form-examples - No-op (cloze ES/RU lines are generated at training time; keeps Makefile compatibility)"
	@echo "  make backfill-verb-lemma-ru-glosses ARGS='--dry-run' — LLM batch for RU glosses in verb_lemmas.metadata_json (optional ARGS; add ARGS='--fill-class' for verb_class only)"
	@echo "  make backfill-verb-template-links ARGS='--dry-run' — offline verb_class + allowed_template_ids for curated lemmas (Spanish DB)"
	@echo "  make preview-verb-templates ARGS='-lemma=hablar' — dump ES/RU example for each paradigm row (Spanish DB)"
	@echo "  make requeue-invalid-cards-es-dry - Dry-run soft cleanup: invalid cards + duplicates + invalid TTS"
	@echo "  make requeue-invalid-cards-es - Commit soft cleanup: invalid cards + duplicates + invalid TTS"
	@echo "  make requeue-invalid-cards-es-no-tts-dry - Dry-run soft cleanup without touching TTS"
	@echo "  make requeue-invalid-cards-es-no-tts - Commit soft cleanup without touching TTS"
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
