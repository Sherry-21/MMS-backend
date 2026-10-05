# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Leverage layer caching by copying go module files first
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy application source code
COPY . .

# Build the main API server binary (statically linked, stripped)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/api \
    ./cmd/api

# Build the standalone migration binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate \
    ./cmd/migrate

# ==============================================================================
# Final Minimal Runtime Stage
# ==============================================================================
FROM alpine:3.21 AS runner

# Install essential runtime tools (ca-certificates for SSL, tzdata, and curl for health check)
RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

# Copy binaries from the build stage
COPY --from=builder /app/bin/api /usr/local/bin/api
COPY --from=builder /app/bin/migrate /usr/local/bin/migrate

# Set executable permissions
RUN chmod +x /usr/local/bin/api /usr/local/bin/migrate

# Use unprivileged non-root user
USER appuser:appgroup

# Expose default HTTP server port
EXPOSE 8080

# Health check using the /health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Start the API service
ENTRYPOINT ["api"]
