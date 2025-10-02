package middleware

import (
	"net/http"
	"os"
)

// SecureHeaders adds security headers to responses
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS (HTTP Strict Transport Security)
		// Only add HSTS header if serving over HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content Security Policy - more restrictive and configurable
		csp := getContentSecurityPolicy()
		w.Header().Set("Content-Security-Policy", csp)

		// Additional security headers
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		// Remove server information
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")

		next.ServeHTTP(w, r)
	})
}

// getContentSecurityPolicy returns a CSP string, allowing environment override
func getContentSecurityPolicy() string {
	// Default CSP - very restrictive
	defaultCSP := "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: https:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"media-src 'self'; " +
		"object-src 'none'; " +
		"child-src 'none'; " +
		"frame-src 'none'; " +
		"worker-src 'none'; " +
		"frame-ancestors 'none'; " +
		"form-action 'self'; " +
		"base-uri 'self'; " +
		"manifest-src 'self'"

	// Allow override via environment variable for development
	if envCSP := os.Getenv("CSP_POLICY"); envCSP != "" {
		return envCSP
	}

	return defaultCSP
}
