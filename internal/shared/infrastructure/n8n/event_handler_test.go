package n8n

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/ddd"
)

// captureServer spins an httptest server that records the inbound POST
// path and JSON body for assertions, then returns 200. Used so the
// handler tests exercise the real Client end-to-end and would catch
// any wiring break between WebhookEventHandler and Client.
func captureServer(t *testing.T) (*httptest.Server, func() (string, map[string]any)) {
	t.Helper()
	var (
		mu   sync.Mutex
		path string
		body map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	get := func() (string, map[string]any) {
		mu.Lock()
		defer mu.Unlock()
		return path, body
	}
	return server, get
}

// TestWebhookEventHandler_KnownEventTypes_RouteToConfiguredPaths is
// table-driven over every event type the handler is expected to know
// about. Adding a new entry to pathMap without a corresponding case
// here is a regression — keep this list in sync with event_handler.go.
//
// Each case asserts the inbound POST hits the expected webhook path
// AND the standard payload envelope (event_type, aggregate_id,
// occurred_at) lands intact. End-to-end through the real Client so a
// future refactor of TriggerAsync cannot silently bypass the test.
func TestWebhookEventHandler_KnownEventTypes_RouteToConfiguredPaths(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		wantPath  string
	}{
		{"document.created -> document-created", "document.created", "/webhook/document-created"},
		{"document.updated -> document-updated", "document.updated", "/webhook/document-updated"},
		{"risk_alert.detected -> risk-alert-detected", "risk_alert.detected", "/webhook/risk-alert-detected"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, captured := captureServer(t)
			client := NewClient(Config{WebhookURL: server.URL, Enabled: true}, newTestLogger())
			handler := NewWebhookEventHandler(client, newTestLogger())

			occurred := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
			event := ddd.BaseDomainEvent{
				EventType:   tc.eventType,
				AggregateID: "agg-42",
				OccurredAt:  occurred,
			}

			require.NoError(t, handler.Handle(event))

			// Wait for the async goroutine; 500ms is generous on a
			// localhost loopback and avoids time.Sleep-flake on CI.
			require.Eventually(t, func() bool {
				p, _ := captured()
				return p != ""
			}, 500*time.Millisecond, 10*time.Millisecond,
				"webhook server never received the request")

			gotPath, gotBody := captured()
			assert.Equal(t, tc.wantPath, gotPath)
			assert.Equal(t, tc.eventType, gotBody["event_type"])
			assert.Equal(t, "agg-42", gotBody["aggregate_id"])
			// occurred_at must round-trip through the RFC3339 formatter
			// the handler uses — the n8n side parses it as a date and a
			// format drift would silently break downstream workflows.
			assert.Equal(t, "2026-05-03T12:00:00Z", gotBody["occurred_at"])
		})
	}
}

// TestWebhookEventHandler_UnknownEventType_NoOp — events not in the
// pathMap must be silently dropped, no HTTP traffic. Otherwise every
// unrelated domain event in the system would hit n8n with an unknown
// path, polluting webhook logs and failing 404 in n8n.
func TestWebhookEventHandler_UnknownEventType_NoOp(t *testing.T) {
	var hit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{WebhookURL: server.URL, Enabled: true}, newTestLogger())
	handler := NewWebhookEventHandler(client, newTestLogger())

	event := ddd.BaseDomainEvent{
		EventType:   "unknown.thing",
		AggregateID: "x",
		OccurredAt:  time.Now(),
	}

	require.NoError(t, handler.Handle(event))
	time.Sleep(100 * time.Millisecond) // give any rogue goroutine a chance
	assert.False(t, hit, "unknown event type must not trigger any webhook")
}
