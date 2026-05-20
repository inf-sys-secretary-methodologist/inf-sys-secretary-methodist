package http_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	handlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
)

// TestBuildContentDisposition guards v0.156.0 ADR-3 (#266): header
// injection via filename parameter. Pre-fix, document_handler.go
// concatenated fileInfo.FileName directly into Content-Disposition
// header value — no CRLF strip / RFC 5987 encoding. Attacker-controlled
// filename containing "\r\n" could inject arbitrary response headers;
// non-ASCII filenames would split на first `:` per RFC 2616.
//
// Post-fix uses mime.FormatMediaType (RFC 2231 / RFC 5987) which:
// - rejects CRLF / control chars (returns empty → fallback к "file")
// - encodes non-ASCII via RFC 5987 filename* parameter (percent-encoded UTF-8)
// - quotes ASCII-only filenames consistently.
func TestBuildContentDisposition(t *testing.T) {
	tests := []struct {
		name         string
		disposition  string
		filename     string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "ascii attachment",
			disposition:  "attachment",
			filename:     "report.pdf",
			wantContains: []string{"attachment", "report.pdf"},
		},
		{
			name:         "ascii inline",
			disposition:  "inline",
			filename:     "preview.png",
			wantContains: []string{"inline", "preview.png"},
		},
		{
			name:         "cyrillic filename uses RFC 5987 encoding",
			disposition:  "attachment",
			filename:     "пример.docx",
			wantContains: []string{"attachment", "filename*="},
			wantAbsent:   []string{"\r", "\n"},
		},
		{
			name:         "crlf injection neutralized",
			disposition:  "attachment",
			filename:     "evil.pdf\r\nX-Injected: yes",
			wantContains: []string{"attachment"},
			wantAbsent:   []string{"\r\n", "X-Injected"},
		},
		{
			name:         "double quotes safely encoded",
			disposition:  "attachment",
			filename:     "file\"with\"quotes.txt",
			wantContains: []string{"attachment"},
			wantAbsent:   []string{"\r", "\n"},
		},
		{
			name:         "empty filename falls back к 'file'",
			disposition:  "attachment",
			filename:     "",
			wantContains: []string{"attachment", "file"},
		},
		{
			name:         "control chars rejected",
			disposition:  "inline",
			filename:     "ok\x00name.txt",
			wantContains: []string{"inline"},
			wantAbsent:   []string{"\x00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.BuildContentDisposition(tt.disposition, tt.filename)
			for _, want := range tt.wantContains {
				assert.Contains(t, got, want, "expected %q к contain %q", got, want)
			}
			for _, bad := range tt.wantAbsent {
				assert.False(t, strings.Contains(got, bad),
					"expected %q NOT к contain %q", got, bad)
			}
		})
	}
}
