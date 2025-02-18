package models

import (
    "time"
)

// ShortURL represents the data structure for a URL record.
type ShortURL struct {
    ID          int       `json:"id"`
    ShortCode   string    `json:"shortCode"`
    OriginalURL string    `json:"url"`
    AccessCount int       `json:"accessCount"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
