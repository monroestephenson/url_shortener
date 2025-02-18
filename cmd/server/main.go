package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"

    "url_shortener/internal/db"
    "url_shortener/internal/handlers"
    "url_shortener/internal/repository"
)

func main() {
    // Example DSN: "user:password@tcp(localhost:3306)/url_shortener?parseTime=true"
    dsn := os.Getenv("MYSQL_DSN")
    if dsn == "" {
        dsn = "root@tcp(127.0.0.1:3306)/url_shortener?parseTime=true"
    }

    // Connect to MySQL
    database, err := db.NewMySQLDB(dsn)
    if err != nil {
        log.Fatalf("Could not connect to MySQL: %v", err)
    }

    // Initialize repository
    repo := repository.NewShortURLRepository(database)

    // Initialize handlers
    shortURLHandler := handlers.NewShortURLHandler(repo)

    // Setup router
    r := mux.NewRouter()

    // Routes
    r.HandleFunc("/shorten", shortURLHandler.CreateShortURL).Methods("POST")
    r.HandleFunc("/shorten/{shortCode}", shortURLHandler.GetShortURL).Methods("GET")
    r.HandleFunc("/shorten/{shortCode}", shortURLHandler.UpdateShortURL).Methods("PUT")
    r.HandleFunc("/shorten/{shortCode}", shortURLHandler.DeleteShortURL).Methods("DELETE")
    r.HandleFunc("/shorten/{shortCode}/stats", shortURLHandler.GetShortURLStats).Methods("GET")
    
    // Redirect route
    r.HandleFunc("/{shortCode}", shortURLHandler.RedirectToOriginalURL).Methods("GET")

    // Start the server
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
