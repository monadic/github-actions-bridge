# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN make build-bridge

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    docker-cli \
    git \
    curl \
    bash

# Create non-root user
RUN addgroup -g 1000 actions && \
    adduser -u 1000 -G actions -D actions

# Create directories
RUN mkdir -p /var/lib/actions-bridge && \
    chown -R actions:actions /var/lib/actions-bridge

# Copy binary from builder
COPY --from=builder /build/bin/actions-bridge /usr/local/bin/actions-bridge

# Switch to non-root user
USER actions

# Set environment variables
ENV ACTIONS_BRIDGE_BASE_DIR=/var/lib/actions-bridge \
    ACT_DEFAULT_IMAGE=catthehacker/ubuntu:act-22.04 \
    ACT_PLATFORM=linux/amd64

# Expose health check port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the bridge
ENTRYPOINT ["/usr/local/bin/actions-bridge"]