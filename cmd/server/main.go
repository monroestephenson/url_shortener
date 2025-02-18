package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	swaggerMiddleware "github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"url_shortener/internal/cache"
	"url_shortener/internal/config"
	"url_shortener/internal/db"
	"url_shortener/internal/handlers"
	"url_shortener/internal/logger"
	"url_shortener/internal/middleware"
	"url_shortener/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override port from environment if provided
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}

	// Initialize logger
	if err := logger.Initialize(cfg.Server.Mode == "development"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	log := logger.GetLogger()

	// Initialize Redis cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6380" // Default to the Docker port
	}
	redisCache, err := cache.NewRedisCache(redisURL, "", 0)
	if err != nil {
		log.Error("Could not connect to Redis", zap.Error(err))
		os.Exit(1)
	}
	defer redisCache.Close()

	// Get MySQL DSN from environment or config
	dsn := os.Getenv("MYSQL_DSN")
	if dsn != "" {
		cfg.Database.DSN = dsn
	}

	// Connect to MySQL
	database, err := db.NewMySQLDB(cfg.Database.DSN)
	if err != nil {
		log.Error("Could not connect to MySQL", zap.Error(err))
		os.Exit(1)
	}
	defer database.Close()

	// Initialize repositories
	repo := repository.NewShortURLRepository(database)
	userRepo := repository.NewUserRepository(database)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiterStore(100, 200) // 100 requests per second, burst of 200

	// Initialize handlers
	shortURLHandler := handlers.NewShortURLHandler(repo, redisCache)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWT.Secret)

	// Setup router
	r := mux.NewRouter()

	// API Documentation
	opts := swaggerMiddleware.SwaggerUIOpts{
		BasePath: "/",
		SpecURL:  "/swagger.yaml",
	}
	sh := swaggerMiddleware.SwaggerUI(opts, nil)
	r.Handle("/docs", sh)
	r.Handle("/swagger.yaml", http.FileServer(http.Dir("api")))

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Auth routes (no auth required)
	r.HandleFunc("/auth/signup", authHandler.Signup).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Add middleware
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.LoggingMiddleware)
	api.Use(middleware.MetricsMiddleware)
	api.Use(middleware.RateLimitMiddleware(rateLimiter))
	api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))

	// Protected routes
	api.HandleFunc("/shorten", shortURLHandler.CreateShortURL).Methods("POST")
	api.HandleFunc("/shorten/{shortCode}", shortURLHandler.GetShortURL).Methods("GET")
	api.HandleFunc("/shorten/{shortCode}", shortURLHandler.UpdateShortURL).Methods("PUT")
	api.HandleFunc("/shorten/{shortCode}", shortURLHandler.DeleteShortURL).Methods("DELETE")
	api.HandleFunc("/shorten/{shortCode}/stats", shortURLHandler.GetShortURLStats).Methods("GET")

	// Redirect route (no auth required)
	redirectRouter := mux.NewRouter()
	redirectRouter.HandleFunc("/{shortCode}", shortURLHandler.RedirectToOriginalURL).Methods("GET")
	r.PathPrefix("/").Handler(redirectRouter)

	// Create server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Starting server",
			zap.String("port", cfg.Server.Port),
			zap.String("mode", cfg.Server.Mode),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", zap.Error(err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
}
