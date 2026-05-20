package http

// BuildContentDisposition formats Content-Disposition header value безопасно.
// Stub returns the raw, unsafe concatenation matching pre-fix behavior —
// GREEN commit will swap to mime.FormatMediaType + control-char strip.
func BuildContentDisposition(disposition, filename string) string {
	return disposition + "; filename=\"" + filename + "\""
}
