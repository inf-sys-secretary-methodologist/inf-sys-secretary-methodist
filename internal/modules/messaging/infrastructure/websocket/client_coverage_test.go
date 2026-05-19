package websocket

// v0.153.8 Phase 6 backfill — closes checkOrigin branches in
// messaging/infrastructure/websocket/client.go. Pure-Go-stdlib tests,
// no network. ServeWs / ReadPump / WritePump remain uncovered — those
// require gorilla websocket dialer + Hub fakes (heavier scope; defer).

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckOrigin_EmptyOriginHeaderAllowed(t *testing.T) {
	// Empty Origin = same-origin or non-browser client (curl/postman)
	// — allow by default.
	r := &http.Request{Header: http.Header{}}
	assert.True(t, checkOrigin(r))
}

func TestCheckOrigin_AllowedOriginMatches(t *testing.T) {
	// Default config from .env defaults loads CORS_ALLOWED_ORIGINS =
	// http://localhost:3000.
	r := &http.Request{Header: http.Header{}}
	r.Header.Set("Origin", "http://localhost:3000")
	assert.True(t, checkOrigin(r))
}

func TestCheckOrigin_OriginMismatchDenied(t *testing.T) {
	r := &http.Request{Header: http.Header{}}
	r.Header.Set("Origin", "https://evil.example.com")
	assert.False(t, checkOrigin(r))
}

func TestCheckOrigin_InvalidOriginURLDenied(t *testing.T) {
	// url.Parse rejects strings с control characters в host.
	r := &http.Request{Header: http.Header{}}
	r.Header.Set("Origin", "http://exa\x00mple.com")
	assert.False(t, checkOrigin(r))
}

func TestCheckOrigin_WildcardAllowed(t *testing.T) {
	// t.Setenv handles save/restore + nil-check automatically (Go 1.17+).
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	r := &http.Request{Header: http.Header{}}
	r.Header.Set("Origin", "https://any.origin.example.com")
	assert.True(t, checkOrigin(r))
}
