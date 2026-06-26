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

.PHONY: all tidy build run test lint fmt setup up up-en up-es down complaints-fetch-en complaints-fetch-es complaints-cluster-en complaints-cluster-es complaints-triage-dry-en complaints-triage-dry-es complaints-journal-new complaints-dry-en complaints-apply-en complaints-dry-es complaints-apply-es complaints-dry-both complaints-apply-both complaints-improve-both complaints-plan-both complaints-prompt-autofix-en complaints-prompt-autofix-es complaints-prompt-autofix-both complaints-prompt-regression complaints-prompt-integration-es complaints-smoke-en complaints-smoke-es complaints-smoke-both complaints-quality-both complaints-quality-baseline-both complaints-regenerate-affected complaints-improve-loop-both complaints-both complaints-cycle-both complaints-loop-tests clean check check-quick ci deploy update status logs docker-build docker-run docker-stop docker-logs docker-clean docker-rebuild docker-dev docker-dev-logs docker-dev-restart webapp-install webapp-dev webapp-build test-postgres test-integration test-integration-verbose grammar-bundle grammar-bundle-list reading-cms reading-cms-stop reading-cms-build reading-covers-batch stop-llm stop-comfy postgres-dev-init-dbs clean-spanish-csv sync-spanish-word-sets prepare-english-csv sync-english-word-sets requeue-invalid-cards-es-dry requeue-invalid-cards-es requeue-invalid-cards-es-no-tts-dry requeue-invalid-cards-es-no-tts import-spanish-verbs import-spanish-verbs-jehle-bundled backfill-word-verb-links build-verb-form-examples backfill-verb-lemma-ru-glosses backfill-verb-template-links preview-verb-templates verb-training-pack-fill verb-training-sync-db help

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

# Local Reading Texts CMS (dev-only authoring UI, not part of prod runtime)
READING_CMS_PORT ?= 8791

reading-cms-stop:
	@port="$(READING_CMS_PORT)"; \
	stopped=0; \
	for pid in $$(lsof -tiTCP:$$port -sTCP:LISTEN 2>/dev/null); do \
		cmd=$$(ps -p $$pid -o command= 2>/dev/null || true); \
		if echo "$$cmd" | grep -qE 'reading_cms|cmd/reading_cms'; then \
			echo "Stopping Reading CMS on port $$port (pid $$pid)..."; \
			kill $$pid 2>/dev/null || true; \
			stopped=1; \
		else \
			echo "ERROR: port $$port in use by another process (pid $$pid):" >&2; \
			echo "  $$cmd" >&2; \
			echo "Stop it manually or set READING_CMS_PORT to a free port." >&2; \
			exit 1; \
		fi; \
	done; \
	for pid in $$(pgrep -f '[g]o run.*cmd/reading_cms' 2>/dev/null); do \
		echo "Stopping stale go run reading_cms (pid $$pid)..."; \
		kill $$pid 2>/dev/null || true; \
		stopped=1; \
	done; \
	if [ $$stopped -eq 1 ]; then sleep 0.4; fi

reading-cms: reading-cms-stop
	@$(GO) run ./cmd/reading_cms -port $(READING_CMS_PORT)

reading-cms-build:
	@$(GO) build -o bin/reading_cms ./cmd/reading_cms

stop-llm:
	@bash scripts/stop-local-llm.sh

stop-comfy:
	@bash scripts/stop-local-comfy.sh

reading-covers-batch:
	@echo "Generating reading covers for Spanish course..."
	@$(MAKE) -C courses/spanish-grammar reading-covers-batch FORCE="$(FORCE)" LIMIT="$(LIMIT)"
	@echo "Generating reading covers for English course..."
	@$(MAKE) -C courses/english-grammar reading-covers-batch FORCE="$(FORCE)" LIMIT="$(LIMIT)"
	@echo "✅ reading-covers-batch complete"

verb-training-pack-fill:
	@$(MAKE) -C courses/spanish-grammar verb-training-pack-fill

verb-training-sync-db:
	@echo "Syncing verb training JSON artifacts into DB..."
	@$(GO) run ./cmd/sync_verb_training_json --course-root courses/spanish-grammar
	@echo "✅ Verb training DB sync complete"

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
	@echo "4b. Validating reading artifacts in courses..."
	@$(MAKE) -C courses/english-grammar reading-validate
	@$(MAKE) -C courses/spanish-grammar reading-validate
	@echo "✅ Reading artifacts validation passed"
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

build-backfill-linglow-events:
	@mkdir -p bin
	$(GO) build -o bin/backfill_linglow_events ./cmd/backfill_linglow_events
	@echo "✅ Linglow event backfill tool built: bin/backfill_linglow_events"

backfill-linglow-events-audit: build-backfill-linglow-events
	@echo "Dry-run audit of legacy attempts missing in Linglow exercise_attempts."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_events

backfill-linglow-events: build-backfill-linglow-events
	@echo "Backfill legacy attempts into Linglow exercise_attempts/learning_events."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_events --commit

build-backfill-linglow-word-srs:
	@mkdir -p bin
	$(GO) build -o bin/backfill_linglow_word_srs ./cmd/backfill_linglow_word_srs
	@echo "✅ Linglow word SRS backfill tool built: bin/backfill_linglow_word_srs"

backfill-linglow-word-srs-audit: build-backfill-linglow-word-srs
	@echo "Dry-run audit of legacy user_cards missing in Linglow srs_items."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_word_srs

backfill-linglow-word-srs: build-backfill-linglow-word-srs
	@echo "Backfill legacy user_cards into Linglow srs_items."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_word_srs --commit

backfill-linglow-word-srs-resync: build-backfill-linglow-word-srs
	@echo "Resync all mapped legacy user_cards into Linglow srs_items and prune orphan due snapshots."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_word_srs --commit --resync

build-backfill-linglow-grammar-srs:
	@mkdir -p bin
	$(GO) build -o bin/backfill_linglow_grammar_srs ./cmd/backfill_linglow_grammar_srs
	@echo "✅ Linglow grammar SRS backfill tool built: bin/backfill_linglow_grammar_srs"

backfill-linglow-grammar-srs-audit: build-backfill-linglow-grammar-srs
	@echo "Dry-run audit of legacy grammar_theory_memory missing in Linglow srs_items."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_grammar_srs

backfill-linglow-grammar-srs: build-backfill-linglow-grammar-srs
	@echo "Backfill legacy grammar_theory_memory into Linglow srs_items."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_grammar_srs --commit

backfill-linglow-grammar-srs-resync: build-backfill-linglow-grammar-srs
	@echo "Resync all mapped legacy grammar_theory_memory into Linglow srs_items and prune orphan due snapshots."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_grammar_srs --commit --resync

build-backfill-linglow-attempt-srs-links:
	@mkdir -p bin
	$(GO) build -o bin/backfill_linglow_attempt_srs_links ./cmd/backfill_linglow_attempt_srs_links
	@echo "✅ Linglow attempt SRS link backfill tool built: bin/backfill_linglow_attempt_srs_links"

backfill-linglow-attempt-srs-links-audit: build-backfill-linglow-attempt-srs-links
	@echo "Dry-run audit of exercise_attempts missing srs_item_id links."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_attempt_srs_links

backfill-linglow-attempt-srs-links: build-backfill-linglow-attempt-srs-links
	@echo "Backfill exercise_attempts.srs_item_id links."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_attempt_srs_links --commit

build-backfill-linglow-media-progress:
	@mkdir -p bin
	$(GO) build -o bin/backfill_linglow_media_progress ./cmd/backfill_linglow_media_progress
	@echo "✅ Linglow media progress backfill tool built: bin/backfill_linglow_media_progress"

backfill-linglow-media-progress-audit: build-backfill-linglow-media-progress
	@echo "Dry-run audit of legacy reading/speaking progress missing in Linglow attempts/events."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_media_progress

backfill-linglow-media-progress: build-backfill-linglow-media-progress
	@echo "Backfill legacy reading/speaking progress into Linglow attempts/events."
	@echo "DATABASE_URL must be set."
	@echo ""
	./bin/backfill_linglow_media_progress --commit

build-merge-language-databases:
	@mkdir -p bin
	$(GO) build -o bin/merge_language_databases ./cmd/merge_language_databases
	@echo "✅ Linglow DB merge audit tool built: bin/merge_language_databases"

merge-language-databases-audit: build-merge-language-databases
	@echo "Dry-run audit before merging English/Spanish DBs into a unified Linglow DB."
	@echo "Set ENGLISH_DATABASE_URL, SPANISH_DATABASE_URL and optional TARGET_DATABASE_URL."
	@echo ""
	./bin/merge_language_databases

merge-language-databases-users: build-merge-language-databases
	@test -n "$$TARGET_DATABASE_URL" || (echo "Set TARGET_DATABASE_URL for write merge"; exit 1)
	./bin/merge_language_databases --commit --phase=users

merge-language-databases-user-courses: build-merge-language-databases
	@test -n "$$TARGET_DATABASE_URL" || (echo "Set TARGET_DATABASE_URL for write merge"; exit 1)
	./bin/merge_language_databases --commit --phase=user-courses

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

up-linglow: postgres-up build
	@echo ""; echo "Web app: http://localhost:8184/app (Linglow unified)"; echo ""
	SERVER_ADDRESS=:8184 SERVER_PORT=8184 ./bin/$(APP_NAME)

down:
	@echo "Stopping all local instances of $(APP_NAME)..."
	@pkill -f "$(APP_NAME)" 2>/dev/null || true
	@echo "Done."

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

# Fetch active content reports snapshot (prod or local; no LLM). Uses secrets/complaints-prod.env when present.
complaints-fetch-en:
	@set -e; \
	set -a; [ -f secrets/complaints-prod.env ] && . ./secrets/complaints-prod.env; set +a; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a; [ -f .env.en ] && . ./.env.en; set +a; \
	URL="$${COMPLAINTS_SERVICE_URL_EN:-$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:$${SERVER_PORT:-8184}}}"; \
	TOKEN="$${COMPLAINTS_SERVICE_TOKEN_EN:-$${COMPLAINTS_SERVICE_TOKEN:-}}"; \
	[ -n "$$TOKEN" ] || (echo "COMPLAINTS_SERVICE_TOKEN_EN или COMPLAINTS_SERVICE_TOKEN пустой"; exit 1); \
	COMPLAINTS_SERVICE_URL="$$URL" COMPLAINTS_SERVICE_TOKEN="$$TOKEN" \
	python3 tools-local/complaints-triage/fetch_reports.py --course en

complaints-fetch-es:
	@set -e; \
	set -a; [ -f secrets/complaints-prod.env ] && . ./secrets/complaints-prod.env; set +a; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a; [ -f .env.es ] && . ./.env.es; set +a; \
	URL="$${COMPLAINTS_SERVICE_URL_ES:-$${COMPLAINTS_SERVICE_URL:-http://127.0.0.1:$${SERVER_PORT:-8284}}}"; \
	TOKEN="$${COMPLAINTS_SERVICE_TOKEN_ES:-$${COMPLAINTS_SERVICE_TOKEN:-}}"; \
	[ -n "$$TOKEN" ] || (echo "COMPLAINTS_SERVICE_TOKEN_ES или COMPLAINTS_SERVICE_TOKEN пустой"; exit 1); \
	COMPLAINTS_SERVICE_URL="$$URL" COMPLAINTS_SERVICE_TOKEN="$$TOKEN" \
	python3 tools-local/complaints-triage/fetch_reports.py --course es

# Cluster latest EN snapshot (dry-run planning aid)
complaints-cluster-en:
	@SNAP=$$(ls -t logs/complaints/snapshot-en-*.json 2>/dev/null | head -1); \
	[ -n "$$SNAP" ] || (echo "Нет snapshot — сначала make complaints-fetch-en"; exit 1); \
	python3 tools-local/complaints-triage/cluster_reports.py "$$SNAP" | tee "logs/complaints/clusters-en-$$(date -u +%Y%m%dT%H%M%SZ).json"

complaints-cluster-es:
	@SNAP=$$(ls -t logs/complaints/snapshot-es-*.json 2>/dev/null | head -1); \
	[ -n "$$SNAP" ] || (echo "Нет snapshot — сначала make complaints-fetch-es"; exit 1); \
	python3 tools-local/complaints-triage/cluster_reports.py "$$SNAP" | tee "logs/complaints/clusters-es-$$(date -u +%Y%m%dT%H%M%SZ).json"

complaints-triage-dry-en: complaints-fetch-en complaints-cluster-en
	@echo "✅ Dry-run EN: см. logs/complaints/clusters-en-*.json и skill content-complaints-triage"

complaints-triage-dry-es: complaints-fetch-es complaints-cluster-es
	@echo "✅ Dry-run ES: см. logs/complaints/clusters-es-*.json"

# Dated triage journal in docs/complaints/ (git). Optional: make complaints-journal-new SLUG=en
complaints-journal-new:
	@python3 tools-local/complaints-triage/new_journal.py $(JOURNAL_DATE) $(SLUG)
	@echo "✅ Допиши блоки в docs/complaints/journal-*.md; см. docs/complaints/README.md"

# DEPRECATED: llama-based worker — use Cursor skill content-complaints-triage instead.
complaints-dry-en:
	@echo "⚠️  DEPRECATED: use Cursor skill .cursor/skills/content-complaints-triage or make complaints-fetch-en"
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
		--lang es \
		--csv "resources/wordsets/spanish_word_freq_pos_ud_top6000.csv" \
		--commit

prepare-english-csv:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a; [ -f .env.en ] && . ./.env.en; set +a; \
	python3 scripts/prepare_english_frequency_csv.py

sync-english-word-sets:
	@set -e; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	set -a && . ./.env.en && set +a; \
	LEARNING_PAIR=ru-en \
	LEARNING_NATIVE_LANG=ru \
	LEARNING_TARGET_LANG=en \
	LEARNING_APP_CODE=english \
	GRAMMAR_BUNDLE_ID=en \
	go run ./cmd/import_word_sets_from_csv \
		--lang en \
		--csv "resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv" \
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
	@echo "Linglow / english-ai-bot — make targets"
	@echo ""
	@echo "Quick start:"
	@echo "  make setup-local         - Initial local project setup"
	@echo "  make postgres-up         - Start local Postgres (port 5433)"
	@echo "  make up-en               - Bot+web English (.env.en), http://127.0.0.1:8184"
	@echo "  make up-es               - Bot+web Spanish (.env.es), http://127.0.0.1:8284"
	@echo "  make up-linglow          - Unified linglow_unified DB profile"
	@echo "  make down                - Stop local bot processes"
	@echo "  make dev                 - Backend dev mode (webapp build + grammar-bundle + go run)"
	@echo ""
	@echo "Build & run:"
	@echo "  make build               - webapp-build + grammar-bundle + bin/universal-ai-bot"
	@echo "  make run                 - postgres-up + build + run bot"
	@echo "  make tidy                - go mod tidy"
	@echo "  make tag                 - Bump patch git tag and push"
	@echo "  make clean               - Remove bin/ and coverage artifacts"
	@echo ""
	@echo "Webapp:"
	@echo "  make webapp-install      - npm ci in webapp/"
	@echo "  make webapp-dev          - Vite dev server"
	@echo "  make webapp-build        - Production webapp build -> webapp/dist"
	@echo ""
	@echo "Tests & CI:"
	@echo "  make test                - go test ./..."
	@echo "  make test-verbose        - go test -v ./..."
	@echo "  make test-postgres       - Start ephemeral Postgres for tests"
	@echo "  make test-integration    - Integration tests (needs DB)"
	@echo "  make check               - Full CI: webapp, tests, lint, coverage, grammar validate"
	@echo "  make check-quick         - check without integration tests"
	@echo "  make ci                  - Alias for check"
	@echo "  make fmt                 - gofmt"
	@echo "  make lint                - fmt + golangci-lint"
	@echo "  make lint-install        - Install golangci-lint locally"
	@echo "  make swagger             - Generate docs/swagger from router"
	@echo ""
	@echo "LLM integration tests (need AI_URL + AI_API_KEY in .env):"
	@echo "  make llm-words           - English word cards LLM tests"
	@echo "  make llm-cards           - English training cards LLM tests"
	@echo "  make llm-all             - llm-words + llm-cards"
	@echo "  make llm-words-es        - Spanish word cards (.env + .env.es)"
	@echo "  make llm-cards-es        - Spanish training cards (.env + .env.es)"
	@echo "  make llm-es              - llm-words-es + llm-cards-es"
	@echo ""
	@echo "Grammar bundles & training:"
	@echo "  make grammar-bundle      - Regenerate internal/grammarbundle + grammartrainingpack"
	@echo "  make grammar-bundle-list - List course dirs -> bundle ids"
	@echo "  make verb-training-pack-fill  - Fill Spanish grammar training pack (course Makefile)"
	@echo "  make verb-training-sync-db    - Sync verb training JSON into DB"
	@echo ""
	@echo "Reading texts (local dev):"
	@echo "  make reading-cms         - Start Reading CMS UI on http://127.0.0.1:8791 (restarts if already running)"
	@echo "  make reading-cms-stop    - Stop Reading CMS on READING_CMS_PORT (default 8791)"
	@echo "  make reading-cms-build   - Build bin/reading_cms"
	@echo "  make stop-llm            - Stop all local llama.cpp (llama-server) processes"
	@echo "  make stop-comfy          - Stop local ComfyUI / Comfy Desktop"
	@echo "  make reading-covers-batch - Batch cover images EN+ES (FORCE=1 LIMIT=N, needs ComfyUI)"
	@echo ""
	@echo "Word sets & Spanish verbs:"
	@echo "  make prepare-english-csv - Build English frequency CSV from wordFrequency.ods"
	@echo "  make sync-english-word-sets - Import English CSV into word sets"
	@echo "  make clean-spanish-csv   - Clean Spanish frequency CSV"
	@echo "  make sync-spanish-word-sets - Import Spanish CSV into word sets"
	@echo "  make build-spanish-gender-lexicon - Rebuild spanish_gender_lexicon.tsv"
	@echo "  make import-spanish-verbs INPUT=... [FORMAT=json|jehle-csv]"
	@echo "  make import-spanish-verbs-jehle-bundled"
	@echo "  make backfill-word-verb-links"
	@echo "  make backfill-verb-lemma-ru-glosses ARGS='--dry-run'"
	@echo "  make backfill-verb-template-links ARGS='--dry-run'"
	@echo "  make preview-verb-templates ARGS='-lemma=hablar'"
	@echo "  make backfill-noun-gender-es-dry / backfill-noun-gender-es"
	@echo "  make normalize-word-pos-es-dry / normalize-word-pos-es"
	@echo "  make requeue-invalid-cards-es-dry / requeue-invalid-cards-es"
	@echo "  make requeue-invalid-cards-es-no-tts-dry / requeue-invalid-cards-es-no-tts"
	@echo ""
	@echo "DB migrations & backfills:"
	@echo "  make migrate-words         - Migrate word cards to structured format"
	@echo "  make migrate-training-cards - Migrate training cards (POS, display_word)"
	@echo "  make backfill-mastering    - user_word_mastering from review_events"
	@echo "  make backfill-linglow-events-audit / backfill-linglow-events"
	@echo "  make backfill-linglow-word-srs-audit / backfill-linglow-word-srs / -resync"
	@echo "  make backfill-linglow-grammar-srs-audit / backfill-linglow-grammar-srs / -resync"
	@echo "  make backfill-linglow-attempt-srs-links-audit / backfill-linglow-attempt-srs-links"
	@echo "  make backfill-linglow-media-progress-audit / backfill-linglow-media-progress"
	@echo "  make merge-language-databases-audit / -users / -user-courses"
	@echo ""
	@echo "Postgres (local):"
	@echo "  make postgres-down         - Stop local Postgres container"
	@echo "  make postgres-dev-init-dbs - CREATE DATABASE spanish on local Postgres"
	@echo ""
	@echo "Content complaints (see docs/COMPLAINTS_AUTOMATION_SPEC.md):"
	@echo "  make complaints-journal-new"
	@echo "  make complaints-fetch-en / complaints-fetch-es"
	@echo "  make complaints-triage-dry-en / complaints-triage-dry-es"
	@echo "  make complaints-dry-en / complaints-apply-en"
	@echo "  make complaints-dry-es / complaints-apply-es"
	@echo "  make complaints-dry-both / complaints-apply-both"
	@echo "  make complaints-improve-both / complaints-plan-both / complaints-both"
	@echo "  make complaints-prompt-autofix-en / -es / -both"
	@echo "  make complaints-prompt-regression / complaints-prompt-integration-es"
	@echo "  make complaints-smoke-en / -es / -both"
	@echo "  make complaints-quality-both / complaints-quality-baseline-both"
	@echo "  make complaints-regenerate-affected / complaints-improve-loop-both"
	@echo "  make complaints-cycle-both / complaints-loop-tests"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build / docker-run / docker-stop / docker-logs / docker-clean"
	@echo "  make docker-rebuild / docker-dev / docker-dev-logs / docker-dev-restart"
	@echo "  make docker-deploy"
	@echo ""
	@echo "Remote deploy (systemd):"
	@echo "  make setup / deploy / update / status / logs"
