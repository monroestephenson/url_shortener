package models

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	MaxURLLength = 2048 // Common maximum URL length
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateURL performs comprehensive validation of URLs
func ValidateURL(urlStr string) error {
	// Check length
	if len(urlStr) > MaxURLLength {
		return &ValidationError{
			Field:   "url",
			Message: fmt.Sprintf("URL length exceeds maximum allowed length of %d characters", MaxURLLength),
		}
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return &ValidationError{
			Field:   "url",
			Message: "Invalid URL format",
		}
	}

	// Check scheme
	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return &ValidationError{
			Field:   "url",
			Message: "URL must use HTTP or HTTPS scheme",
		}
	}

	// Check for host
	if parsedURL.Host == "" {
		return &ValidationError{
			Field:   "url",
			Message: "URL must contain a valid host",
		}
	}

	// Check for malicious patterns (basic check - expand as needed)
	maliciousPatterns := []string{
		"javascript:",
		"data:",
		"vbscript:",
		"file:",
		"about:",
		"<script",
		"</script>",
	}

	urlLower := strings.ToLower(urlStr)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(urlLower, pattern) {
			return &ValidationError{
				Field:   "url",
				Message: "URL contains potentially malicious content",
			}
		}
	}

	return nil
}

// SanitizeURL cleans the URL while preserving its functionality
func SanitizeURL(urlStr string) string {
	// Trim whitespace
	urlStr = strings.TrimSpace(urlStr)

	// Remove common tracking parameters (expand as needed)
	if parsedURL, err := url.Parse(urlStr); err == nil {
		q := parsedURL.Query()
		trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "fbclid", "gclid"}
		for _, param := range trackingParams {
			q.Del(param)
		}
		parsedURL.RawQuery = q.Encode()
		return parsedURL.String()
	}

	return urlStr
}
