# Personal Website

A modern, secure Go web application showcasing blog posts and portfolio projects with comprehensive testing and security-first architecture.

## ✨ Features

- **Blog System**: Markdown-based blog posts with automatic parsing
- **Portfolio Showcase**: Project gallery with detailed descriptions  
- **Server-Side Rendering**: Fast loading with Templ template engine
- **Security First**: Rate limiting, input validation, and comprehensive security headers
- **Production Ready**: Docker containerization with HTTPS support
- **Comprehensive Testing**: 84.9% coverage with 241 tests including security and performance

## 🛠️ Tech Stack

- **Backend**: Go 1.25 with Gorilla Mux router
- **Templates**: Templ for type-safe HTML templating
- **Markdown**: GoMarkdown for blog post rendering
- **Testing**: Comprehensive test suite with security focus
- **Container**: Multi-stage Docker build with security hardening

## 🚀 Quick Start

### Using Make (Recommended)

```bash
# Clone and setup
git clone https://github.com/claykom/website.git
cd website

# Install dependencies and run tests
make deps
make test

# Run the application
make run
```

### Manual Setup

```bash
# Install dependencies
go mod tidy

# Run all tests with coverage
go test -cover ./internal/config ./internal/handlers ./internal/middleware

# Build and run
go build -o website ./
./website
```

Visit http://localhost:8080 to see your site!

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build manually
make docker-build
make docker-run
```

## 📁 Project Structure

```
├── main.go                    # Application entry point
├── Makefile                   # Development workflow automation
├── content/blog/             # Markdown blog posts
├── static/                   # CSS, images, and assets
├── internal/
│   ├── config/              # Configuration management + tests
│   ├── handlers/            # HTTP request handlers + tests  
│   ├── middleware/          # Security middleware + comprehensive tests
│   ├── models/              # Data structures
│   ├── router/              # Route definitions
│   ├── testutils/           # Shared testing utilities
│   └── views/               # Templ templates
├── Dockerfile               # Container build configuration
└── docker-compose.yml       # Multi-service deployment
```

## 🧪 Testing & Quality

### Test Coverage Summary

| Package | Coverage | Tests | Focus Areas |
|---------|----------|-------|-------------|
| **Middleware** | 84.9% | 186 tests | Security, validation, performance |
| **Handlers** | 79.5% | 35 tests | HTTP endpoints, error handling |
| **Config** | 93.5% | 20 tests | Environment setup, validation |
| **Overall** | **82.8%** | **241 tests** | **Production readiness** |

### Security Testing

**Attack Vector Coverage:**
- **Path Traversal**: `/../../../etc/passwd`, URL encoding, Windows paths  
- **Input Validation**: Null bytes, Unicode normalization, buffer overflow
- **Rate Limiting**: Token bucket with 100+ concurrent request testing
- **File Security**: Extension filtering, dangerous file blocking
- **OWASP Top 10**: XSS, injection, broken access control coverage

**Performance Benchmarks:**
- Slug validation: **268ns** (0 allocations)
- Rate limiting: **756ns** per check  
- Static files: **3.2μs** with security
- Full test suite: **<3 seconds**

### Running Tests

```bash
# Make commands (recommended)
make test          # Run all tests
make coverage      # Run with coverage report
make test-race     # Detect race conditions  
make test-verbose  # Verbose test output
make bench         # Performance benchmarks
make security      # Security analysis

# Manual commands (equivalent)
go test ./...              # Run all tests
go test -cover ./...       # Run with coverage
go test -race ./...        # Race condition detection
go test -v ./...           # Verbose output
go test -bench=. ./...     # Run benchmarks
```

## 🔒 Security Features

### Built-in Security

- **Rate Limiting**: 100 req/min per IP with token bucket algorithm
- **Security Headers**: HSTS, CSP, XSS protection, content-type validation  
- **Input Validation**: Regex-based with path traversal prevention
- **File Security**: Extension allowlisting, dangerous type blocking
- **Container Security**: Non-root user, read-only filesystem

### Validated Attack Prevention

- **Directory Traversal**: `../`, URL encoding, mixed separators
- **File Upload**: PHP, executable, config file blocking  
- **Injection**: Null byte, CRLF, Unicode normalization
- **DoS Protection**: Rate limiting, content-length limits
- **Information Disclosure**: Server header removal, error sanitization

## ⚙️ Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment mode | `development` |
| `TLS_CERT_FILE` | SSL certificate path | - |
| `TLS_KEY_FILE` | SSL private key path | - |

## 🌐 API Endpoints

- `GET /` - Homepage with portfolio overview
- `GET /blog` - Blog post listing with pagination
- `GET /blog/{slug}` - Individual blog post rendering
- `GET /portfolio` - Portfolio project showcase  
- `GET /portfolio/{slug}` - Detailed project information
- `GET /health` - Health check with system status
- `GET /static/*` - Secure static file serving

## 📊 Monitoring & Observability

- **Health Checks**: Application status and dependency validation
- **Request Logging**: Structured logs with timing and status codes
- **Error Tracking**: Comprehensive error handling and reporting
- **Performance Metrics**: Response times and throughput monitoring
- **Security Events**: Rate limit violations and attack attempt logging

## 🚀 Development Workflow

### Local Development

```bash
# Format and validate code
make fmt vet

# Run comprehensive test suite
make test coverage

# Build optimized binary
make build

# Development server with hot reload
make dev
```

### Available Make Targets

```bash
# Development Commands
make help          # Show all available targets
make deps          # Download and tidy dependencies
make dev           # Run in development mode
make run           # Build and run the application

# Code Quality
make fmt           # Format Go code
make vet           # Run go vet static analysis
make lint          # Run golint (requires golint)
make security      # Run security checks

# Testing
make test          # Run all tests (uses go test ./...)
make coverage      # Run tests with coverage report
make coverage-html # Generate HTML coverage report  
make test-verbose  # Run tests with verbose output
make test-race     # Run tests with race detection
make bench         # Run benchmarks

# Build & Deploy
make build         # Build optimized binary
make clean         # Clean build artifacts

# Docker
make docker-build  # Build Docker image
make docker-run    # Run Docker container
```

## 🤝 Contributing

### Requirements

1. **Minimum 80% test coverage** for new code
2. **Security tests required** for middleware/handlers
3. **Benchmark tests** for performance-critical paths
4. **All quality checks must pass** (format, vet, security, tests)

### Workflow

```bash
# 1. Create feature branch
git checkout -b feature/amazing-feature

# 2. Develop with testing
make test           # Run tests frequently
make coverage       # Verify coverage
make security       # Check security

# 3. Ensure quality
make fmt vet        # Format and validate
make test           # Verify all tests pass

# 4. Submit PR with tests and documentation
```

### Testing Standards

- **Unit Tests**: All public functions and methods
- **Integration Tests**: HTTP endpoints with real middleware
- **Security Tests**: Attack simulation and prevention validation
- **Performance Tests**: Benchmark critical paths
- **Edge Cases**: Error conditions and boundary testing

## 📄 License

MIT License - see LICENSE file for details.

---

**Built with security, performance, and reliability in mind.** 🛡️⚡🔧