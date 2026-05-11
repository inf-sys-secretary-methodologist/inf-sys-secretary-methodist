package logging

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	// NOT NULL DEFAULT, but the writer should also send a serialisable
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
