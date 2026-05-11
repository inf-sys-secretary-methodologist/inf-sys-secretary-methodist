package logging

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// captureStdout swaps os.Stdout for a pipe, runs fn, restores stdout
// and returns everything fn wrote to the original target. Used to
// verify error-level log lines emitted by the Logger backend (which is
// hard-wired to os.Stdout).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	origStdout := os.Stdout
	origLog := log.Default().Writer()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	log.SetOutput(w)
	t.Cleanup(func() {
		os.Stdout = origStdout
		log.SetOutput(origLog)
	})

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-done
}

// fakeAuditWriter records every Write call so the persistence path can
// be asserted without a real *sql.DB. Optional err returned to simulate
// DB failure (ADR-3 fallback semantics).
type fakeAuditWriter struct {
	calls   []*AuditLog
	err     error
	receive context.Context //nolint:containedctx // recording received ctx for assertions
}

func (f *fakeAuditWriter) Write(ctx context.Context, log *AuditLog) error {
	f.calls = append(f.calls, log)
	f.receive = ctx
	return f.err
}

func TestAuditLogger_WithRepository_PersistsLog(t *testing.T) {
	spy := &fakeAuditWriter{}
	al := NewAuditLogger(NewLogger("debug")).WithRepository(spy)

	ctx := context.WithValue(context.Background(), ContextKeyUserID, int64(42))
	ctx = context.WithValue(ctx, ContextKeyIPAddress, "10.0.0.1")
	ctx = context.WithValue(ctx, ContextKeyCorrelationID, "trace-7c")

	al.LogAuditEvent(ctx, "curriculum.created", "curriculum", map[string]interface{}{
		"curriculum_id": int64(7),
		"title":         "ИС-21",
	})

	require.Len(t, spy.calls, 1, "Write must be called exactly once for one LogAuditEvent")
	got := spy.calls[0]
	require.Equal(t, "curriculum.created", got.Action)
	require.Equal(t, "curriculum", got.Resource)

	require.NotNil(t, got.ActorUserID)
	require.Equal(t, int64(42), *got.ActorUserID)
	require.NotNil(t, got.ActorIP)
	require.Equal(t, "10.0.0.1", *got.ActorIP)
	require.NotNil(t, got.CorrelationID)
	require.Equal(t, "trace-7c", *got.CorrelationID)

	require.Equal(t, int64(7), got.Fields["curriculum_id"])
	require.Equal(t, "ИС-21", got.Fields["title"])
}

func TestAuditLogger_WithRepository_NoActor_LeavesNullable(t *testing.T) {
	spy := &fakeAuditWriter{}
	al := NewAuditLogger(NewLogger("debug")).WithRepository(spy)

	al.LogAuditEvent(context.Background(), "system.startup", "system", nil)

	require.Len(t, spy.calls, 1)
	got := spy.calls[0]
	require.Nil(t, got.ActorUserID, "missing user_id must persist as SQL NULL via *int64 nil")
	require.Nil(t, got.ActorIP)
	require.Nil(t, got.CorrelationID)
	// Empty Fields must persist as empty JSONB ({}), not nil — DB has
	// NOT NULL DEFAULT, but the writer should also send a serializable
	// empty map rather than panic on Marshal(nil).
	require.NotNil(t, got.Fields)
}

func TestAuditLogger_WithRepository_WriteError_DoesNotPropagate(t *testing.T) {
	// ADR-3: a failed audit persist must not crash the calling code.
	// LogAuditEvent returns nothing, so "does not propagate" reduces to
	// "does not panic" + attempt was made + structured log still emitted.
	spy := &fakeAuditWriter{err: errors.New("conn refused")}
	al := NewAuditLogger(NewLogger("debug")).WithRepository(spy)

	require.NotPanics(t, func() {
		al.LogAuditEvent(context.Background(), "auth.login", "session", map[string]interface{}{
			"email": "user@example.com",
		})
	})

	require.Len(t, spy.calls, 1, "writer was attempted exactly once even when it failed")
}

func TestAuditLogger_NoRepository_BackwardsCompatible(t *testing.T) {
	// Backwards compatibility per ADR-7: existing call sites using
	// NewAuditLogger(logger) WITHOUT .WithRepository keep emitting only
	// to the structured log. No Writer is invoked because none is set.
	spy := &fakeAuditWriter{}
	al := NewAuditLogger(NewLogger("debug")) // intentionally no WithRepository

	require.NotPanics(t, func() {
		al.LogAuditEvent(context.Background(), "x", "y", nil)
	})

	// Spy stays untouched — there is no link from `al` to it.
	require.Empty(t, spy.calls)
}

func TestAuditLogger_WithRepository_ChainsAndReturnsReceiver(t *testing.T) {
	// WithRepository must return the same *AuditLogger so consumers can
	// chain: NewAuditLogger(...).WithRepository(repo). Returning a copy
	// would silently break call sites that store the original.
	spy := &fakeAuditWriter{}
	original := NewAuditLogger(NewLogger("debug"))
	chained := original.WithRepository(spy)

	require.Same(t, original, chained,
		"WithRepository must return the receiver, not a copy")
}

// T2-3 — extractor type-mismatch table-driven coverage (reviewer round-1).
// The persist() helpers ignore ctx values whose runtime type does not
// match the expected one; this test pins that fall-through so future
// refactors do not silently start emitting junk values into audit_logs.

func TestExtractActorUserID_TypeMismatchFallsThrough(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		want *int64
	}{
		{"nil ctx-value", context.Background(), nil},
		{"int instead of int64",
			context.WithValue(context.Background(), ContextKeyUserID, int(42)), nil},
		{"string instead of int64",
			context.WithValue(context.Background(), ContextKeyUserID, "42"), nil},
		{"valid int64",
			context.WithValue(context.Background(), ContextKeyUserID, int64(42)), int64Ptr(42)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractActorUserID(tc.ctx)
			if tc.want == nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, *tc.want, *got)
			}
		})
	}
}

func TestExtractActorIP_TypeMismatchFallsThrough(t *testing.T) {
	ip := "10.0.0.5"
	cases := []struct {
		name string
		ctx  context.Context
		want *string
	}{
		{"nil ctx-value", context.Background(), nil},
		{"int instead of string",
			context.WithValue(context.Background(), ContextKeyIPAddress, 12345), nil},
		{"empty string falls through",
			context.WithValue(context.Background(), ContextKeyIPAddress, ""), nil},
		{"valid non-empty string",
			context.WithValue(context.Background(), ContextKeyIPAddress, ip), &ip},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractActorIP(tc.ctx)
			if tc.want == nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, *tc.want, *got)
			}
		})
	}
}

func TestExtractCorrelationID_TypeMismatchFallsThrough(t *testing.T) {
	cid := "trace-7c"
	cases := []struct {
		name string
		ctx  context.Context
		want *string
	}{
		{"nil ctx-value", context.Background(), nil},
		{"bool instead of string",
			context.WithValue(context.Background(), ContextKeyCorrelationID, true), nil},
		{"empty string falls through",
			context.WithValue(context.Background(), ContextKeyCorrelationID, ""), nil},
		{"valid non-empty string",
			context.WithValue(context.Background(), ContextKeyCorrelationID, cid), &cid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractCorrelationID(tc.ctx)
			if tc.want == nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, *tc.want, *got)
			}
		})
	}
}

func int64Ptr(v int64) *int64 { return &v }

// T2-4 — ADR-3 fallback log-line assertion. Reviewer round-1 noted the
// commit message claimed "Writer error is logged at error level with
// action+resource+cause" without test coverage. This captures the
// stdout emitted by the underlying Logger when the writer fails and
// asserts the diagnostic line shape.

func TestAuditLogger_WithRepository_WriteError_EmitsErrorLogWithCause(t *testing.T) {
	spy := &fakeAuditWriter{err: errors.New("conn refused")}

	// Logger captures os.Stdout reference at construction time, so we
	// must build it INSIDE the captureStdout closure for the pipe swap
	// to actually intercept its output.
	out := captureStdout(t, func() {
		al := NewAuditLogger(NewLogger("debug")).WithRepository(spy)
		al.LogAuditEvent(context.Background(), "auth.login", "session", nil)
	})

	require.Contains(t, out, "Audit event persistence failed",
		"writer failure must surface as a dedicated error-level diagnostic line")
	require.Contains(t, out, `"level":"ERROR"`,
		"diagnostic line must carry ERROR level so SRE alerting can target it")
	require.Contains(t, out, "auth.login", "action must be in the diagnostic for triage")
	require.Contains(t, out, "session", "resource must be in the diagnostic for triage")
	require.Contains(t, out, "conn refused",
		"underlying writer error must be in the diagnostic as cause")

	// Sanity check: the happy-path Audit event line is also still
	// emitted — the failure does not suppress the structured log emit.
	require.True(t, strings.Contains(out, `"message":"Audit event"`),
		"structured log emission must precede and survive the writer failure path")
}
