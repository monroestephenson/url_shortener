package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"log"
	"url_shortener/internal/cache"
	"url_shortener/internal/models"
	"url_shortener/internal/repository"

	"github.com/gorilla/mux"
)

// ShortURLHandler handles all short URL related HTTP requests.
type ShortURLHandler struct {
	repo  repository.ShortURLRepository
	cache *cache.RedisCache
}

// NewShortURLHandler returns a new ShortURLHandler instance.
func NewShortURLHandler(repo repository.ShortURLRepository, cache *cache.RedisCache) *ShortURLHandler {
	return &ShortURLHandler{
		repo:  repo,
		cache: cache,
	}
}

// CreateShortURL - POST /shorten
func (h *ShortURLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the URL
	if strings.TrimSpace(req.URL) == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// Validate URL using the ValidateURL function
	if err := models.ValidateURL(req.URL); err != nil {
		if validationErr, ok := err.(*models.ValidationError); ok {
			http.Error(w, validationErr.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Generate a random short code
	shortCode := generateShortCode(6) // 6 characters

	su := models.ShortURL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
	}

	// Create in repository
	if err := h.repo.Create(&su); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(su)
}

// GetShortURL - GET /shorten/{shortCode}
func (h *ShortURLHandler) GetShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	su, err := h.repo.GetByShortCode(shortCode)
	if err == repository.ErrShortURLNotFound {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Optionally, increment access count if this is an actual "use" of the short URL
	// In many services, we do this in a redirect handler. For demonstration:
	if err := h.repo.IncrementAccessCount(shortCode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	su.AccessCount++ // reflect the increment in the current object

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(su)
}

// UpdateShortURL - PUT /shorten/{shortCode}
func (h *ShortURLHandler) UpdateShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.URL) == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// First, check if the short code exists
	su, err := h.repo.GetByShortCode(shortCode)
	if err == repository.ErrShortURLNotFound {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update original URL
	su.OriginalURL = req.URL

	if err := h.repo.Update(su); err != nil {
		if err == repository.ErrShortURLNotFound {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(su)
}

// DeleteShortURL - DELETE /shorten/{shortCode}
func (h *ShortURLHandler) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	err := h.repo.DeleteByShortCode(shortCode)
	if err == repository.ErrShortURLNotFound {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetShortURLStats - GET /shorten/{shortCode}/stats
func (h *ShortURLHandler) GetShortURLStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	su, err := h.repo.GetByShortCode(shortCode)
	if err == repository.ErrShortURLNotFound {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(su)
}

// RedirectToOriginalURL - GET /{shortCode}
func (h *ShortURLHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	su, err := h.repo.GetByShortCode(shortCode)
	if err == repository.ErrShortURLNotFound {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Increment access count asynchronously to not block the redirect
	go func() {
		if err := h.repo.IncrementAccessCount(shortCode); err != nil {
			// Log the error but don't affect the user experience
			log.Printf("Failed to increment access count for %s: %v", shortCode, err)
		}
	}()

	http.Redirect(w, r, su.OriginalURL, http.StatusMovedPermanently)
}

// generateShortCode creates a random alphanumeric string of length n.
func generateShortCode(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
