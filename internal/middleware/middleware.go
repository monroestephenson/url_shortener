package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"url_shortener/internal/logger"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request details
		logger.GetLogger().Info("request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Duration("latency", time.Since(start)),
		)
	})
}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for the redirect endpoint
			if r.Method == "GET" && !strings.HasPrefix(r.URL.Path, "/shorten") {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract the token
			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Add user ID to context
				ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			}
		})
	}
} 