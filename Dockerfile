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

# Build backfill tool for one-time mastering score backfill
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill_mastering ./cmd/backfill_mastering

# Build word-sets import tooling for one-time/k3s maintenance runs
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o import_word_sets_from_csv ./cmd/import_word_sets_from_csv
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o fill_missing_set_pos_cards ./cmd/fill_missing_set_pos_cards
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o revalidate_training_cards ./cmd/revalidate_training_cards
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill_noun_gender ./cmd/backfill_noun_gender
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o normalize_word_pos ./cmd/normalize_word_pos
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o import_spanish_verbs ./cmd/import_spanish_verbs
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill_word_verb_links ./cmd/backfill_word_verb_links
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o build_verb_form_examples ./cmd/build_verb_form_examples
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill_verb_lemma_ru_glosses ./cmd/backfill_verb_lemma_ru_glosses
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill_verb_template_links ./cmd/backfill_verb_template_links
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o preview_verb_templates ./cmd/preview_verb_templates
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o sync_verb_training_json ./cmd/sync_verb_training_json

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

# Copy binaries from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/backfill_mastering .
COPY --from=builder /app/import_word_sets_from_csv .
COPY --from=builder /app/fill_missing_set_pos_cards .
COPY --from=builder /app/revalidate_training_cards .
COPY --from=builder /app/backfill_noun_gender .
COPY --from=builder /app/normalize_word_pos .
COPY --from=builder /app/import_spanish_verbs .
COPY --from=builder /app/backfill_word_verb_links .
COPY --from=builder /app/build_verb_form_examples .
COPY --from=builder /app/backfill_verb_lemma_ru_glosses .
COPY --from=builder /app/backfill_verb_template_links .
COPY --from=builder /app/preview_verb_templates .
COPY --from=builder /app/sync_verb_training_json .
COPY --from=builder /app/scripts/requeue_invalid_training_cards.sh ./scripts/requeue_invalid_training_cards.sh
COPY --from=builder /app/prompts ./prompts
# Ship static Spanish frequency CSV for in-cluster imports (independent from grammar submodule).
COPY --from=builder /app/resources/wordsets/spanish_word_freq_pos_ud_top6000.csv ./data/spanish_word_freq_pos_ud_top6000.csv
COPY --from=builder /app/resources/wordsets/spanish_gender_lexicon.tsv ./data/spanish_gender_lexicon.tsv
# Fred Jehle Spanish verb paradigms (CC BY-NC-SA 3.0) — see resources/verbs/ATTRIBUTION.txt
COPY --from=builder /app/resources/verbs/jehle_verb_database.csv ./data/verbs/jehle_verb_database.csv
COPY --from=builder /app/resources/verbs/jehle_supplement_aux_haber.csv ./data/verbs/jehle_supplement_aux_haber.csv
COPY --from=builder /app/resources/verbs/ATTRIBUTION.txt ./data/verbs/ATTRIBUTION.txt
COPY --from=builder /app/resources/verbs/SUPPLEMENT_HABER.txt ./data/verbs/SUPPLEMENT_HABER.txt
COPY --from=builder /app/courses/spanish-grammar/training_pack/verb_forms ./courses/spanish-grammar/training_pack/verb_forms

RUN chmod +x /app/scripts/requeue_invalid_training_cards.sh

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
