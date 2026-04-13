package sanitization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizer_SanitizeHTML(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "hello world", "hello world"},
		{"script tag", `<script>alert('xss')</script>hello`, "hello"},
		{"script tag case insensitive", `<SCRIPT>alert('xss')</SCRIPT>hello`, "hello"},
		{"style tag", `<style>body{display:none}</style>hello`, "hello"},
		{"iframe tag", `<iframe src="evil.com"></iframe>hello`, "hello"},
		{"object tag", `<object data="evil.swf"></object>hello`, "hello"},
		{"embed tag", `<embed src="evil.swf"></embed>hello`, "hello"},
		{"event handler", `<div onclick= "alert(1)">hello</div>`, `&lt;div  &#34;alert(1)&#34;&gt;hello&lt;/div&gt;`},
		{"javascript protocol", `<a href="javascript:alert(1)">click</a>`, `&lt;a href=&#34;alert(1)&#34;&gt;click&lt;/a&gt;`},
		{"html entities escaped", `<b>bold</b>`, `&lt;b&gt;bold&lt;/b&gt;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.SanitizeHTML(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_SanitizeString(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"normal", "hello world", "hello world"},
		{"trims spaces", "  hello  ", "hello"},
		{"normalizes spaces", "hello    world", "hello world"},
		{"removes control chars", "hello\x00\x01world", "helloworld"},
		{"keeps newline tab cr", "hello\n\tworld", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.SanitizeString(tt.input))
		})
	}
}

func TestSanitizer_SanitizeEmail(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"lowercase", "User@EXAMPLE.COM", "user@example.com"},
		{"trim spaces", "  user@example.com  ", "user@example.com"},
		{"remove control chars", "user\x00@example.com", "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.SanitizeEmail(tt.input))
		})
	}
}

func TestSanitizer_SanitizeUsername(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"valid", "user_name-123", "user_name-123"},
		{"removes special chars", "user@name!#$%", "username"},
		{"trims spaces", "  user  ", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.SanitizeUsername(tt.input))
		})
	}
}

func TestSanitizer_SanitizeAlphanumeric(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"letters and digits", "abc123", "abc123"},
		{"removes special", "abc!@#123", "abc123"},
		{"removes spaces", "a b c", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.SanitizeAlphanumeric(tt.input))
		})
	}
}

func TestSanitizer_RemoveSQLDangerousChars(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"clean", "hello world", "hello world"},
		{"removes quotes", `it's a "test"`, "its a test"},
		{"removes semicolons", "SELECT; DROP TABLE", "SELECT DROP TABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.RemoveSQLDangerousChars(tt.input))
		})
	}
}

func TestSanitizer_TrimWhitespace(t *testing.T) {
	s := NewSanitizer()
	assert.Equal(t, "hello", s.TrimWhitespace("  hello  "))
	assert.Equal(t, "", s.TrimWhitespace(""))
}

func TestSanitizer_NormalizeWhitespace(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"normal", "hello world", "hello world"},
		{"multiple spaces", "hello    world", "hello world"},
		{"tabs and newlines", "hello\t\n\nworld", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.NormalizeWhitespace(tt.input))
		})
	}
}

func TestSanitizer_RemoveControlCharacters(t *testing.T) {
	s := NewSanitizer()

	assert.Equal(t, "", s.RemoveControlCharacters(""))
	assert.Equal(t, "hello", s.RemoveControlCharacters("hello"))
	assert.Equal(t, "helloworld", s.RemoveControlCharacters("hello\x00\x01\x02world"))
}

func TestSanitizer_SanitizeText(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"normal", "hello world", "hello world"},
		{"html escaped", "<b>bold</b>", "&lt;b&gt;bold&lt;/b&gt;"},
		{"control chars removed", "hello\x00world", "helloworld"},
		{"whitespace normalized", "hello    world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, s.SanitizeText(tt.input))
		})
	}
}

func TestPackageLevelHelpers(t *testing.T) {
	assert.Equal(t, "hello", SanitizeString("  hello  "))
	assert.Equal(t, "hello", SanitizeHTML("hello"))
	assert.Equal(t, "user@example.com", SanitizeEmail("  User@EXAMPLE.COM  "))
	assert.Equal(t, "username", SanitizeUsername("user@name"))
	assert.Equal(t, "hello world", SanitizeText("hello world"))
}
