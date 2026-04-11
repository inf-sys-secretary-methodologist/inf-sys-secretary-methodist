package n8n

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *logging.Logger {
	return logging.NewLogger("error")
}

func TestTriggerWorkflow_Success(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webhook/test-path", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{WebhookURL: server.URL, Enabled: true}, newTestLogger())

	err := client.TriggerWorkflow(context.Background(), "test-path", map[string]any{
		"key": "value",
	})

	require.NoError(t, err)
	assert.Equal(t, "value", received["key"])
}

func TestTriggerWorkflow_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(Config{WebhookURL: server.URL, Enabled: true}, newTestLogger())

	err := client.TriggerWorkflow(context.Background(), "fail", map[string]any{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestTriggerWorkflow_Disabled(t *testing.T) {
	client := NewClient(Config{WebhookURL: "http://should-not-be-called", Enabled: false}, newTestLogger())

	err := client.TriggerWorkflow(context.Background(), "any", map[string]any{})

	assert.NoError(t, err)
}

func TestTriggerWorkflow_EmptyURL(t *testing.T) {
	client := NewClient(Config{WebhookURL: "", Enabled: true}, newTestLogger())

	err := client.TriggerWorkflow(context.Background(), "any", map[string]any{})

	assert.NoError(t, err)
}

func TestTriggerWorkflow_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{WebhookURL: server.URL, Enabled: true, Timeout: 100 * time.Millisecond}, newTestLogger())

	err := client.TriggerWorkflow(context.Background(), "slow", map[string]any{})

	assert.Error(t, err)
}

func TestTriggerAsync(t *testing.T) {
	var called atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{WebhookURL: server.URL, Enabled: true}, newTestLogger())
	client.TriggerAsync("async-test", map[string]any{"data": 1})

	// Wait for async goroutine
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), called.Load())
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected bool
	}{
		{"enabled with url", Config{WebhookURL: "http://localhost", Enabled: true}, true},
		{"disabled", Config{WebhookURL: "http://localhost", Enabled: false}, false},
		{"enabled no url", Config{WebhookURL: "", Enabled: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.cfg, newTestLogger())
			assert.Equal(t, tt.expected, client.IsEnabled())
		})
	}
}
