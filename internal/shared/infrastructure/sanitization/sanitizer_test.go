package sanitization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeHTML(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove script tags",
			input:    "<script>alert('xss')</script>hello",
			expected: "hello",
		},
		{
			name:     "remove style tags",
			input:    "<style>body{display:none}</style>text",
			expected: "text",
		},
		{
			name:     "remove iframe tags",
			input:    "<iframe src='evil.com'></iframe>content",
			expected: "content",
		},
		{
			name:     "remove event handlers",
			input:    "<div onclick='alert(1)'>text</div>",
			expected: "&lt;div &#39;alert(1)&#39;&gt;text&lt;/div&gt;",
		},
		{
			name:     "remove javascript protocol",
			input:    "<a href='javascript:alert(1)'>link</a>",
			expected: "&lt;a href=&#39;alert(1)&#39;&gt;link&lt;/a&gt;",
		},
		{
			name:     "escape remaining HTML",
			input:    "<b>bold</b>",
			expected: "&lt;b&gt;bold&lt;/b&gt;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim whitespace",
			input:    "  text  ",
			expected: "text",
		},
		{
			name:     "normalize multiple spaces",
			input:    "text    with    spaces",
			expected: "text with spaces",
		},
		{
			name:     "remove control characters",
			input:    "text\x00with\x01controls",
			expected: "textwithcontrols",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "normal text",
			input:    "normal text",
			expected: "normal text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "convert to lowercase",
			input:    "User@Example.COM",
			expected: "user@example.com",
		},
		{
			name:     "trim whitespace",
			input:    "  user@example.com  ",
			expected: "user@example.com",
		},
		{
			name:     "remove control characters",
			input:    "user\x00@example.com",
			expected: "user@example.com",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already clean",
			input:    "user@example.com",
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "keep alphanumeric",
			input:    "user123",
			expected: "user123",
		},
		{
			name:     "keep underscore and dash",
			input:    "user_name-123",
			expected: "user_name-123",
		},
		{
			name:     "remove special chars",
			input:    "user@name!",
			expected: "username",
		},
		{
			name:     "remove spaces",
			input:    "user name",
			expected: "username",
		},
		{
			name:     "trim whitespace",
			input:    "  username  ",
			expected: "username",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeAlphanumeric(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "keep letters and digits",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "remove special chars",
			input:    "abc@123#def",
			expected: "abc123def",
		},
		{
			name:     "remove spaces",
			input:    "abc 123 def",
			expected: "abc123def",
		},
		{
			name:     "remove underscores",
			input:    "abc_123",
			expected: "abc123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeAlphanumeric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveSQLDangerousChars(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove semicolon",
			input:    "text;more",
			expected: "textmore",
		},
		{
			name:     "remove quotes",
			input:    "text'with\"quotes",
			expected: "textwithquotes",
		},
		{
			name:     "remove SQL comment",
			input:    "text--comment",
			expected: "textcomment",
		},
		{
			name:     "remove slashes",
			input:    "text/*comment*/",
			expected: "textcomment",
		},
		{
			name:     "normal text",
			input:    "normaltext",
			expected: "normaltext",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.RemoveSQLDangerousChars(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimWhitespace(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim spaces",
			input:    "  text  ",
			expected: "text",
		},
		{
			name:     "trim tabs",
			input:    "\t\ttext\t\t",
			expected: "text",
		},
		{
			name:     "trim newlines",
			input:    "\n\ntext\n\n",
			expected: "text",
		},
		{
			name:     "no trimming needed",
			input:    "text",
			expected: "text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.TrimWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple spaces",
			input:    "text    with    spaces",
			expected: "text with spaces",
		},
		{
			name:     "mixed whitespace",
			input:    "text \t\n with \t whitespace",
			expected: "text with whitespace",
		},
		{
			name:     "trim and normalize",
			input:    "  text  with  spaces  ",
			expected: "text with spaces",
		},
		{
			name:     "normal text",
			input:    "text with spaces",
			expected: "text with spaces",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.NormalizeWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveControlCharacters(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove null bytes",
			input:    "text\x00with\x00nulls",
			expected: "textwithnulls",
		},
		{
			name:     "remove control chars",
			input:    "text\x01\x02\x03controls",
			expected: "textcontrols",
		},
		{
			name:     "keep newlines and tabs",
			input:    "text\nwith\ttabs",
			expected: "text\nwith\ttabs",
		},
		{
			name:     "normal text",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.RemoveControlCharacters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeText(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "comprehensive sanitization",
			input:    "  <b>text</b>  with   spaces\x00  ",
			expected: "&lt;b&gt;text&lt;/b&gt; with spaces",
		},
		{
			name:     "remove control chars and escape HTML",
			input:    "text\x01<script>alert(1)</script>",
			expected: "text&lt;script&gt;alert(1)&lt;/script&gt;",
		},
		{
			name:     "normal text",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPackageLevelHelpers(t *testing.T) {
	t.Run("SanitizeHTML helper", func(t *testing.T) {
		result := SanitizeHTML("<script>alert(1)</script>text")
		assert.Equal(t, "text", result)
	})

	t.Run("SanitizeString helper", func(t *testing.T) {
		result := SanitizeString("  text  ")
		assert.Equal(t, "text", result)
	})

	t.Run("SanitizeEmail helper", func(t *testing.T) {
		result := SanitizeEmail("User@Example.COM")
		assert.Equal(t, "user@example.com", result)
	})

	t.Run("SanitizeUsername helper", func(t *testing.T) {
		result := SanitizeUsername("user@name")
		assert.Equal(t, "username", result)
	})

	t.Run("SanitizeText helper", func(t *testing.T) {
		result := SanitizeText("  <b>text</b>  ")
		assert.Equal(t, "&lt;b&gt;text&lt;/b&gt;", result)
	})
}
