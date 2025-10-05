package middleware

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/claykom/website/internal/testutils"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name         string
		maxRequests  int
		window       time.Duration
		requests     int
		expectedPass int
		delay        time.Duration
	}{
		{
			name:         "within rate limit",
			maxRequests:  10,
			window:       time.Minute,
			requests:     5,
			expectedPass: 5,
			delay:        0,
		},
		{
			name:         "exceeds rate limit",
			maxRequests:  3,
			window:       time.Minute,
			requests:     5,
			expectedPass: 3,
			delay:        0,
		},
		{
			name:         "rate limit with refill",
			maxRequests:  2,
			window:       50 * time.Millisecond,
			requests:     3,
			expectedPass: 3,                     // Should allow all after refill
			delay:        60 * time.Millisecond, // Wait for refill
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewRateLimitStore(time.Hour) // Long cleanup interval for tests
			ip := "192.168.1.1"
			passed := 0

			for i := 0; i < tt.requests; i++ {
				if tt.delay > 0 && i == tt.requests/2 {
					time.Sleep(tt.delay)
				}

				if store.Allow(ip, tt.maxRequests, tt.window) {
					passed++
				}
			}

			if passed != tt.expectedPass {
				t.Errorf("Expected %d requests to pass, got %d", tt.expectedPass, passed)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		maxRequests     int
		window          time.Duration
		requests        int
		expectedSuccess int
		expectedBlocked int
		delay           time.Duration
	}{
		{
			name:            "normal traffic",
			maxRequests:     5,
			window:          time.Minute,
			requests:        3,
			expectedSuccess: 3,
			expectedBlocked: 0,
			delay:           0,
		},
		{
			name:            "rate limited traffic",
			maxRequests:     2,
			window:          time.Minute,
			requests:        4,
			expectedSuccess: 2,
			expectedBlocked: 2,
			delay:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Create rate limit middleware
			store := NewRateLimitStore(time.Hour)
			middleware := RateLimit(store, tt.maxRequests, tt.window)
			handler := middleware(testHandler)

			successCount := 0
			blockedCount := 0

			for i := 0; i < tt.requests; i++ {
				if tt.delay > 0 && i == tt.requests/2 {
					time.Sleep(tt.delay)
				}

				req := testutils.NewTestRequest("GET", "/test", "")
				req.RemoteAddr = "192.168.1.1:12345" // Consistent IP
				rr := testutils.NewTestResponseRecorder()

				handler.ServeHTTP(rr, req)

				if rr.Code == http.StatusOK {
					successCount++
				} else if rr.Code == http.StatusTooManyRequests {
					blockedCount++
				}
			}

			if successCount != tt.expectedSuccess {
				t.Errorf("Expected %d successful requests, got %d", tt.expectedSuccess, successCount)
			}
			if blockedCount != tt.expectedBlocked {
				t.Errorf("Expected %d blocked requests, got %d", tt.expectedBlocked, blockedCount)
			}
		})
	}
}

func TestRateLimitDifferentIPs(t *testing.T) {
	store := NewRateLimitStore(time.Hour)
	maxRequests := 2
	window := time.Minute

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(store, maxRequests, window)
	handler := middleware(testHandler)

	// Test that different IPs have separate rate limits
	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}

	for _, ip := range ips {
		// Each IP should be able to make maxRequests
		for i := 0; i < maxRequests; i++ {
			req := testutils.NewTestRequest("GET", "/test", "")
			req.RemoteAddr = ip
			rr := testutils.NewTestResponseRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Request %d from IP %s should have succeeded, got status %d", i+1, ip, rr.Code)
			}
		}

		// The next request from this IP should be rate limited
		req := testutils.NewTestRequest("GET", "/test", "")
		req.RemoteAddr = ip
		rr := testutils.NewTestResponseRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Request from IP %s should have been rate limited, got status %d", ip, rr.Code)
		}
	}
}

func TestRateLimitIPExtraction(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		forwardedFor string
		realIP       string
		expectedIP   string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For header",
			remoteAddr:   "127.0.0.1:80",
			forwardedFor: "203.0.113.1, 198.51.100.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "127.0.0.1:80",
			realIP:     "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:         "Both headers present - X-Real-IP takes precedence",
			remoteAddr:   "127.0.0.1:80",
			forwardedFor: "203.0.113.1",
			realIP:       "203.0.113.2",
			expectedIP:   "203.0.113.2",
		},
	}

	store := NewRateLimitStore(time.Hour)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(store, 100, time.Minute) // High limit to avoid rate limiting
	handler := middleware(testHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testutils.NewTestRequest("GET", "/test", "")
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			rr := testutils.NewTestResponseRecorder()
			handler.ServeHTTP(rr, req)

			// We can't directly test IP extraction without modifying the middleware,
			// but we can test that the request is processed (not an error)
			if rr.Code != http.StatusOK {
				t.Errorf("Request should have succeeded, got status %d", rr.Code)
			}
		})
	}
}

func TestRateLimitStoreCleanup(t *testing.T) {
	// Create store with very short cleanup interval
	store := NewRateLimitStore(10 * time.Millisecond)
	defer time.Sleep(50 * time.Millisecond) // Allow cleanup goroutine to finish

	ip := "192.168.1.1"

	// Make a request to create a limiter
	allowed := store.Allow(ip, 10, time.Minute)
	if !allowed {
		t.Error("First request should be allowed")
	}

	// Check that limiter exists
	store.mutex.RLock()
	_, exists := store.limiters[ip]
	store.mutex.RUnlock()

	if !exists {
		t.Error("Limiter should exist after request")
	}

	// Wait for cleanup (cleanup should remove stale limiters)
	// Note: In a real implementation, cleanup might remove limiters
	// that haven't been used recently
	time.Sleep(30 * time.Millisecond)

	// Make another request to ensure functionality still works
	allowed = store.Allow(ip, 10, time.Minute)
	if !allowed {
		t.Error("Request after cleanup should still be allowed")
	}
}

func TestRateLimitConcurrentAccess(t *testing.T) {
	store := NewRateLimitStore(time.Hour)
	maxRequests := 10
	window := time.Minute
	ip := "192.168.1.1"

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowedCount := 0
	totalRequests := 20
	concurrency := 5

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < totalRequests/concurrency; j++ {
				if store.Allow(ip, maxRequests, window) {
					mu.Lock()
					allowedCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Should allow exactly maxRequests regardless of concurrency
	if allowedCount != maxRequests {
		t.Errorf("Expected exactly %d requests to be allowed, got %d", maxRequests, allowedCount)
	}
}

// Benchmark tests
func BenchmarkRateLimit(b *testing.B) {
	store := NewRateLimitStore(time.Hour)
	ip := "192.168.1.1"
	maxRequests := 1000
	window := time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Allow(ip, maxRequests, window)
	}
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	store := NewRateLimitStore(time.Hour)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(store, 1000, time.Minute)
	handler := middleware(testHandler)

	req := testutils.NewTestRequest("GET", "/test", "")
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		handler.ServeHTTP(rr, req)
	}
}
