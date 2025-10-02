package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter represents a rate limiter for a specific IP
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// RateLimitStore stores rate limiters per IP
type RateLimitStore struct {
	limiters map[string]*RateLimiter
	mutex    sync.RWMutex
	cleanup  time.Duration
}

// NewRateLimitStore creates a new rate limit store
func NewRateLimitStore(cleanupInterval time.Duration) *RateLimitStore {
	store := &RateLimitStore{
		limiters: make(map[string]*RateLimiter),
		cleanup:  cleanupInterval,
	}

	// Start cleanup goroutine
	go store.cleanupStale()

	return store
}

// Allow checks if a request is allowed for the given IP
func (r *RateLimitStore) Allow(ip string, maxRequests int, window time.Duration) bool {
	r.mutex.Lock()
	limiter, exists := r.limiters[ip]
	if !exists {
		limiter = &RateLimiter{
			tokens:     maxRequests,
			maxTokens:  maxRequests,
			refillRate: window / time.Duration(maxRequests),
			lastRefill: time.Now(),
		}
		r.limiters[ip] = limiter
	}
	r.mutex.Unlock()

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(limiter.lastRefill)
	tokensToAdd := int(elapsed / limiter.refillRate)

	if tokensToAdd > 0 {
		limiter.tokens = min(limiter.maxTokens, limiter.tokens+tokensToAdd)
		limiter.lastRefill = now
	}

	// Check if we have tokens available
	if limiter.tokens > 0 {
		limiter.tokens--
		return true
	}

	return false
}

// cleanupStale removes old rate limiters
func (r *RateLimitStore) cleanupStale() {
	ticker := time.NewTicker(r.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()
		now := time.Now()
		for ip, limiter := range r.limiters {
			limiter.mutex.Lock()
			// Remove limiters that haven't been used in the last hour
			if now.Sub(limiter.lastRefill) > time.Hour {
				delete(r.limiters, ip)
			}
			limiter.mutex.Unlock()
		}
		r.mutex.Unlock()
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxy/load balancer scenarios)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if there are multiple
		return xff[:findFirstComma(xff)]
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func findFirstComma(s string) int {
	for i, c := range s {
		if c == ',' {
			return i
		}
	}
	return len(s)
}

// RateLimit creates a rate limiting middleware
func RateLimit(store *RateLimitStore, maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !store.Allow(ip, maxRequests, window) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
