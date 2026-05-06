package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitAudit_NilSinkIsNoOp(t *testing.T) {
	// Must not panic. v0.116.0 use cases were already tolerant of a
	// nil sink via per-method guards; the package-level helper makes
	// that contract explicit and centralised.
	emitAudit(nil, context.Background(), "curriculum.test", map[string]any{"k": "v"})
}

func TestEmitAudit_NonNilSinkForwardsActionAndFields(t *testing.T) {
	sink := &recordingAuditSink{}
	fields := map[string]any{
		"actor_user_id": int64(7),
		"curriculum_id": int64(42),
		"reason":        "invalid",
	}
	emitAudit(sink, context.Background(), "curriculum.test_event", fields)

	require.Len(t, sink.events, 1)
	ev := sink.events[0]
	assert.Equal(t, "curriculum.test_event", ev.Action)
	assert.Equal(t, "curriculum", ev.Resource,
		"helper must use the package-level auditResource constant")
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(42), ev.Fields["curriculum_id"])
	assert.Equal(t, "invalid", ev.Fields["reason"])
}
