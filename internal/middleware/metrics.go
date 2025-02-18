package middleware

import (
	"net/http"
	"strconv"
	"time"

	"url_shortener/internal/metrics"
)

// MetricsMiddleware wraps an http.Handler and records metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		rw := metrics.NewResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		metrics.RequestDuration.WithLabelValues(
			r.URL.Path,
			r.Method,
			strconv.Itoa(rw.StatusCode()),
		).Observe(duration.Seconds())
	})
}
