# Personal Website

A modern, secure Go web application showcasing blog posts and portfolio projects with server-side rendering using Templ templates.

## ✨ Features

- **Blog System**: Markdown-based blog posts with automatic parsing
- **Portfolio Showcase**: Project gallery with detailed descriptions  
- **Server-Side Rendering**: Fast loading with Templ template engine
- **Security First**: Rate limiting, input validation, and comprehensive security headers
- **Production Ready**: Docker containerization with HTTPS support
- **Health Monitoring**: Built-in health check endpoint for monitoring

## 🛠️ Tech Stack

- **Backend**: Go 1.25 with Gorilla Mux router
- **Templates**: Templ for type-safe HTML templating
- **Markdown**: GoMarkdown for blog post rendering
- **Container**: Multi-stage Docker build
- **Proxy**: Nginx reverse proxy configuration
- **Security**: Custom middleware for headers, rate limiting, and validation

## 🚀 Quick Start

### Local Development

```bash
# Clone the repository
git clone https://github.com/claykom/website.git
cd website

# Install dependencies
go mod tidy

# Copy environment template
cp .env.example .env

# Run the application
go run main.go
```

Visit http://localhost:8080 to see your site!

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build manually
docker build -t website .
docker run -p 8080:8080 website
```

### Production Deployment

```bash
# Deploy with HTTPS support
docker run -d \
  --name website \
  -p 443:443 \
  -v /path/to/certs:/certs:ro \
  -e TLS_CERT_FILE=/certs/cert.pem \
  -e TLS_KEY_FILE=/certs/key.pem \
  -e ENV=production \
  website
```

## 📁 Project Structure

```
├── main.go                 # Application entry point
├── content/blog/          # Markdown blog posts
├── static/               # CSS, images, and assets
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # Security and logging middleware
│   ├── models/         # Data structures
│   ├── router/         # Route definitions
│   └── views/          # Templ templates
├── Dockerfile           # Container build configuration
└── docker-compose.yml   # Multi-service deployment
```

## ⚙️ Configuration

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment mode | `development` |
| `TLS_CERT_FILE` | SSL certificate path | - |
| `TLS_KEY_FILE` | SSL private key path | - |

## 🔒 Security Features

- **Rate Limiting**: 100 requests per minute per IP
- **Security Headers**: HSTS, CSP, XSS protection, and more
- **Input Validation**: Protection against path traversal and injection
- **Container Security**: Non-root user and read-only filesystem
- **HTTPS Support**: TLS encryption for production deployments

## 🧪 Testing

### Test Coverage

The application includes comprehensive test suites with **84.9% coverage** for critical middleware:

| Package | Coverage | Tests | Description |
|---------|----------|-------|-------------|
| **Middleware** | 84.9% | 186 tests | Security, validation, static file handling |
| **Handlers** | 79.5% | 35 tests | HTTP request handling, responses |
| **Config** | 93.5% | 20 tests | Environment variables, validation |

### Running Tests

```bash
# Run all tests
go test ./internal/config ./internal/handlers ./internal/middleware

# Run with coverage report  
go test -cover ./internal/config ./internal/handlers ./internal/middleware

# Run specific package tests
go test ./internal/middleware

# Run with verbose output
go test -v ./internal/middleware

# Generate detailed coverage report
go test -coverprofile=config_coverage.out ./internal/config
go test -coverprofile=handlers_coverage.out ./internal/handlers  
go test -coverprofile=middleware_coverage.out ./internal/middleware
go tool cover -html=middleware_coverage.out
```

### Test Categories

#### Security Testing
- **Path Traversal Protection**: Tests against `../` attacks, URL encoding, double encoding
- **Input Validation**: Slug validation, filename sanitization, content-type checking
- **Rate Limiting**: Token bucket algorithm, concurrent access, IP-based limiting
- **Static File Security**: Extension filtering, header validation, path sanitization

#### Edge Cases & Error Handling
- **Malformed Requests**: Invalid URLs, Unicode attacks, null byte injection
- **Concurrent Access**: 100+ simultaneous requests, thread safety validation
- **Memory Protection**: Large payload handling, content-length limits
- **File System Security**: Dangerous file types, symbolic links, directory traversal

#### Performance Testing
- **Benchmark Tests**: Validation performance, middleware overhead measurement
- **Stress Testing**: High-concurrency scenarios, memory usage validation
- **Cache Testing**: Static file caching, header optimization

### Test Structure

```
internal/
├── middleware/
│   ├── validation_test.go      # Input validation (71 tests)
│   ├── static_test.go          # Static file handling (45 tests)
│   ├── error_edge_cases_test.go # Error handling (46 tests)
│   ├── headers_test.go         # Security headers (12 tests)
│   └── ratelimit_test.go       # Rate limiting (12 tests)
├── handlers/
│   ├── handlers_test.go        # HTTP handlers (25 tests)
│   └── blog_portfolio_test.go  # Content handling (10 tests)
└── testutils/
    └── testutils.go           # Shared testing utilities
```

### Security Test Examples

```go
// Path traversal protection
func TestPathTraversal(t *testing.T) {
    attacks := []string{
        "/../../../etc/passwd",
        "/%2E%2E%2F%2E%2E%2F",
        "/..\\windows\\system32",
    }
    // All attacks should be blocked
}

// Concurrent validation safety
func TestConcurrentValidation(t *testing.T) {
    // 100 simultaneous requests
    // Validates thread safety
}
```

## 📊 Monitoring

- **Health Check**: `GET /health` returns application status
- **Logging**: Structured request logging with timestamps
- **Graceful Shutdown**: Proper signal handling for clean restarts

## 🧪 Development

```bash
# Run all tests
go test ./internal/config ./internal/handlers ./internal/middleware

# Run tests with coverage
go test -cover ./internal/config ./internal/handlers ./internal/middleware

# Run tests with race detection
go test -race ./internal/config ./internal/handlers ./internal/middleware

# Run specific test package
go test ./internal/middleware

# Run security analysis
go vet ./...

# Format code
go fmt ./...

# Build for production
CGO_ENABLED=0 go build -ldflags='-w -s' -o website

# Generate test coverage report  
go test -coverprofile=middleware_coverage.out ./internal/middleware
go tool cover -html=middleware_coverage.out -o coverage.html
```

## � API Endpoints

- `GET /` - Homepage
- `GET /blog` - Blog post listing  
- `GET /blog/{slug}` - Individual blog post
- `GET /portfolio` - Portfolio projects
- `GET /portfolio/{slug}` - Project details
- `GET /health` - Health check endpoint

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Write tests** for new functionality (minimum 80% coverage required)
4. **Run the test suite** (`go test ./...`) and ensure all tests pass
5. **Follow Go standards** (`go fmt`, `go vet`, `golint`)
6. **Test security** if adding middleware or handlers
7. **Update documentation** including README if needed
8. **Submit** a pull request with a clear description

### Testing Guidelines

- **Security First**: All security-related code must have comprehensive tests
- **Coverage Target**: Maintain or improve existing coverage percentages
- **Edge Cases**: Include tests for error conditions and edge cases  
- **Concurrency**: Test concurrent access patterns where applicable
- **Performance**: Add benchmark tests for performance-critical code

## 📄 License

MIT License - see LICENSE file for details.