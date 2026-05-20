// Package scheduler contains background job scheduling for the AI module.
package scheduler

import (
	"context"
	"log/slog"

	"github.com/go-co-op/gocron/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/usecases"
)

// IndexingScheduler runs periodic document indexing for the RAG pipeline.
type IndexingScheduler struct {
	scheduler        gocron.Scheduler
	embeddingUseCase *usecases.EmbeddingUseCase
	batchSize        int
	logger           *slog.Logger
	serverCtx        context.Context
}

// NewIndexingScheduler creates a new IndexingScheduler. serverCtx is the
// application lifecycle context — tick body derives its own ctx from it
// so SIGTERM cancels in-flight indexing rather than running к completion
// on context.Background() (issue #263 ADR-4).
func NewIndexingScheduler(
	serverCtx context.Context,
	embeddingUseCase *usecases.EmbeddingUseCase,
	batchSize int,
	logger *slog.Logger,
) (*IndexingScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	if batchSize <= 0 {
		batchSize = 10
	}

	return &IndexingScheduler{
		scheduler:        s,
		embeddingUseCase: embeddingUseCase,
		batchSize:        batchSize,
		logger:           logger,
		serverCtx:        serverCtx,
	}, nil
}

// Start starts the indexing scheduler (every 5 minutes).
func (is *IndexingScheduler) Start() error {
	_, err := is.scheduler.NewJob(
		gocron.CronJob("*/5 * * * *", false),
		gocron.NewTask(is.indexPendingDocuments),
	)
	if err != nil {
		return err
	}

	is.scheduler.Start()
	is.logger.Info("Indexing scheduler started - runs every 5 minutes")
	return nil
}

// Stop stops the scheduler gracefully.
func (is *IndexingScheduler) Stop() error {
	return is.scheduler.Shutdown()
}

func (is *IndexingScheduler) indexPendingDocuments() {
	ctx := is.serverCtx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		is.logger.Info("indexing scheduler tick skipped — server shutting down", "error", err)
		return
	}

	indexed, err := is.embeddingUseCase.IndexPendingDocuments(ctx, is.batchSize)
	if err != nil {
		is.logger.Error("indexing scheduler: failed to index pending documents", "error", err)
		return
	}

	if indexed > 0 {
		is.logger.Info("indexing scheduler: indexed documents", "count", indexed)
	}
}
