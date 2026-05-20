package http

import (
	"strings"
)

// inlineSafeMimes is the whitelist of MIME types eligible для
// `Content-Disposition: inline` preview. Limited к non-scriptable,
// non-executable payloads. Keys are lower-cased media types without
// parameters (charset etc. stripped before lookup).
var inlineSafeMimes = map[string]struct{}{
	"image/png":       {},
	"image/jpeg":      {},
	"image/gif":       {},
	"image/webp":      {},
	"image/svg+xml":   {},
	"application/pdf": {},
	"text/plain":      {},
}

// IsInlineSafeMime reports whether a MIME type is safe to serve with
// `Content-Disposition: inline` для browser preview without inviting
// clickjacking or scriptable-resource execution.
//
// v0.156.0 ADR-2 (#266): the DownloadFile handler used к unconditionally
// strip X-Frame-Options and rewrite CSP frame-ancestors `*` whenever
// `?inline=true` was passed — for ANY authenticated download. This
// whitelist scopes that loosening к preview-friendly types only;
// executable / scriptable / unknown payloads are forced к attachment
// regardless of query.
func IsInlineSafeMime(mime string) bool {
	if mime == "" {
		return false
	}
	// Strip parameters (e.g. "text/plain; charset=utf-8" → "text/plain").
	if idx := strings.Index(mime, ";"); idx >= 0 {
		mime = mime[:idx]
	}
	mime = strings.ToLower(strings.TrimSpace(mime))
	_, ok := inlineSafeMimes[mime]
	return ok
}
