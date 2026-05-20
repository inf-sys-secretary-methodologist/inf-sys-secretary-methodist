package scheduler

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AI schedulers (FactScheduler + IndexingScheduler) depend on concrete
// *usecases.FunFactUseCase / *usecases.MoodUseCase / *usecases.
// EmbeddingUseCase pointers that themselves carry concrete dependencies
// (LLM clients, vector store, etc.). Without a narrow-port refactor of
// those use cases we cannot exercise the internal tick body
// (deliverDailyFact / indexPendingDocuments) — those funcs would NPE on
// the nil pointers. We do, however, cover the three structural funcs
// per scheduler: constructor wiring, Start (cron registration +
// scheduler.Start), Stop (Shutdown). Tick bodies remain 0% и documented
// as DIP-blocked in v0.153.6 release notes.
//
// Cron expressions in production code are hardcoded ("0 9 * * *",
// "*/5 * * * *") — valid strings, so Start cannot exercise the
// gocron.NewJob error branch from a unit test без monkey-patching.

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// --- FactScheduler ---

func TestNewFactScheduler_HappyPath(t *testing.T) {
	// All use-case pointers can be nil — constructor only stores them;
	// they're only dereferenced inside deliverDailyFact (which our cron
	// schedule "0 9 * * *" never fires during the test window).
	fs, err := NewFactScheduler(context.Background(), nil, nil, nil, nil, quietLogger())
	require.NoError(t, err)
	require.NotNil(t, fs)
	require.NotNil(t, fs.scheduler, "internal gocron scheduler must be wired")
}

func TestFactScheduler_StartStop(t *testing.T) {
	fs, err := NewFactScheduler(context.Background(), nil, nil, nil, nil, quietLogger())
	require.NoError(t, err)

	require.NoError(t, fs.Start(), "Start с valid cron must not error")
	// Brief sleep так gocron has time to boot worker goroutines before Shutdown.
	time.Sleep(20 * time.Millisecond)
	assert.NoError(t, fs.Stop(), "Shutdown must succeed")
}

// --- IndexingScheduler ---

func TestNewIndexingScheduler_HappyPath(t *testing.T) {
	is, err := NewIndexingScheduler(context.Background(), nil, 25, quietLogger())
	require.NoError(t, err)
	require.NotNil(t, is)
	assert.Equal(t, 25, is.batchSize, "explicit batchSize must be preserved")
}

func TestNewIndexingScheduler_DefaultBatchSize(t *testing.T) {
	cases := []struct {
		name      string
		inSize    int
		wantBatch int
	}{
		{name: "zero defaults to 10", inSize: 0, wantBatch: 10},
		{name: "negative defaults to 10", inSize: -5, wantBatch: 10},
		{name: "positive preserved", inSize: 50, wantBatch: 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			is, err := NewIndexingScheduler(context.Background(), nil, tc.inSize, quietLogger())
			require.NoError(t, err)
			assert.Equal(t, tc.wantBatch, is.batchSize)
		})
	}
}

func TestIndexingScheduler_StartStop(t *testing.T) {
	is, err := NewIndexingScheduler(context.Background(), nil, 10, quietLogger())
	require.NoError(t, err)

	require.NoError(t, is.Start())
	time.Sleep(20 * time.Millisecond)
	assert.NoError(t, is.Stop())
}

// TestFactScheduler_TickShortCircuitsOnCanceledCtx pins the lifecycle-ctx
// contract from issue #263 ADR-4: when serverCtx is canceled, the next
// scheduler tick body must short-circuit on ctx.Err() before touching any
// nil dependency. If a future refactor drops the ctx.Err() check the
// test will panic on the nil funFactUseCase deref instead of returning
// cleanly.
func TestFactScheduler_TickShortCircuitsOnCanceledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before any tick

	fs, err := NewFactScheduler(ctx, nil, nil, nil, nil, quietLogger())
	require.NoError(t, err)

	// Invoke the tick directly. Without the ctx.Err() guard this would
	// dereference the nil funFactUseCase and panic.
	assert.NotPanics(t, func() { fs.deliverDailyFact() })
}

func TestIndexingScheduler_TickShortCircuitsOnCanceledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	is, err := NewIndexingScheduler(ctx, nil, 10, quietLogger())
	require.NoError(t, err)

	assert.NotPanics(t, func() { is.indexPendingDocuments() })
}
