# Multi-stage build
# Stage 1: Builder
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o devitri ./cmd/devitri

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates and wget for HTTPS + health checks; su-exec to drop privileges
RUN apk --no-cache add ca-certificates wget su-exec

# Create a non-root user
RUN adduser -D -s /bin/sh devitri

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/devitri .
# Copy migration files
COPY internal/db/migrations ./internal/db/migrations

# Change ownership to non-root user
RUN chown -R devitri:devitri /app

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Entrypoint runs as root to fix volume permissions, then exec as devitri
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]