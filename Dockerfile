# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata nodejs npm

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install webapp dependencies first for better Docker layer caching
COPY webapp/package.json webapp/package-lock.json ./webapp/
RUN cd webapp && npm ci

# Copy source code
COPY . .

# Build webapp dist that is embedded into Go binary.
RUN cd webapp && npm run build

# Build the application (CGO disabled, fast incremental-compatible build)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/bot

# Build only the maintenance binaries actually invoked in-cluster.
# Keep import_learning_content available for DB-first content refreshes after a new bundle
# is rolled out; it embeds grammar/reading assets and must run against the deployed image.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o merge_language_databases ./cmd/merge_language_databases
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dedup_word_forms ./cmd/dedup_word_forms
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o import_learning_content ./cmd/import_learning_content

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS; ffmpeg for OpenRouter TTS (PCM→MP3)
RUN apk --no-cache add ca-certificates tzdata ffmpeg

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create data directories (DB + cached pronunciation audio)
RUN mkdir -p /app/data/tts && chown -R appuser:appgroup /app/data

# Copy binaries from builder stage (main + in-cluster job/cronjob tools only)
COPY --from=builder /app/main .
COPY --from=builder /app/merge_language_databases .
COPY --from=builder /app/dedup_word_forms .
COPY --from=builder /app/import_learning_content .
COPY --from=builder /app/scripts/requeue_invalid_training_cards.sh ./scripts/requeue_invalid_training_cards.sh
COPY --from=builder /app/prompts ./prompts
# Ship static Spanish frequency CSV for in-cluster imports (independent from grammar submodule).
COPY --from=builder /app/resources/wordsets/spanish_word_freq_pos_ud_top6000.csv ./data/spanish_word_freq_pos_ud_top6000.csv
COPY --from=builder /app/resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv ./data/english_word_freq_pos_ud_top6000.filtered.csv
COPY --from=builder /app/resources/wordsets/english_word_sets_must_have.yaml ./data/english_word_sets_must_have.yaml
COPY --from=builder /app/resources/wordsets/spanish_gender_lexicon.tsv ./data/spanish_gender_lexicon.tsv
# Fred Jehle Spanish verb paradigms (CC BY-NC-SA 3.0) — see resources/verbs/ATTRIBUTION.txt
COPY --from=builder /app/resources/verbs/jehle_verb_database.csv ./data/verbs/jehle_verb_database.csv
COPY --from=builder /app/resources/verbs/jehle_supplement_aux_haber.csv ./data/verbs/jehle_supplement_aux_haber.csv
COPY --from=builder /app/resources/verbs/ATTRIBUTION.txt ./data/verbs/ATTRIBUTION.txt
COPY --from=builder /app/resources/verbs/SUPPLEMENT_HABER.txt ./data/verbs/SUPPLEMENT_HABER.txt
COPY --from=builder /app/internal/grammartrainingpack/es/verb_forms ./internal/grammartrainingpack/es/verb_forms

RUN chmod +x /app/scripts/requeue_invalid_training_cards.sh

# Note: no recursive `chown -R /app` — it would duplicate every binary into a new layer
# (~500MB of waste). Binaries/data are world-readable; only /app/data (chowned above) is
# written at runtime.

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
