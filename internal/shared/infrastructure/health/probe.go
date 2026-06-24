// Package health provides a self-contained liveness probe used by the
// `server -healthcheck` subcommand.
//
// The production image is built FROM scratch (see Dockerfile), so it has
// no shell, wget or curl for a Docker/compose healthcheck to call. The
// binary therefore probes its own /health endpoint and reports liveness
// through its exit code.
package health

import (
	"context"
	"net/http"
	"time"
)

// Probe performs a GET against url and returns the process exit code:
// 0 when the endpoint responds with a 2xx status, 1 otherwise (non-2xx,
// connection error or timeout). timeout bounds the whole request.
func Probe(url string, timeout time.Duration) int {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 1
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return 0
	}
	return 1
}
