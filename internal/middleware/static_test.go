package middleware

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/claykom/website/internal/testutils"
)

func TestSecureStaticHandler(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"test.css":    ".body { color: red; }",
		"script.js":   "console.log('test');",
		"image.png":   "fake png content",
		"doc.pdf":     "fake pdf content",
		"unsafe.php":  "<?php echo 'hacked'; ?>",
		"config.conf": "secret=password",
	}

	// Create subdirectories and files
	subDir := filepath.Join(tempDir, "subdir")
	os.MkdirAll(subDir, 0755)

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		os.WriteFile(filePath, []byte(content), 0644)

		// Also create in subdir
		subFilePath := filepath.Join(subDir, filename)
		os.WriteFile(subFilePath, []byte(content), 0644)
	}

	// Create a file outside the static directory for path traversal tests
	outsideDir := filepath.Join(filepath.Dir(tempDir), "outside")
	os.MkdirAll(outsideDir, 0755)
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	os.WriteFile(outsideFile, []byte("secret content"), 0644)

	handler := SecureStaticHandler(http.Dir(tempDir))

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		checkHeaders   map[string]string
		checkBody      string
		shouldNotExist bool
	}{
		// Valid file access
		{
			name:           "valid CSS file",
			path:           "/test.css",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Content-Type":           "text/css; charset=utf-8",
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"Cache-Control":          "public, max-age=31536000, immutable",
			},
			checkBody: ".body { color: red; }",
		},
		{
			name:           "valid JS file",
			path:           "/script.js",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Content-Type":  "application/javascript; charset=utf-8",
				"Cache-Control": "public, max-age=31536000, immutable",
			},
			checkBody: "console.log('test');",
		},
		{
			name:           "valid PNG image",
			path:           "/image.png",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Content-Type": "image/png",
			},
		},
		{
			name:           "subdirectory file",
			path:           "/subdir/test.css",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Content-Type": "text/css; charset=utf-8",
			},
		},

		// Path traversal attempts
		{
			name:           "dot dot traversal",
			path:           "/../outside/secret.txt",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "encoded dot dot traversal",
			path:           "/%2E%2E/outside/secret.txt",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "multiple dot dot",
			path:           "/../../outside/secret.txt",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "mixed separators",
			path:           "/..\\outside\\secret.txt",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "nested traversal",
			path:           "/subdir/../../../outside/secret.txt",
			expectedStatus: http.StatusForbidden,
		},

		// Forbidden file extensions
		{
			name:           "PHP file blocked",
			path:           "/unsafe.php",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "config file blocked",
			path:           "/config.conf",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "executable blocked",
			path:           "/test.exe",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "shell script blocked",
			path:           "/script.sh",
			expectedStatus: http.StatusForbidden,
		},

		// File not found
		{
			name:           "nonexistent file",
			path:           "/nonexistent.css",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "no extension",
			path:           "/README",
			expectedStatus: http.StatusNotFound, // File doesn't exist
		},

		// Edge cases
		{
			name:           "root path",
			path:           "/",
			expectedStatus: http.StatusOK, // Directory listing allowed
		},

		{
			name:           "case sensitivity",
			path:           "/TEST.CSS",
			expectedStatus: http.StatusOK, // Windows file system is case insensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testutils.NewTestRequest("GET", tt.path, "")
			rr := testutils.NewTestResponseRecorder()

			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check headers
			for headerName, expectedValue := range tt.checkHeaders {
				actualValue := rr.Header().Get(headerName)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s: %s, got: %s", headerName, expectedValue, actualValue)
				}
			}

			// Check body content
			if tt.checkBody != "" {
				body := rr.Body.String()
				if body != tt.checkBody {
					t.Errorf("Expected body %q, got %q", tt.checkBody, body)
				}
			}
		})
	}
}

func TestSecureStaticHandlerMethods(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test CSS file
	testFile := filepath.Join(tempDir, "test.css")
	os.WriteFile(testFile, []byte(".test {}"), 0644)

	handler := SecureStaticHandler(http.Dir(tempDir))

	methods := []struct {
		method         string
		expectedStatus int
	}{
		{"GET", http.StatusOK},
		{"HEAD", http.StatusOK},
		{"POST", http.StatusOK},    // File server allows all methods
		{"PUT", http.StatusOK},     // File server allows all methods
		{"DELETE", http.StatusOK},  // File server allows all methods
		{"PATCH", http.StatusOK},   // File server allows all methods
		{"OPTIONS", http.StatusOK}, // File server allows all methods
	}

	for _, tt := range methods {
		t.Run(tt.method, func(t *testing.T) {
			req := testutils.NewTestRequest(tt.method, "/test.css", "")
			rr := testutils.NewTestResponseRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Method %s: expected status %d, got %d", tt.method, tt.expectedStatus, rr.Code)
			}

			// Check that HEAD doesn't return body
			if tt.method == "HEAD" && rr.Body.Len() > 0 {
				t.Error("HEAD method should not return body")
			}
		})
	}
}

func TestContentTypeDetection(t *testing.T) {
	tempDir := t.TempDir()
	handler := SecureStaticHandler(http.Dir(tempDir))

	files := map[string]struct {
		content     string
		contentType string
	}{
		"test.css":   {".body {}", "text/css; charset=utf-8"},
		"test.js":    {"console.log();", "application/javascript; charset=utf-8"},
		"test.png":   {"PNG", "image/png"},
		"test.jpg":   {"JPG", "image/jpeg"},
		"test.gif":   {"GIF", "image/gif"},
		"test.svg":   {"<svg/>", "image/svg+xml"},
		"test.ico":   {"ICO", "image/x-icon"},
		"test.woff":  {"WOFF", "text/plain"},  // No built-in type for woff
		"test.woff2": {"WOFF2", "text/plain"}, // No built-in type for woff2
		"test.webp":  {"WEBP", "image/webp"},
		"test.avif":  {"AVIF", "image/avif"},
	}

	for filename, fileData := range files {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(tempDir, filename)
			os.WriteFile(filePath, []byte(fileData.content), 0644)

			req := testutils.NewTestRequest("GET", "/"+filename, "")
			rr := testutils.NewTestResponseRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
				return
			}

			contentType := rr.Header().Get("Content-Type")
			if !strings.HasPrefix(contentType, fileData.contentType) {
				t.Errorf("Expected Content-Type to start with %s, got %s", fileData.contentType, contentType)
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.css")
	os.WriteFile(testFile, []byte(".test {}"), 0644)

	handler := SecureStaticHandler(http.Dir(tempDir))

	req := testutils.NewTestRequest("GET", "/test.css", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Cache-Control":          "public, max-age=31536000, immutable",
	}

	for headerName, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(headerName)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", headerName, expectedValue, actualValue)
		}
	}

	// Ensure Server header is not set
	serverHeader := rr.Header().Get("Server")
	if serverHeader != "" {
		t.Errorf("Server header should be empty, got: %s", serverHeader)
	}
}

func TestStaticHandlerEdgeCases(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		emptyDir := t.TempDir()
		handler := SecureStaticHandler(http.Dir(emptyDir))

		req := testutils.NewTestRequest("GET", "/nonexistent.css", "")
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for nonexistent file, got %d", rr.Code)
		}
	})

	t.Run("invalid directory", func(t *testing.T) {
		// Test with non-existent directory
		handler := SecureStaticHandler(http.Dir("/nonexistent/directory"))

		req := testutils.NewTestRequest("GET", "/test.css", "")
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		// Should return forbidden or not found
		if rr.Code != http.StatusNotFound && rr.Code != http.StatusForbidden {
			t.Errorf("Expected 404 or 403 for invalid directory, got %d", rr.Code)
		}
	})

	t.Run("symbolic links", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a file
		realFile := filepath.Join(tempDir, "real.css")
		os.WriteFile(realFile, []byte(".real {}"), 0644)

		// Create a symbolic link (if supported on the system)
		linkFile := filepath.Join(tempDir, "link.css")
		err := os.Symlink(realFile, linkFile)
		if err != nil {
			t.Skip("Symbolic links not supported on this system")
		}

		handler := SecureStaticHandler(http.Dir(tempDir))

		req := testutils.NewTestRequest("GET", "/link.css", "")
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		// Should handle symbolic links appropriately
		// (The behavior may vary based on the implementation)
		if rr.Code != http.StatusOK && rr.Code != http.StatusForbidden {
			t.Errorf("Unexpected status for symbolic link: %d", rr.Code)
		}
	})

	t.Run("very long paths", func(t *testing.T) {
		tempDir := t.TempDir()
		handler := SecureStaticHandler(http.Dir(tempDir))

		// Create a very long path
		longPath := "/" + strings.Repeat("a", 1000) + ".css"

		req := testutils.NewTestRequest("GET", longPath, "")
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		// Should handle long paths gracefully
		expectedStatuses := []int{http.StatusNotFound, http.StatusForbidden, http.StatusBadRequest, http.StatusInternalServerError}
		statusOK := false
		for _, status := range expectedStatuses {
			if rr.Code == status {
				statusOK = true
				break
			}
		}
		if !statusOK {
			t.Errorf("Unexpected status for long path: %d", rr.Code)
		}
	})
}

// Benchmark tests
func BenchmarkSecureStaticHandler(b *testing.B) {
	tempDir := b.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.css")
	os.WriteFile(testFile, []byte(".benchmark { color: blue; }"), 0644)

	handler := SecureStaticHandler(http.Dir(tempDir))
	req := testutils.NewTestRequest("GET", "/test.css", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkPathTraversalCheck(b *testing.B) {
	tempDir := b.TempDir()
	handler := SecureStaticHandler(http.Dir(tempDir))

	maliciousPaths := []string{
		"/../../../etc/passwd",
		"/%2E%2E/secret",
		"/subdir/../../../outside/file",
		"/..\\windows\\system32",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := maliciousPaths[i%len(maliciousPaths)]
		req := testutils.NewTestRequest("GET", path, "")
		rr := testutils.NewTestResponseRecorder()
		handler.ServeHTTP(rr, req)
	}
}
