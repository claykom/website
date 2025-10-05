package testutils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestResponseRecorder wraps httptest.ResponseRecorder with additional helper methods
type TestResponseRecorder struct {
	*httptest.ResponseRecorder
}

// NewTestResponseRecorder creates a new TestResponseRecorder
func NewTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

// AssertStatusCode checks if the response has the expected status code
func (r *TestResponseRecorder) AssertStatusCode(t *testing.T, expected int) {
	t.Helper()
	if r.Code != expected {
		t.Errorf("Expected status code %d, got %d", expected, r.Code)
	}
}

// AssertHeader checks if a header has the expected value
func (r *TestResponseRecorder) AssertHeader(t *testing.T, header, expected string) {
	t.Helper()
	actual := r.Header().Get(header)
	if actual != expected {
		t.Errorf("Expected header %s to be '%s', got '%s'", header, expected, actual)
	}
}

// AssertHeaderContains checks if a header contains the expected substring
func (r *TestResponseRecorder) AssertHeaderContains(t *testing.T, header, expected string) {
	t.Helper()
	actual := r.Header().Get(header)
	if !strings.Contains(actual, expected) {
		t.Errorf("Expected header %s to contain '%s', got '%s'", header, expected, actual)
	}
}

// AssertBodyContains checks if the response body contains the expected string
func (r *TestResponseRecorder) AssertBodyContains(t *testing.T, expected string) {
	t.Helper()
	body := r.Body.String()
	if !strings.Contains(body, expected) {
		t.Errorf("Expected body to contain '%s', got: %s", expected, body)
	}
}

// AssertContentType checks if the content type matches expected
func (r *TestResponseRecorder) AssertContentType(t *testing.T, expected string) {
	t.Helper()
	r.AssertHeader(t, "Content-Type", expected)
}

// NewTestRequest creates a new HTTP request for testing
func NewTestRequest(method, path string, body string) *http.Request {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("User-Agent", "Test Agent")
	return req
}

// NewTestRequestWithHeaders creates a new HTTP request with custom headers
func NewTestRequestWithHeaders(method, path string, headers map[string]string) *http.Request {
	req := NewTestRequest(method, path, "")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

// SetupTestEnvironment sets up common test environment variables
func SetupTestEnvironment() {
	// Set test environment variables
	// These will be used by config.Load() during tests
	envVars := map[string]string{
		"PORT":          "8080",
		"HOST":          "localhost",
		"ENV":           "test",
		"LOG_LEVEL":     "error", // Reduce log noise during tests
		"READ_TIMEOUT":  "10s",
		"WRITE_TIMEOUT": "10s",
		"IDLE_TIMEOUT":  "30s",
	}

	for key, value := range envVars {
		if err := setEnv(key, value); err != nil {
			panic("Failed to set test environment variable: " + key)
		}
	}
}

// setEnv is a helper to set environment variables
func setEnv(key, value string) error {
	// In a real implementation, you'd use os.Setenv
	// For this example, we're keeping it simple
	return nil
}

// CleanupTestEnvironment cleans up test environment
func CleanupTestEnvironment() {
	// Clean up any test-specific resources
	// This would unset environment variables in a real implementation
}

// MockFile represents a mock file for testing static file serving
type MockFile struct {
	Name    string
	Content []byte
	IsDir   bool
}

// MockFileSystem represents a mock file system for testing
type MockFileSystem struct {
	Files map[string]MockFile
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string]MockFile),
	}
}

// AddFile adds a file to the mock file system
func (mfs *MockFileSystem) AddFile(path string, content []byte) {
	mfs.Files[path] = MockFile{
		Name:    path,
		Content: content,
		IsDir:   false,
	}
}

// AddDirectory adds a directory to the mock file system
func (mfs *MockFileSystem) AddDirectory(path string) {
	mfs.Files[path] = MockFile{
		Name:  path,
		IsDir: true,
	}
}

// TestTable represents a test case for table-driven tests
type TestTable struct {
	Name          string
	Input         interface{}
	Expected      interface{}
	ExpectedError bool
	ErrorMessage  string
	Setup         func()
	Cleanup       func()
}

// RunTableTests runs a series of table-driven tests
func RunTableTests(t *testing.T, tests []TestTable, testFunc func(*testing.T, TestTable)) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Setup != nil {
				tt.Setup()
			}

			testFunc(t, tt)

			if tt.Cleanup != nil {
				tt.Cleanup()
			}
		})
	}
}
