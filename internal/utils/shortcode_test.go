package utils

import (
	"regexp"
	"testing"
)

func TestGenerateSecureShortCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
		regex  string
	}{
		{
			name:   "Default length",
			length: 0,
			want:   DefaultShortCodeLength,
			regex:  "^[a-zA-Z0-9]{6}$",
		},
		{
			name:   "Custom length",
			length: 8,
			want:   8,
			regex:  "^[a-zA-Z0-9]{8}$",
		},
		{
			name:   "Minimum length",
			length: 4,
			want:   4,
			regex:  "^[a-zA-Z0-9]{4}$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple codes to ensure randomness
			codes := make(map[string]bool)
			for i := 0; i < 100; i++ {
				got, err := GenerateSecureShortCode(tt.length)
				if err != nil {
					t.Errorf("GenerateSecureShortCode() error = %v", err)
					return
				}

				// Check length
				if len(got) != tt.want {
					t.Errorf("GenerateSecureShortCode() length = %v, want %v", len(got), tt.want)
				}

				// Check format
				matched, err := regexp.MatchString(tt.regex, got)
				if err != nil {
					t.Errorf("Regex error: %v", err)
				}
				if !matched {
					t.Errorf("GenerateSecureShortCode() = %v, doesn't match regex %v", got, tt.regex)
				}

				// Check uniqueness
				if codes[got] {
					t.Errorf("GenerateSecureShortCode() generated duplicate code: %v", got)
				}
				codes[got] = true
			}
		})
	}
}

func TestGenerateBase64ShortCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
		regex  string
	}{
		{
			name:   "Default length",
			length: 6,
			want:   6,
			regex:  "^[A-Za-z0-9_-]{6}$",
		},
		{
			name:   "Custom length",
			length: 8,
			want:   8,
			regex:  "^[A-Za-z0-9_-]{8}$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple codes to ensure randomness
			codes := make(map[string]bool)
			for i := 0; i < 100; i++ {
				got, err := GenerateBase64ShortCode(tt.length)
				if err != nil {
					t.Errorf("GenerateBase64ShortCode() error = %v", err)
					return
				}

				// Check length
				if len(got) != tt.want {
					t.Errorf("GenerateBase64ShortCode() length = %v, want %v", len(got), tt.want)
				}

				// Check format
				matched, err := regexp.MatchString(tt.regex, got)
				if err != nil {
					t.Errorf("Regex error: %v", err)
				}
				if !matched {
					t.Errorf("GenerateBase64ShortCode() = %v, doesn't match regex %v", got, tt.regex)
				}

				// Check uniqueness
				if codes[got] {
					t.Errorf("GenerateBase64ShortCode() generated duplicate code: %v", got)
				}
				codes[got] = true
			}
		})
	}
}
