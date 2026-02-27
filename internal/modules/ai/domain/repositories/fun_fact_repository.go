// Package repositories contains repository interfaces for the AI module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// FunFactRepository defines the interface for fun fact persistence
type FunFactRepository interface {
	Create(ctx context.Context, fact *entities.FunFact) error
	BulkCreate(ctx context.Context, facts []entities.FunFact) error
	GetRandom(ctx context.Context) (*entities.FunFact, error)
	GetLeastUsed(ctx context.Context) (*entities.FunFact, error)
	IncrementUsedCount(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
}
