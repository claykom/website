package middleware

import (
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateInput provides input validation utilities
type ValidateInput struct {
	slugRegex *regexp.Regexp
}

// NewValidator creates a new input validator
func NewValidator() *ValidateInput {
	return &ValidateInput{
		// Allow alphanumeric characters, hyphens, underscores
		slugRegex: regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`),
	}
}

// ValidateSlug validates URL slugs to prevent path traversal and injection
func (v *ValidateInput) ValidateSlug(slug string) bool {
	if slug == "" || len(slug) > 100 {
		return false
	}

	// Check for path traversal attempts
	if strings.Contains(slug, "..") || strings.Contains(slug, "/") || strings.Contains(slug, "\\") {
		return false
	}

	return v.slugRegex.MatchString(slug)
}

// SanitizeFilename sanitizes filenames to prevent directory traversal
func (v *ValidateInput) SanitizeFilename(filename string) string {
	// Clean the path and get just the base filename
	cleaned := filepath.Base(path.Clean("/" + filename))

	// Remove any remaining path separators
	cleaned = strings.ReplaceAll(cleaned, "/", "")
	cleaned = strings.ReplaceAll(cleaned, "\\", "")

	return cleaned
}

// ValidateContentType validates that the content type is allowed
func (v *ValidateInput) ValidateContentType(contentType string, allowedTypes []string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	for _, allowed := range allowedTypes {
		if strings.HasPrefix(contentType, strings.ToLower(allowed)) {
			return true
		}
	}

	return false
}

// InputValidation middleware to validate common input parameters
func InputValidation(validator *ValidateInput) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate URL parameters if they exist
			if slug := r.URL.Query().Get("slug"); slug != "" {
				if !validator.ValidateSlug(slug) {
					http.Error(w, "Invalid slug parameter", http.StatusBadRequest)
					return
				}
			}

			// Validate Content-Length to prevent large payloads
			if r.ContentLength > 10*1024*1024 { // 10MB limit
				http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
