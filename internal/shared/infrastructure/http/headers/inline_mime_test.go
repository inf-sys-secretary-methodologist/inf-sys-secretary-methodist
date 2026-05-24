package headers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/headers"
)

// TestIsInlineSafeMime guards the v0.156.0 ADR-2 whitelist (#266).
//
// `?inline=true` strips X-Frame-Options + rewrites CSP frame-ancestors *
// для preview functionality. Pre-fix it did this для ANY authenticated
// download — clickjacking vector. Post-fix only whitelisted preview-safe
// MIME types are eligible; non-whitelisted forced к attachment.
func TestIsInlineSafeMime(t *testing.T) {
	tests := []struct {
		name     string
		mime     string
		expected bool
	}{
		// Preview-safe (whitelisted)
		{"png image", "image/png", true},
		{"jpeg image", "image/jpeg", true},
		{"gif image", "image/gif", true},
		{"webp image", "image/webp", true},
		{"svg image", "image/svg+xml", true},
		{"pdf document", "application/pdf", true},
		{"plain text", "text/plain", true},
		// Mime type with charset parameter — should still match
		{"plain text with charset", "text/plain; charset=utf-8", true},
		{"pdf with charset", "application/pdf; charset=binary", true},

		// Executable / scriptable — rejected
		{"html executable", "text/html", false},
		{"javascript", "application/javascript", false},
		{"javascript text", "text/javascript", false},
		{"executable", "application/x-msdownload", false},
		{"zip archive", "application/zip", false},
		{"office docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"unknown stream", "application/octet-stream", false},
		{"empty mime", "", false},
		{"weird case", "IMAGE/PNG", true}, // case-insensitive whitelist
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, headers.IsInlineSafeMime(tt.mime),
				"mime %q expected safe=%v", tt.mime, tt.expected)
		})
	}
}
