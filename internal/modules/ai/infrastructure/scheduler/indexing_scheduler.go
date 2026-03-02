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
}

// NewIndexingScheduler creates a new IndexingScheduler.
func NewIndexingScheduler(
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
	ctx := context.Background()

	indexed, err := is.embeddingUseCase.IndexPendingDocuments(ctx, is.batchSize)
	if err != nil {
		is.logger.Error("indexing scheduler: failed to index pending documents", "error", err)
		return
	}

	if indexed > 0 {
		is.logger.Info("indexing scheduler: indexed documents", "count", indexed)
	}
}
