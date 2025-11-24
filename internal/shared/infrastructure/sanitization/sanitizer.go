// Package sanitization provides input sanitization utilities for security.
package sanitization

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	// SQL sanitization - remove dangerous characters
	sqlDangerousChars = regexp.MustCompile(`[;'"\-\-\/\*\*\/]`)

	// XSS patterns to remove
	scriptTags = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	styleTags  = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	iframeTags = regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`)
	objectTags = regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`)
	embedTags  = regexp.MustCompile(`(?i)<embed[^>]*>.*?</embed>`)

	// JavaScript event handlers
	eventHandlers = regexp.MustCompile(`(?i)on\w+\s*=`)

	// JavaScript protocol
	jsProtocol = regexp.MustCompile(`(?i)javascript:`)

	// Multiple spaces
	multipleSpaces = regexp.MustCompile(`\s+`)

	// Control characters (except newline, tab, carriage return)
	controlChars = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
)

// Sanitizer provides input sanitization methods
type Sanitizer struct{}

// NewSanitizer creates a new sanitizer instance
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

// SanitizeHTML removes dangerous HTML/JS content and escapes remaining HTML
func (s *Sanitizer) SanitizeHTML(input string) string {
	if input == "" {
		return ""
	}

	// Remove dangerous tags
	result := scriptTags.ReplaceAllString(input, "")
	result = styleTags.ReplaceAllString(result, "")
	result = iframeTags.ReplaceAllString(result, "")
	result = objectTags.ReplaceAllString(result, "")
	result = embedTags.ReplaceAllString(result, "")

	// Remove event handlers
	result = eventHandlers.ReplaceAllString(result, "")

	// Remove javascript: protocol
	result = jsProtocol.ReplaceAllString(result, "")

	// Escape remaining HTML
	result = html.EscapeString(result)

	return result
}

// SanitizeString performs basic string sanitization
func (s *Sanitizer) SanitizeString(input string) string {
	if input == "" {
		return ""
	}

	// Remove control characters
	result := controlChars.ReplaceAllString(input, "")

	// Trim whitespace
	result = strings.TrimSpace(result)

	// Normalize multiple spaces to single space
	result = multipleSpaces.ReplaceAllString(result, " ")

	return result
}

// SanitizeEmail sanitizes email addresses
func (s *Sanitizer) SanitizeEmail(email string) string {
	if email == "" {
		return ""
	}

	// Convert to lowercase
	email = strings.ToLower(email)

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Remove control characters
	email = controlChars.ReplaceAllString(email, "")

	return email
}

// SanitizeUsername sanitizes usernames (alphanumeric + underscore + dash)
func (s *Sanitizer) SanitizeUsername(username string) string {
	if username == "" {
		return ""
	}

	// Trim whitespace
	username = strings.TrimSpace(username)

	// Keep only alphanumeric, underscore, and dash
	var result strings.Builder
	for _, r := range username {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeAlphanumeric keeps only letters and digits
func (s *Sanitizer) SanitizeAlphanumeric(input string) string {
	if input == "" {
		return ""
	}

	var result strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// RemoveSQLDangerousChars removes potentially dangerous SQL characters
// Note: This is NOT a replacement for prepared statements, just an additional layer
func (s *Sanitizer) RemoveSQLDangerousChars(input string) string {
	if input == "" {
		return ""
	}

	// Remove dangerous SQL characters
	result := sqlDangerousChars.ReplaceAllString(input, "")

	// Trim whitespace
	result = strings.TrimSpace(result)

	return result
}

// TrimWhitespace trims leading and trailing whitespace
func (s *Sanitizer) TrimWhitespace(input string) string {
	return strings.TrimSpace(input)
}

// NormalizeWhitespace normalizes all whitespace to single spaces
func (s *Sanitizer) NormalizeWhitespace(input string) string {
	if input == "" {
		return ""
	}

	// Normalize multiple spaces
	result := multipleSpaces.ReplaceAllString(input, " ")

	// Trim
	result = strings.TrimSpace(result)

	return result
}

// RemoveControlCharacters removes control characters from string
func (s *Sanitizer) RemoveControlCharacters(input string) string {
	if input == "" {
		return ""
	}

	return controlChars.ReplaceAllString(input, "")
}

// SanitizeText performs comprehensive text sanitization for general text input
func (s *Sanitizer) SanitizeText(input string) string {
	if input == "" {
		return ""
	}

	// Remove control characters
	result := s.RemoveControlCharacters(input)

	// HTML escape for safety
	result = html.EscapeString(result)

	// Normalize whitespace
	result = s.NormalizeWhitespace(result)

	return result
}

// Helper functions for common sanitization patterns

// SanitizeHTML is a package-level helper
func SanitizeHTML(input string) string {
	s := NewSanitizer()
	return s.SanitizeHTML(input)
}

// SanitizeString is a package-level helper
func SanitizeString(input string) string {
	s := NewSanitizer()
	return s.SanitizeString(input)
}

// SanitizeEmail is a package-level helper
func SanitizeEmail(email string) string {
	s := NewSanitizer()
	return s.SanitizeEmail(email)
}

// SanitizeUsername is a package-level helper
func SanitizeUsername(username string) string {
	s := NewSanitizer()
	return s.SanitizeUsername(username)
}

// SanitizeText is a package-level helper
func SanitizeText(input string) string {
	s := NewSanitizer()
	return s.SanitizeText(input)
}
