#!/bin/bash

# Security-focused build and deployment script
# Usage: ./deploy.sh [build|run|deploy]

set -euo pipefail

# Configuration
APP_NAME="website"
DOCKER_IMAGE="${APP_NAME}:latest"
CONTAINER_NAME="${APP_NAME}-container"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Security checks
security_check() {
    log_info "Running security checks..."
    
    # Check for .env file in production
    if [[ -f .env ]] && [[ "${ENV:-development}" == "production" ]]; then
        log_warn ".env file found in production environment"
    fi
    
    # Check Go version
    GO_VERSION=$(go version | grep -o 'go[0-9.]*' | sed 's/go//')
    MIN_VERSION="1.25"
    if [[ "$(printf '%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]]; then
        log_error "Go version $GO_VERSION is below minimum required $MIN_VERSION"
        exit 1
    fi
    
    # Run Go security checks
    log_info "Running go vet..."
    go vet ./...
    
    log_info "Checking for vulnerabilities..."
    if command -v govulncheck &> /dev/null; then
        govulncheck ./...
    else
        log_warn "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"
    fi
    
    log_info "Security checks completed"
}

# Build function
build_app() {
    log_info "Building application..."
    
    # Security checks first
    security_check
    
    # Clean build
    go clean
    go mod tidy
    
    # Build with security flags
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags='-w -s -extldflags "-static"' \
        -a -installsuffix cgo \
        -o "${APP_NAME}" .
    
    log_info "Build completed successfully"
}

# Docker build function
build_docker() {
    log_info "Building Docker image..."
    
    # Build multi-stage Docker image
    docker build \
        --no-cache \
        --pull \
        -t "${DOCKER_IMAGE}" \
        .
    
    # Security scan if available
    if command -v docker &> /dev/null; then
        log_info "Running Docker security scan..."
        docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
            -v $(pwd):/app aquasec/trivy:latest image "${DOCKER_IMAGE}" || log_warn "Trivy not available"
    fi
    
    log_info "Docker image built successfully"
}

# Run function
run_app() {
    log_info "Starting application..."
    
    # Stop existing container if running
    docker stop "${CONTAINER_NAME}" 2>/dev/null || true
    docker rm "${CONTAINER_NAME}" 2>/dev/null || true
    
    # Run with security options
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --restart unless-stopped \
        --security-opt no-new-privileges:true \
        --read-only \
        --tmpfs /tmp:noexec,nosuid,size=100m \
        -p 8080:8080 \
        -e ENV=production \
        -e HOST=0.0.0.0 \
        -e PORT=8080 \
        "${DOCKER_IMAGE}"
    
    log_info "Application started successfully"
    log_info "Health check: http://localhost:8080/health"
}

# Deploy with docker-compose
deploy_compose() {
    log_info "Deploying with docker-compose..."
    
    # Build and start services
    docker-compose build --no-cache
    docker-compose up -d
    
    # Wait for health check
    log_info "Waiting for application to be healthy..."
    sleep 10
    
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_info "Deployment successful - application is healthy"
    else
        log_error "Deployment failed - application is not responding"
        docker-compose logs
        exit 1
    fi
}

# Main script logic
case "${1:-help}" in
    "build")
        build_app
        ;;
    "docker")
        build_docker
        ;;
    "run")
        build_docker
        run_app
        ;;
    "deploy")
        deploy_compose
        ;;
    "security")
        security_check
        ;;
    "help"|*)
        echo "Usage: $0 {build|docker|run|deploy|security}"
        echo ""
        echo "Commands:"
        echo "  build    - Build the Go application with security flags"
        echo "  docker   - Build Docker image with security scanning"
        echo "  run      - Build and run the application in Docker"
        echo "  deploy   - Deploy using docker-compose"
        echo "  security - Run security checks only"
        exit 1
        ;;
esac