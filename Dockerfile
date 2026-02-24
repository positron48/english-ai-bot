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

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS (Postgres driver is compiled in)
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create data directory for database
RUN mkdir -p /app/data && chown -R appuser:appgroup /app/data

# Copy binaries from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/backfill_mastering .
COPY --from=builder /app/prompts ./prompts

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
