package http

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
//
// Stub returns false для всех — GREEN commit will implement the list.
func IsInlineSafeMime(mime string) bool {
	return false
}
