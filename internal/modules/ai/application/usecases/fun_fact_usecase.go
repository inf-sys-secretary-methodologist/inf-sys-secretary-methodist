// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
)

// FunFactUseCase handles fun fact operations
type FunFactUseCase struct {
	factRepo    repositories.FunFactRepository
	personality *services.PersonalityService
}

// NewFunFactUseCase creates a new FunFactUseCase
func NewFunFactUseCase(
	factRepo repositories.FunFactRepository,
	personality *services.PersonalityService,
) *FunFactUseCase {
	return &FunFactUseCase{
		factRepo:    factRepo,
		personality: personality,
	}
}

// GetRandomFact returns a random fun fact
func (uc *FunFactUseCase) GetRandomFact(ctx context.Context) (*dto.FunFactResponse, error) {
	fact, err := uc.factRepo.GetRandom(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get random fact: %w", err)
	}

	if fact == nil {
		return &dto.FunFactResponse{
			Content:  "Методыч пока не нашёл интересных фактов. Загляните позже!",
			Category: "system",
		}, nil
	}

	// Increment used count
	if err := uc.factRepo.IncrementUsedCount(ctx, fact.ID); err != nil {
		// Non-critical error, log but continue
		_ = err
	}

	return &dto.FunFactResponse{
		ID:       fact.ID,
		Content:  fact.Content,
		Category: fact.Category,
		Source:   fact.Source,
	}, nil
}
