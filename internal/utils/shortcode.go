package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

const (
	// DefaultShortCodeLength is the default length for generated short codes
	DefaultShortCodeLength = 6

	// Characters used in short code generation
	shortCodeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateSecureShortCode generates a cryptographically secure random short code
func GenerateSecureShortCode(length int) (string, error) {
	if length <= 0 {
		length = DefaultShortCodeLength
	}

	// Create byte slice to store result
	result := make([]byte, length)

	// Get the number of possible characters
	charsetLength := big.NewInt(int64(len(shortCodeCharset)))

	for i := 0; i < length; i++ {
		// Generate random index
		index, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}

		// Set character at current position
		result[i] = shortCodeCharset[index.Int64()]
	}

	return string(result), nil
}

// GenerateBase64ShortCode generates a URL-safe base64 encoded short code
func GenerateBase64ShortCode(length int) (string, error) {
	// Calculate number of random bytes needed
	// base64 encoding: 4 characters per 3 bytes
	numBytes := (length * 3) / 4
	if (length*3)%4 != 0 {
		numBytes++
	}

	// Generate random bytes
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode to base64 URL-safe string
	encoded := base64.URLEncoding.EncodeToString(randomBytes)

	// Trim to desired length
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}
