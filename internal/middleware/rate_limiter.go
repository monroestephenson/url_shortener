package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket algorithm
type RateLimiter struct {
	rate       float64 // tokens per second
	bucketSize float64
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, bucketSize float64) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     bucketSize,
		lastRefill: time.Now(),
	}
}

// RateLimiterStore manages rate limiters for different IPs
type RateLimiterStore struct {
	limiters sync.Map
	rate     float64
	burst    float64
}

func NewRateLimiterStore(rate, burst float64) *RateLimiterStore {
	return &RateLimiterStore{
		rate:  rate,
		burst: burst,
	}
}

func (s *RateLimiterStore) getLimiter(ip string) *RateLimiter {
	limiter, exists := s.limiters.Load(ip)
	if !exists {
		limiter = NewRateLimiter(s.rate, s.burst)
		s.limiters.Store(ip, limiter)
	}
	return limiter.(*RateLimiter)
}

func (rl *RateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = min(rl.bucketSize, rl.tokens+elapsed*rl.rate)
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimitMiddleware creates a middleware that limits requests per IP
func RateLimitMiddleware(store *RateLimiterStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get IP address from request
			ip := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = forwardedFor
			}

			// Get rate limiter for this IP
			limiter := store.getLimiter(ip)

			if !limiter.allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
