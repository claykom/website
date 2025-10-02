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
		// Clean the URL path to prevent directory traversal
		cleanPath := filepath.Clean(r.URL.Path)

		// Ensure the path doesn't go outside the root directory
		if strings.Contains(cleanPath, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Set security headers for static files
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Set cache headers for static assets
		ext := filepath.Ext(cleanPath)
		switch ext {
		case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".woff", ".woff2":
			// Cache static assets for 1 year
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		default:
			// No cache for other files
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		// Update the request URL to the cleaned path
		r.URL.Path = cleanPath

		fileServer.ServeHTTP(w, r)
	})
}
