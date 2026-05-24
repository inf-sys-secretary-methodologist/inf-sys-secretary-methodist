package headers

import (
	"mime"
	"strings"
)

// BuildContentDisposition formats a Content-Disposition header value
// safely (v0.156.0 ADR-3 #266). Pre-fix the documents DownloadFile
// concatenated attacker-controlled filename directly into the header —
// CRLF injection vector; non-ASCII filenames rendered incorrectly per
// RFC 2616.
//
// Post-fix:
//   - control bytes (\r, \n, \x00 etc.) stripped from filename
//   - empty result falls back к "file"
//   - mime.FormatMediaType emits RFC 2231 filename* parameter for
//     non-ASCII payloads и quotes ASCII consistently
//   - on FormatMediaType error (very unlikely after sanitize) falls
//     back к a hardcoded safe value
func BuildContentDisposition(disposition, filename string) string {
	safe := stripControlBytes(filename)
	if safe == "" {
		safe = "file"
	}
	formatted := mime.FormatMediaType(disposition, map[string]string{"filename": safe})
	if formatted == "" {
		// FormatMediaType returns empty on invalid input; should not
		// happen after stripControlBytes, но use a safe fallback to
		// avoid panic / blank header.
		return disposition + `; filename="file"`
	}
	return formatted
}

// stripControlBytes removes ASCII control bytes (< 0x20 or DEL) and
// double-quote characters from s. mime.FormatMediaType handles encoding
// of other non-ASCII bytes via RFC 2231; control bytes are the actual
// injection / smuggling vector.
func stripControlBytes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r < 0x20:
			continue
		case r == 0x7F:
			continue
		case r == '"':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
