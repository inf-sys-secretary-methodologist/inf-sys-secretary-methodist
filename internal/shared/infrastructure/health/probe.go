// Package health provides a self-contained liveness probe used by the
// `server -healthcheck` subcommand.
//
// The production image is built FROM scratch (see Dockerfile), so it has
// no shell, wget or curl for a Docker/compose healthcheck to call. The
// binary therefore probes its own /health endpoint and reports liveness
// through its exit code.
package health

import "time"

// Probe performs a GET against url and returns the process exit code:
// 0 when the endpoint responds with a 2xx status, 1 otherwise (non-2xx,
// connection error or timeout). timeout bounds the whole request.
func Probe(url string, timeout time.Duration) int {
	// RED stub — real implementation lands in the GREEN commit.
	return 1
}
