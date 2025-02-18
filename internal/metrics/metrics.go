package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestDuration tracks the duration of HTTP requests
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "url_shortener_http_request_duration_seconds",
		Help: "Duration of HTTP requests in seconds",
	}, []string{"path", "method", "status"})

	// URLAccessCount tracks the number of times each short URL is accessed
	URLAccessCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "url_shortener_url_access_total",
		Help: "Total number of times a short URL has been accessed",
	}, []string{"short_code"})

	// CacheHits tracks cache hit/miss ratio
	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "url_shortener_cache_hits_total",
		Help: "Total number of cache hits/misses",
	}, []string{"result"})

	// ActiveUsers tracks the number of active users
	ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "url_shortener_active_users",
		Help: "Number of active users in the last 5 minutes",
	})

	// RateLimitExceeded tracks rate limit violations
	RateLimitExceeded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "url_shortener_rate_limit_exceeded_total",
		Help: "Total number of rate limit exceeded events",
	}, []string{"ip"})
)

// URLAccess represents a URL access event
type URLAccess struct {
	ShortCode    string
	AccessTime   time.Time
	UserAgent    string
	IPAddress    string
	RefererURL   string
	ResponseTime time.Duration
}

// MetricsMiddleware wraps an http.Handler and records metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		rw := NewResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		RequestDuration.WithLabelValues(
			r.URL.Path,
			r.Method,
			string(rw.statusCode),
		).Observe(duration.Seconds())
	})
}

// ResponseWriter wraps http.ResponseWriter to capture the status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}

// RecordURLAccess records metrics for a URL access
func RecordURLAccess(access *URLAccess) {
	URLAccessCount.WithLabelValues(access.ShortCode).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	CacheHits.WithLabelValues("hit").Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	CacheHits.WithLabelValues("miss").Inc()
}

// RecordRateLimitExceeded records a rate limit violation
func RecordRateLimitExceeded(ip string) {
	RateLimitExceeded.WithLabelValues(ip).Inc()
}
