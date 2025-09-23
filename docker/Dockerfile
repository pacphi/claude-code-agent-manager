# Multi-stage build for security and minimal image size
# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Create non-root user for build
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Verify dependencies
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o agent-manager \
    cmd/agent-manager/main.go

# Final stage - minimal runtime image
FROM scratch

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy user from builder
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary from builder
COPY --from=builder /build/agent-manager /usr/local/bin/agent-manager

# Create necessary directories with proper permissions
WORKDIR /app

# Use non-root user
USER appuser

# Set security labels
LABEL maintainer="Chris Phillipson" \
      version="1.0" \
      description="Secure Agent Manager for Claude Code subagents" \
      security.scan="true" \
      security.nonroot="true"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/agent-manager", "version"]

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/agent-manager"]

# Default command (can be overridden)
CMD ["--help"]