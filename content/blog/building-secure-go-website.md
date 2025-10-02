---
title: Building a Secure Go Website with Modern Best Practices
slug: building-secure-go-website
author: Clayton
date: 2025-10-01
tags: [go, security, web, docker, production]
excerpt: A comprehensive guide to building a production-ready Go web application with security-first design
---

# Building a Secure Go Website with Modern Best Practices

In this post, I'll walk you through the process of building a modern, secure Go web application that's ready for production deployment. This project demonstrates how to implement security best practices, containerization, and modern Go development patterns.

## Project Overview

This website is built as a personal portfolio and blog platform using Go 1.25, showcasing:

- **Security-First Design**: Comprehensive security headers, rate limiting, and input validation
- **Modern Architecture**: Clean separation of concerns with internal package structure
- **Production Ready**: Docker containerization with multi-stage builds
- **Performance**: Server-side rendering with Templ templates
- **Monitoring**: Health checks and structured logging

## Key Technologies

- **Go 1.25**: Latest Go version with improved performance
- **Gorilla Mux**: Flexible HTTP routing and middleware
- **Templ**: Type-safe HTML templating
- **Docker**: Multi-stage containerization
- **Nginx**: Reverse proxy with additional security

## Security Implementation

Security was a primary focus throughout development. Here are the key security measures implemented:

### 1. Comprehensive Security Headers

```go
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // HSTS for HTTPS enforcement
        if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        }
        
        // Content Security Policy
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'...")
        
        // Additional security headers
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        next.ServeHTTP(w, r)
    })
}
```

### 2. Rate Limiting

Implemented a token bucket rate limiter to prevent abuse:

```go
// Rate limit: 100 requests per minute per IP
rateLimitStore := middleware.NewRateLimitStore(5 * time.Minute)
r.Use(middleware.RateLimit(rateLimitStore, 100, time.Minute))
```

### 3. Input Validation

All user inputs are validated using custom middleware:

```go
func (v *ValidateInput) ValidateSlug(slug string) bool {
    if slug == "" || len(slug) > 100 {
        return false
    }
    
    // Prevent path traversal
    if strings.Contains(slug, "..") || strings.Contains(slug, "/") {
        return false
    }
    
    return v.slugRegex.MatchString(slug)
}
```

## Architecture Decisions

### Project Structure

The project follows Go best practices with a clean internal package structure:

```
internal/
├── config/          # Environment-based configuration
├── handlers/        # HTTP request handlers
├── middleware/      # Security and logging middleware
├── models/         # Data structures
├── router/         # Route definitions and setup
└── views/          # Templ templates
```

### Configuration Management

All configuration is environment-based with sensible defaults:

```go
type Config struct {
    Server ServerConfig
    TLS    TLSConfig
    App    AppConfig
}
```

Environment variables are used for all deployment-specific settings, with a comprehensive `.env.example` file provided.

## Container Security

The Docker configuration uses security best practices:

### Multi-Stage Build

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
# ... build steps

# Runtime stage  
FROM scratch
COPY --from=builder /app/website /website
USER appuser
```

### Security Features

- **Non-root user**: Application runs as unprivileged user
- **Read-only filesystem**: Container filesystem is read-only
- **Minimal attack surface**: Using scratch base image
- **Security options**: `no-new-privileges` flag enabled

## Performance Optimizations

### Server-Side Rendering

Using Templ for type-safe, compiled templates:

```go
func BlogPost(post models.BlogPost) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // Compiled template rendering
    })
}
```

### Efficient Static File Serving

Custom static file handler with security and caching:

```go
func SecureStaticHandler(root http.Dir) http.Handler {
    // Path traversal protection
    // Appropriate cache headers
    // Security headers for static files
}
```

## Deployment Strategy

The application supports multiple deployment scenarios:

1. **Local Development**: Simple `go run main.go`
2. **Docker**: Single container deployment
3. **Docker Compose**: Multi-service with nginx proxy
4. **Production**: HTTPS with proper certificate management

### Health Monitoring

Comprehensive health check endpoint:

```go
func Health(w http.ResponseWriter, r *http.Request) {
    response := map[string]interface{}{
        "status":    "ok",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "version":   getVersion(),
        "uptime":    getUptime(),
    }
    respondWithJSON(w, http.StatusOK, response)
}
```

## Lessons Learned

1. **Security by Default**: Implementing security from the start is easier than retrofitting
2. **Environment Configuration**: Externalized config makes deployment flexible
3. **Container Security**: Multi-stage builds significantly reduce attack surface
4. **Testing**: Comprehensive testing including security validation is essential
5. **Documentation**: Good documentation aids both development and operations

## Future Enhancements

Potential improvements for future versions:

- Database integration for dynamic content
- Authentication and user management
- Advanced monitoring and metrics
- CDN integration for global performance
- Automated testing in CI/CD pipeline

## Conclusion

This project demonstrates that building secure, production-ready Go applications doesn't require compromising on development velocity. By implementing security best practices from the start and using modern tooling, we can create applications that are both robust and maintainable.

The complete source code is available on [GitHub](https://github.com/claykom/website), showcasing all the security implementations, configuration management, and deployment strategies discussed in this post.

Building secure web applications is an ongoing process, and this project serves as a foundation for future development while maintaining high security standards.