package middleware

import (
	"net/http"
	"strings"
	"testing"

	"github.com/claykom/website/internal/testutils"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Error("NewValidator should return a non-nil validator")
	}
	if validator.slugRegex == nil {
		t.Error("Validator should have initialized regex")
	}
}

func TestValidateSlug(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		slug     string
		expected bool
	}{
		// Valid slugs
		{"valid simple slug", "hello", true},
		{"valid slug with hyphens", "hello-world", true},
		{"valid slug with underscores", "hello_world", true},
		{"valid slug with numbers", "post123", true},
		{"valid mixed", "test-post_123", true},

		// Invalid slugs - empty/length
		{"empty slug", "", false},
		{"too long slug", strings.Repeat("a", 101), false},

		// Invalid slugs - path traversal
		{"path traversal dots", "test..test", false},
		{"path traversal slash", "test/admin", false},
		{"path traversal backslash", "test\\admin", false},
		{"just dots", "..", false},
		{"dots at start", "..test", false},
		{"dots at end", "test..", false},

		// Invalid slugs - special characters
		{"spaces", "hello world", false},
		{"special chars", "test@example", false},
		{"query params", "test?id=1", false},
		{"hash", "test#section", false},
		{"brackets", "test[1]", false},
		{"parentheses", "test(1)", false},
		{"percent encoding", "test%20", false},

		// Edge cases
		{"single char", "a", true},
		{"single number", "1", true},
		{"single hyphen", "-", true},
		{"single underscore", "_", true},
		{"all caps", "TEST", true},
		{"mixed case", "TeSt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateSlug(tt.slug)
			if result != tt.expected {
				t.Errorf("ValidateSlug(%q) = %v, want %v", tt.slug, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple filename", "test.txt", "test.txt"},
		{"path traversal", "../../../etc/passwd", "passwd"},
		{"with directory", "dir/file.txt", "file.txt"},
		{"windows path", "C:\\Windows\\file.txt", "file.txt"},
		{"mixed separators", "dir\\../file.txt", "file.txt"},
		{"multiple dots", "../../file.txt", "file.txt"},
		{"absolute path", "/etc/passwd", "passwd"},
		{"empty string", "", ""},
		{"just dots", "..", ""},
		{"hidden file", ".htaccess", ".htaccess"},
		{"complex traversal", "dir/../../../admin/config.php", "config.php"},
		{"null bytes", "file\x00.txt", "file\x00.txt"}, // Note: filepath.Base handles this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateContentType(t *testing.T) {
	validator := NewValidator()

	allowedTypes := []string{"text/html", "application/json", "image/"}

	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		// Valid content types
		{"exact match html", "text/html", true},
		{"exact match json", "application/json", true},
		{"image prefix match", "image/png", true},
		{"image with params", "image/jpeg; charset=utf-8", true},

		// Case insensitive
		{"uppercase", "TEXT/HTML", true},
		{"mixed case", "Application/JSON", true},

		// With extra whitespace
		{"with spaces", " text/html ", true},
		{"with tabs", "\timage/png\t", true},

		// Invalid types
		{"not allowed", "text/plain", false},
		{"partial match", "text/htm", false},
		{"empty string", "", false},
		{"wrong prefix", "video/mp4", false},

		// Edge cases
		{"just type", "image", false}, // Needs the slash
		{"malformed", "image/", true}, // Prefix match works
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateContentType(tt.contentType, allowedTypes)
			if result != tt.expected {
				t.Errorf("ValidateContentType(%q) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestInputValidationMiddleware(t *testing.T) {
	validator := NewValidator()
	middleware := InputValidation(validator)

	tests := []struct {
		name           string
		queryParams    map[string]string
		contentLength  int64
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "no validation needed",
			queryParams:    map[string]string{},
			contentLength:  0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid slug",
			queryParams:    map[string]string{"slug": "valid-post"},
			contentLength:  0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid slug with traversal",
			queryParams:    map[string]string{"slug": "../admin"},
			contentLength:  0,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid slug parameter",
		},
		{
			name:           "invalid slug empty",
			queryParams:    map[string]string{"slug": ""},
			contentLength:  0,
			expectedStatus: http.StatusOK, // Empty slug is not checked by middleware
		},
		{
			name:           "invalid slug special chars",
			queryParams:    map[string]string{"slug": "test@example"},
			contentLength:  0,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid slug parameter",
		},
		{
			name:           "content too large",
			queryParams:    map[string]string{},
			contentLength:  11 * 1024 * 1024, // 11MB
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectedBody:   "Request too large",
		},
		{
			name:           "content at limit",
			queryParams:    map[string]string{},
			contentLength:  10 * 1024 * 1024, // 10MB exactly
			expectedStatus: http.StatusOK,
		},
		{
			name:           "multiple issues - slug takes precedence",
			queryParams:    map[string]string{"slug": "../bad"},
			contentLength:  11 * 1024 * 1024,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid slug parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Wrap with middleware
			handler := middleware(testHandler)

			// Create request with query parameters
			req := testutils.NewTestRequest("GET", "/", "")
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()
			req.ContentLength = tt.contentLength

			rr := testutils.NewTestResponseRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check body if error expected
			if tt.expectedBody != "" {
				body := rr.Body.String()
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body)
				}
			}
		})
	}
}

// Edge case tests
func TestValidationEdgeCases(t *testing.T) {
	t.Run("nil validator", func(t *testing.T) {
		// Test graceful handling if validator is nil (defensive programming)
		middleware := InputValidation(nil)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(testHandler)
		req := testutils.NewTestRequest("GET", "/?slug=test", "")
		rr := testutils.NewTestResponseRecorder()

		// This should panic or handle gracefully
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with nil validator")
			}
		}()

		handler.ServeHTTP(rr, req)
	})

	t.Run("regex compilation failure simulation", func(t *testing.T) {
		// This tests that our regex is valid
		validator := NewValidator()
		if validator.slugRegex == nil {
			t.Error("Regex should compile successfully")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		// Test that validator is thread-safe
		validator := NewValidator()
		middleware := InputValidation(validator)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(testHandler)

		// Run multiple concurrent requests
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				req := testutils.NewTestRequest("GET", "/?slug=test-concurrent", "")
				rr := testutils.NewTestResponseRecorder()
				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("Concurrent request failed with status %d", rr.Code)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Benchmark tests
func BenchmarkValidateSlug(b *testing.B) {
	validator := NewValidator()
	testSlugs := []string{"valid-slug", "test123", "invalid../path", "test@example"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateSlug(testSlugs[i%len(testSlugs)])
	}
}

func BenchmarkInputValidationMiddleware(b *testing.B) {
	validator := NewValidator()
	middleware := InputValidation(validator)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)
	req := testutils.NewTestRequest("GET", "/?slug=test-post", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		handler.ServeHTTP(rr, req)
	}
}
