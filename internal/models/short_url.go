package models

import (
	"time"
)

// ShortURL represents a shortened URL in the system
type ShortURL struct {
	ID          int       `json:"id"`
	ShortCode   string    `json:"shortCode" validate:"required,min=6,max=10"`
	OriginalURL string    `json:"originalUrl" validate:"required,url"`
	AccessCount int       `json:"accessCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserID      string    `json:"userId,omitempty"`
}

// CreateShortURLRequest represents the request body for creating a short URL
type CreateShortURLRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// UpdateShortURLRequest represents the request body for updating a short URL
type UpdateShortURLRequest struct {
	URL string `json:"url" validate:"required,url"`
}
