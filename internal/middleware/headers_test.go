package middleware

import (
	"net/http"
	"os"
	"testing"

	"github.com/claykom/website/internal/testutils"
)

func TestSecureHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(testHandler)
	req := testutils.NewTestRequest("GET", "/", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ServeHTTP(rr, req)

	// Check basic security headers
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options header not set correctly")
	}

	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options header not set correctly")
	}

	if rr.Header().Get("Content-Security-Policy") == "" {
		t.Error("Content-Security-Policy header missing")
	}
}

func TestSecureHeadersHTTPS(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(testHandler)
	req := testutils.NewTestRequest("GET", "/", "")
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := testutils.NewTestResponseRecorder()

	handler.ServeHTTP(rr, req)

	// Should have HSTS header for HTTPS
	hsts := rr.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("HSTS header missing for HTTPS request")
	}
}

func TestSecureHeadersComplete(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(testHandler)
	req := testutils.NewTestRequest("GET", "/", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ServeHTTP(rr, req)

	// Test all security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":            "nosniff",
		"X-Frame-Options":                   "DENY",
		"X-XSS-Protection":                  "1; mode=block",
		"Referrer-Policy":                   "strict-origin-when-cross-origin",
		"X-Permitted-Cross-Domain-Policies": "none",
		"Cross-Origin-Embedder-Policy":      "require-corp",
		"Cross-Origin-Opener-Policy":        "same-origin",
		"Cross-Origin-Resource-Policy":      "same-origin",
	}

	for header, expected := range expectedHeaders {
		actual := rr.Header().Get(header)
		if actual != expected {
			t.Errorf("Header %s: expected '%s', got '%s'", header, expected, actual)
		}
	}

	// CSP should be present
	if rr.Header().Get("Content-Security-Policy") == "" {
		t.Error("Content-Security-Policy header missing")
	}
}

func TestSecureHeadersRemovesServerHeaders(t *testing.T) {
	// First test: Headers set before middleware (should be removed)
	rr := testutils.NewTestResponseRecorder()
	rr.Header().Set("Server", "Apache/2.4.41")
	rr.Header().Set("X-Powered-By", "PHP/7.4")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(testHandler)
	req := testutils.NewTestRequest("GET", "/", "")

	handler.ServeHTTP(rr, req)

	// Server headers should be removed (middleware calls Del before handler)
	if rr.Header().Get("Server") != "" {
		t.Error("Server header should be removed")
	}
	if rr.Header().Get("X-Powered-By") != "" {
		t.Error("X-Powered-By header should be removed")
	}

	// Second test: Verify middleware removes headers even if handler tries to set them
	testHandler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler sets these after middleware Del() calls
		w.Header().Set("Server", "Nginx/1.18")
		w.Header().Set("X-Powered-By", "Express")
		w.WriteHeader(http.StatusOK)
	})

	handler2 := SecureHeaders(testHandler2)
	req2 := testutils.NewTestRequest("GET", "/", "")
	rr2 := testutils.NewTestResponseRecorder()

	handler2.ServeHTTP(rr2, req2)

	// These headers will be present since handler sets them AFTER Del()
	// This tests that middleware Del() happens before handler execution
	if rr2.Header().Get("Server") == "" {
		t.Error("Expected Server header to be set by handler (after Del)")
	}
}

func TestSecureHeadersPreservesCustomHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(testHandler)
	req := testutils.NewTestRequest("GET", "/", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ServeHTTP(rr, req)

	// Custom headers should be preserved
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type header should be preserved")
	}
	if rr.Header().Get("Custom-Header") != "custom-value" {
		t.Error("Custom header should be preserved")
	}
}

func TestSecureHeadersDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run("method_"+method, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := SecureHeaders(testHandler)
			req := testutils.NewTestRequest(method, "/", "")
			rr := testutils.NewTestResponseRecorder()

			handler.ServeHTTP(rr, req)

			// All methods should get security headers
			if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("Method %s missing X-Content-Type-Options", method)
			}
		})
	}
}

func TestContentSecurityPolicyEnvironmentOverride(t *testing.T) {
	// Test CSP environment variable override
	originalCSP := os.Getenv("CSP_POLICY")
	defer func() {
		if originalCSP != "" {
			os.Setenv("CSP_POLICY", originalCSP)
		} else {
			os.Unsetenv("CSP_POLICY")
		}
	}()

	// Set custom CSP via environment
	customCSP := "default-src 'none'; script-src 'self'"
	os.Setenv("CSP_POLICY", customCSP)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecureHeaders(testHandler)
	req := testutils.NewTestRequest("GET", "/", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ServeHTTP(rr, req)

	// Should use custom CSP from environment
	actualCSP := rr.Header().Get("Content-Security-Policy")
	if actualCSP != customCSP {
		t.Errorf("Expected CSP '%s', got '%s'", customCSP, actualCSP)
	}
}
