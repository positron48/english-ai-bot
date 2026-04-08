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
COPY --from=builder /app/prompts ./prompts
# Ship static Spanish frequency CSV for in-cluster imports (independent from grammar submodule).
COPY --from=builder /app/resources/wordsets/spanish_word_freq_pos_ud_top6000.csv ./data/spanish_word_freq_pos_ud_top6000.csv
COPY --from=builder /app/resources/wordsets/spanish_gender_lexicon.tsv ./data/spanish_gender_lexicon.tsv

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
