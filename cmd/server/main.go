package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

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

	// Initialize logger
	if err := logger.Initialize(cfg.Server.Mode == "development"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	log := logger.GetLogger()

	// Connect to MySQL
	database, err := db.NewMySQLDB(cfg.Database.DSN)
	if err != nil {
		log.Error("Could not connect to MySQL", zap.Error(err))
		os.Exit(1)
	}
	defer database.Close()

	// Initialize repository
	repo := repository.NewShortURLRepository(database)
	userRepo := repository.NewUserRepository(database)

	// Initialize handlers
	shortURLHandler := handlers.NewShortURLHandler(repo)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWT.Secret)

	// Setup router
	r := mux.NewRouter()

	// Auth routes (no auth required)
	r.HandleFunc("/auth/signup", authHandler.Signup).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Add middleware
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.LoggingMiddleware)
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

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
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
