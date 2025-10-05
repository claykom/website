package middleware

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/claykom/website/internal/testutils"
)

// TestErrorHandlingEdgeCases tests various error conditions and edge cases
func TestErrorHandlingEdgeCases(t *testing.T) {
	validator := NewValidator()

	t.Run("malformed requests", func(t *testing.T) {
		middleware := InputValidation(validator)
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := middleware(testHandler)

		tests := []struct {
			name           string
			path           string
			expectedStatus int
		}{
			{"URL encoded path traversal", "/?slug=%2E%2E%2F", http.StatusBadRequest}, // Validation catches ../
			{"Double URL encoding", "/?slug=%252E%252E%252F", http.StatusBadRequest},  // Still catches pattern
			{"Unicode normalization attack", "/?slug=test\u2044admin", http.StatusBadRequest},
			{"Very long parameter", "/?slug=" + strings.Repeat("a", 200), http.StatusBadRequest},
			{"Space in slug", "/?slug=test+admin", http.StatusBadRequest},
			{"Query injection", "/?slug=test", http.StatusOK}, // Only looks at slug parameter
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := testutils.NewTestRequest("GET", tt.path, "")
				rr := testutils.NewTestResponseRecorder()

				handler.ServeHTTP(rr, req)

				if rr.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
				}
			})
		}
	})

	t.Run("concurrent validation", func(t *testing.T) {
		middleware := InputValidation(validator)
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := middleware(testHandler)

		// Test concurrent access to validation doesn't cause races
		done := make(chan bool, 100)
		for i := 0; i < 100; i++ {
			go func(i int) {
				defer func() { done <- true }()

				path := "/?slug=test" + string(rune(i%26+'a'))
				req := testutils.NewTestRequest("GET", path, "")
				rr := testutils.NewTestResponseRecorder()
				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("Concurrent request %d failed: %d", i, rr.Code)
				}
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})

	t.Run("memory exhaustion protection", func(t *testing.T) {
		middleware := InputValidation(validator)
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := middleware(testHandler)

		// Test with maximum allowed content length
		req := testutils.NewTestRequest("POST", "/", strings.Repeat("x", 1024))
		req.ContentLength = 10*1024*1024 - 1 // Just under limit
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200 for content under limit, got %d", rr.Code)
		}
	})

	t.Run("regex edge cases", func(t *testing.T) {
		// Test slugs that might break regex processing
		edgeCases := []string{
			"test-",         // ends with hyphen (valid)
			"_test",         // starts with underscore (valid)
			"123test",       // starts with number (valid)
			"Test_Case-123", // mixed case and chars (valid)
			"",              // empty (invalid but handled gracefully)
			"test..test",    // double dots (invalid)
			"test/test",     // slash (invalid)
			"test\\test",    // backslash (invalid)
			"test test",     // space (invalid)
			"test\ttest",    // tab (invalid)
			"test\ntest",    // newline (invalid)
		}

		for _, slug := range edgeCases {
			t.Run("slug: "+slug, func(t *testing.T) {
				// This should not panic or crash
				result := validator.ValidateSlug(slug)
				// We don't care about the result, just that it doesn't crash
				_ = result
			})
		}
	})
}

// TestStaticHandlerErrorCases tests error conditions in static file handling
func TestStaticHandlerErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	handler := SecureStaticHandler(http.Dir(tempDir))

	t.Run("malicious file access attempts", func(t *testing.T) {
		maliciousPaths := []string{
			"/../../../../../etc/passwd",
			"/..\\..\\..\\windows\\system32\\config\\sam",
			"/%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			"/%252e%252e%252f%252e%252e%252f%252e%252e%252fetc%252fpasswd",
			"/\x2e\x2e/\x2e\x2e/etc/passwd",
			"/..\\/..\\/../etc/passwd",
			"/subdir/../../../sensitive.txt",
			"/file.txt/../../etc/passwd",
		}

		for _, maliciousPath := range maliciousPaths {
			t.Run("path: "+maliciousPath, func(t *testing.T) {
				req := testutils.NewTestRequest("GET", maliciousPath, "")
				rr := testutils.NewTestResponseRecorder()

				handler.ServeHTTP(rr, req)

				// Should be blocked (forbidden or not found)
				if rr.Code != http.StatusForbidden && rr.Code != http.StatusNotFound {
					t.Errorf("Expected 403 or 404 for %s, got %d", maliciousPath, rr.Code)
				}
			})
		}
	})

	t.Run("dangerous file types", func(t *testing.T) {
		dangerousFiles := []string{
			"/malware.exe",
			"/script.php",
			"/config.conf",
			"/database.db",
			"/private.key",
			"/certificate.pem",
			"/script.sh",
			"/batch.bat",
			"/code.py",
			"/source.go",
		}

		for _, file := range dangerousFiles {
			t.Run("file: "+file, func(t *testing.T) {
				req := testutils.NewTestRequest("GET", file, "")
				rr := testutils.NewTestResponseRecorder()

				handler.ServeHTTP(rr, req)

				// Should be forbidden
				if rr.Code != http.StatusForbidden {
					t.Errorf("Expected 403 Forbidden for %s, got %d", file, rr.Code)
				}
			})
		}
	})

	t.Run("stress test with concurrent requests", func(t *testing.T) {
		// Create a test file
		testFile := tempDir + "/stress.css"
		content := ".stress { color: red; }"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		const numRequests = 50
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(i int) {
				defer func() { done <- true }()

				req := testutils.NewTestRequest("GET", "/stress.css", "")
				rr := testutils.NewTestResponseRecorder()
				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("Concurrent request %d failed: %d", i, rr.Code)
				}
			}(i)
		}

		for i := 0; i < numRequests; i++ {
			<-done
		}
	})

	t.Run("error response consistency", func(t *testing.T) {
		// Test that error responses are consistent and don't leak information
		req := testutils.NewTestRequest("GET", "/../../../etc/passwd", "")
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected 403, got %d", rr.Code)
		}

		body := rr.Body.String()
		if body != "Forbidden\n" {
			t.Errorf("Expected 'Forbidden\\n', got %q", body)
		}

		// Ensure no sensitive headers are leaked
		serverHeader := rr.Header().Get("Server")
		if serverHeader != "" {
			t.Errorf("Server header should be empty, got: %s", serverHeader)
		}
	})
}

// TestValidationPanic tests that validation doesn't panic on edge cases
func TestValidationPanic(t *testing.T) {
	validator := NewValidator()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Validation panicked: %v", r)
		}
	}()

	// Test various inputs that might cause panics
	testCases := []string{
		"",
		strings.Repeat("a", 1000),
		"\x00\x01\x02",
		"test\u0000admin",
		"../../../",
		"test\r\ninjection",
		"\xFF\xFE\x00\x00",
		"test\xc0\x80", // overlong encoding
	}

	for _, tc := range testCases {
		validator.ValidateSlug(tc)
		validator.SanitizeFilename(tc)
		validator.ValidateContentType(tc, []string{"text/plain"})
	}
}
