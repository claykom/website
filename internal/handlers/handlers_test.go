package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/claykom/website/internal/testutils"
)

func TestHome(t *testing.T) {
	req := testutils.NewTestRequest("GET", "/", "")
	rr := testutils.NewTestResponseRecorder()

	Home(rr, req)

	// Check that we got a successful response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Check that response contains HTML (from templ rendering)
	body := rr.Body.String()
	if !strings.Contains(body, "<html") || !strings.Contains(body, "</html>") {
		t.Error("Expected response to contain HTML content")
	}
}

func TestHealth(t *testing.T) {
	// Record the time before calling the handler to test uptime
	beforeTest := time.Now()

	req := testutils.NewTestRequest("GET", "/health", "")
	rr := testutils.NewTestResponseRecorder()

	Health(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, ct)
	}

	// Parse response body
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Check required fields
	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", status)
	}

	if version, ok := response["version"]; !ok || version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", version)
	}

	// Check timestamp format
	if timestamp, ok := response["timestamp"]; ok {
		if timestampStr, ok := timestamp.(string); ok {
			if _, err := time.Parse(time.RFC3339, timestampStr); err != nil {
				t.Errorf("Invalid timestamp format: %v", err)
			}
		} else {
			t.Error("Timestamp should be a string")
		}
	} else {
		t.Error("Expected timestamp field in response")
	}

	// Check uptime exists and is reasonable
	if uptime, ok := response["uptime"]; ok {
		if uptimeStr, ok := uptime.(string); ok {
			if uptimeDuration, err := time.ParseDuration(uptimeStr); err != nil {
				t.Errorf("Invalid uptime format: %v", err)
			} else {
				// Uptime should be positive and less than test duration
				if uptimeDuration < 0 || uptimeDuration > time.Since(beforeTest)+time.Second {
					t.Errorf("Uptime seems unreasonable: %v", uptimeDuration)
				}
			}
		} else {
			t.Error("Uptime should be a string")
		}
	} else {
		t.Error("Expected uptime field in response")
	}
}

func TestHealthDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run("method_"+method, func(t *testing.T) {
			req := testutils.NewTestRequest(method, "/health", "")
			rr := testutils.NewTestResponseRecorder()

			Health(rr, req)

			// Health endpoint should work with any method
			if rr.Code != http.StatusOK {
				t.Errorf("Expected status code %d for method %s, got %d", http.StatusOK, method, rr.Code)
			}

			// Should always return JSON
			expectedContentType := "application/json"
			if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("Expected content type %s for method %s, got %s", expectedContentType, method, ct)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	req := testutils.NewTestRequest("GET", "/nonexistent", "")
	rr := testutils.NewTestResponseRecorder()

	NotFound(rr, req)

	// Check status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rr.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, ct)
	}

	// Parse response body
	var errorResponse ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Check error response fields
	if errorResponse.Code != http.StatusNotFound {
		t.Errorf("Expected error code %d, got %d", http.StatusNotFound, errorResponse.Code)
	}

	expectedError := "Not Found"
	if errorResponse.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse.Error)
	}

	expectedMessage := "The requested resource was not found"
	if errorResponse.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, errorResponse.Message)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	req := testutils.NewTestRequest("POST", "/", "")
	rr := testutils.NewTestResponseRecorder()

	MethodNotAllowed(rr, req)

	// Check status code
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, ct)
	}

	// Parse response body
	var errorResponse ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
		t.Errorf("Error unmarshaling response: %v", err)
	}

	// Check error response fields
	if errorResponse.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected error code %d, got %d", http.StatusMethodNotAllowed, errorResponse.Code)
	}

	expectedError := "Method Not Allowed"
	if errorResponse.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse.Error)
	}

	expectedMessage := "Method not allowed"
	if errorResponse.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, errorResponse.Message)
	}
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		payload  interface{}
		expected string
	}{
		{
			name:     "simple string",
			code:     http.StatusOK,
			payload:  "test",
			expected: `"test"`,
		},
		{
			name: "map payload",
			code: http.StatusCreated,
			payload: map[string]string{
				"message": "success",
			},
			expected: `{"message":"success"}`,
		},
		{
			name: "error response struct",
			code: http.StatusBadRequest,
			payload: ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid input",
				Code:    400,
			},
			expected: `{"error":"Bad Request","message":"Invalid input","code":400}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testutils.NewTestResponseRecorder()

			respondWithJSON(rr, tt.code, tt.payload)

			// Check status code
			if rr.Code != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, rr.Code)
			}

			// Check content type
			expectedContentType := "application/json"
			if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("Expected content type %s, got %s", expectedContentType, ct)
			}

			// Check response body (normalize whitespace for comparison)
			actual := strings.TrimSpace(rr.Body.String())
			if actual != tt.expected {
				t.Errorf("Expected body %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name            string
		code            int
		message         string
		expectedError   string
		expectedMessage string
	}{
		{
			name:            "not found error",
			code:            http.StatusNotFound,
			message:         "Resource not found",
			expectedError:   "Not Found",
			expectedMessage: "Resource not found",
		},
		{
			name:            "bad request error",
			code:            http.StatusBadRequest,
			message:         "Invalid input",
			expectedError:   "Bad Request",
			expectedMessage: "Invalid input",
		},
		{
			name:            "internal server error",
			code:            http.StatusInternalServerError,
			message:         "Something went wrong",
			expectedError:   "Internal Server Error",
			expectedMessage: "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testutils.NewTestResponseRecorder()

			respondWithError(rr, tt.code, tt.message)

			// Check status code
			if rr.Code != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, rr.Code)
			}

			// Parse response body
			var errorResponse ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("Error unmarshaling response: %v", err)
			}

			// Check error response fields
			if errorResponse.Code != tt.code {
				t.Errorf("Expected error code %d, got %d", tt.code, errorResponse.Code)
			}

			if errorResponse.Error != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, errorResponse.Error)
			}

			if errorResponse.Message != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, errorResponse.Message)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	version := getVersion()
	expectedVersion := "1.0.0"

	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestGetUptime(t *testing.T) {
	// Reset start time for predictable test
	originalStartTime := startTime
	startTime = time.Now()
	defer func() {
		startTime = originalStartTime
	}()

	// Wait a small amount of time
	time.Sleep(10 * time.Millisecond)

	uptime := getUptime()

	// Parse the uptime duration
	uptimeDuration, err := time.ParseDuration(uptime)
	if err != nil {
		t.Errorf("Error parsing uptime duration: %v", err)
	}

	// Uptime should be at least 10ms but less than 1 second
	if uptimeDuration < 10*time.Millisecond || uptimeDuration > time.Second {
		t.Errorf("Uptime seems unreasonable: %v", uptimeDuration)
	}
}

// Benchmark tests
func BenchmarkHome(b *testing.B) {
	req := testutils.NewTestRequest("GET", "/", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		Home(rr, req)
	}
}

func BenchmarkHealth(b *testing.B) {
	req := testutils.NewTestRequest("GET", "/health", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		Health(rr, req)
	}
}

func BenchmarkRespondWithJSON(b *testing.B) {
	payload := map[string]string{"message": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		respondWithJSON(rr, http.StatusOK, payload)
	}
}
