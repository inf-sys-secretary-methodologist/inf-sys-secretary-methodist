package persistence

import "testing"

func TestSanitizeForTsquery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single word",
			input:    "со",
			expected: "со:*",
		},
		{
			name:     "single word with spaces",
			input:    "  собака  ",
			expected: "собака:*",
		},
		{
			name:     "multiple words",
			input:    "собака документ",
			expected: "собака:* & документ:*",
		},
		{
			name:     "special characters escaped",
			input:    "test&query|here",
			expected: "test:* & query:* & here:*",
		},
		{
			name:     "parentheses removed",
			input:    "test(query)",
			expected: "test:* & query:*",
		},
		{
			name:     "colon and asterisk removed",
			input:    "test:*query",
			expected: "test:* & query:*",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "&|!()",
			expected: "",
		},
		{
			name:     "russian text with mixed case",
			input:    "Документ Приказ",
			expected: "Документ:* & Приказ:*",
		},
		{
			name:     "mixed alphanumeric",
			input:    "документ123 test456",
			expected: "документ123:* & test456:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeForTsquery(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeForTsquery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
