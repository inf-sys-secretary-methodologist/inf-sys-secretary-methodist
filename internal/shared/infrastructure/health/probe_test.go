package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/health"
)

func TestProbe(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantExit   int
	}{
		{name: "200 OK is healthy", statusCode: http.StatusOK, wantExit: 0},
		{name: "204 No Content is healthy", statusCode: http.StatusNoContent, wantExit: 0},
		{name: "503 Service Unavailable is unhealthy", statusCode: http.StatusServiceUnavailable, wantExit: 1},
		{name: "404 Not Found is unhealthy", statusCode: http.StatusNotFound, wantExit: 1},
		{name: "500 Internal Server Error is unhealthy", statusCode: http.StatusInternalServerError, wantExit: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer srv.Close()

			if got := health.Probe(srv.URL, 2*time.Second); got != tt.wantExit {
				t.Errorf("Probe() exit = %d, want %d", got, tt.wantExit)
			}
		})
	}
}

func TestProbe_UnreachableHost(t *testing.T) {
	// Closed server: connection refused must map to the unhealthy exit code.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()

	if got := health.Probe(url, 1*time.Second); got != 1 {
		t.Errorf("Probe() on unreachable host exit = %d, want 1", got)
	}
}

func TestProbe_InvalidURL(t *testing.T) {
	if got := health.Probe("://not-a-url", time.Second); got != 1 {
		t.Errorf("Probe() on invalid URL exit = %d, want 1", got)
	}
}
