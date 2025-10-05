package middleware

import (
	"net/http"
	"path/filepath"
	"strings"
)

// SecureStaticHandler creates a secure static file handler that prevents directory traversal
func SecureStaticHandler(root http.Dir) http.Handler {
	fileServer := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enhanced path traversal protection
		path := r.URL.Path

		// Check for various path traversal patterns
		if strings.Contains(path, "..") ||
			strings.Contains(path, "\\") ||
			strings.Contains(path, "\x00") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Clean the path and ensure it's safe
		cleanPath := filepath.Clean(path)
		// Ensure the cleaned path doesn't try to escape the directory
		if strings.Contains(cleanPath, "..") || cleanPath == "." {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Restrict to allowed file extensions for security
		ext := strings.ToLower(filepath.Ext(path))
		allowedExtensions := map[string]bool{
			".css":   true,
			".js":    true,
			".png":   true,
			".jpg":   true,
			".jpeg":  true,
			".gif":   true,
			".ico":   true,
			".svg":   true,
			".woff":  true,
			".woff2": true,
			".webp":  true,
			".avif":  true,
		}

		if ext != "" && !allowedExtensions[ext] {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Set comprehensive security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Set appropriate Content-Type headers
		switch ext {
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		// Set cache headers for static assets
		switch ext {
		case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".woff", ".woff2", ".webp", ".avif":
			// Cache static assets for 1 year
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			w.Header().Set("Expires", "Thu, 31 Dec 2026 23:59:59 GMT")
		default:
			// No cache for other files
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		// Serve the file
		fileServer.ServeHTTP(w, r)
	})
}
