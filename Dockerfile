# Multi-stage Dockerfile for secure Go web application

# Build stage
FROM golang:1.25-alpine AS builder

# Install security updates and required packages
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Create non-root user for building
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o website .

# Runtime stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /app/website /website

# Copy static files and content
COPY --from=builder /app/static /static
COPY --from=builder /app/content /content

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/website", "healthcheck"] || exit 1

# Set security labels
LABEL \
    org.opencontainers.image.title="Website" \
    org.opencontainers.image.description="Secure Go web application" \
    org.opencontainers.image.vendor="Clay Kom" \
    org.opencontainers.image.source="https://github.com/claykom/website" \
    security.non-root="true" \
    security.no-new-privileges="true"

# Run the application
ENTRYPOINT ["/website"]