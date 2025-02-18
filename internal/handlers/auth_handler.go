package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"

	"url_shortener/internal/logger"
	"url_shortener/internal/models"
	"url_shortener/internal/repository"
)

type AuthHandler struct {
	userRepo repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(userRepo repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.GetLogger().Error("Failed to hash password", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username: req.Username,
	}

	if err := h.userRepo.Create(user, string(hashedPassword)); err != nil {
		if err == repository.ErrUserAlreadyExists {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		logger.GetLogger().Error("Failed to create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := h.generateToken(user)
	if err != nil {
		logger.GetLogger().Error("Failed to generate token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err == repository.ErrUserNotFound {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		logger.GetLogger().Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := h.generateToken(user)
	if err != nil {
		logger.GetLogger().Error("Failed to generate token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"username": user.Username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(h.jwtSecret))
} 