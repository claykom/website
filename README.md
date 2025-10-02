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

## 📊 Monitoring

- **Health Check**: `GET /health` returns application status
- **Logging**: Structured request logging with timestamps
- **Graceful Shutdown**: Proper signal handling for clean restarts

## 🧪 Development

```bash
# Run tests
go test ./...

# Security check
go vet ./...

# Build for production
CGO_ENABLED=0 go build -ldflags='-w -s' -o website
```

## � API Endpoints

- `GET /` - Homepage
- `GET /blog` - Blog post listing  
- `GET /blog/{slug}` - Individual blog post
- `GET /portfolio` - Portfolio projects
- `GET /portfolio/{slug}` - Project details
- `GET /health` - Health check endpoint

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Follow Go coding standards
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.