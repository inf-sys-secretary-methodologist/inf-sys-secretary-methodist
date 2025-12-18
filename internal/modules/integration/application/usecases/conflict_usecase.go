package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// ConflictUseCase handles sync conflict operations
type ConflictUseCase struct {
	conflictRepo repositories.SyncConflictRepository
}

// NewConflictUseCase creates a new conflict use case
func NewConflictUseCase(conflictRepo repositories.SyncConflictRepository) *ConflictUseCase {
	return &ConflictUseCase{
		conflictRepo: conflictRepo,
	}
}

// List retrieves conflicts with filtering
func (uc *ConflictUseCase) List(ctx context.Context, req *dto.ConflictListRequest) (*dto.ConflictListResponse, error) {
	filter := entities.SyncConflictFilter{
		SyncLogID:  req.SyncLogID,
		EntityType: req.EntityType,
		Resolution: req.Resolution,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	conflicts, total, err := uc.conflictRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list conflicts: %w", err)
	}

	items := make([]*dto.SyncConflictDTO, len(conflicts))
	for i, conflict := range conflicts {
		items[i] = dto.FromSyncConflict(conflict)
	}

	return &dto.ConflictListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetByID retrieves a conflict by ID
func (uc *ConflictUseCase) GetByID(ctx context.Context, id int64) (*dto.SyncConflictDTO, error) {
	conflict, err := uc.conflictRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict: %w", err)
	}
	if conflict == nil {
		return nil, nil
	}
	return dto.FromSyncConflict(conflict), nil
}

// GetPending retrieves pending conflicts
func (uc *ConflictUseCase) GetPending(ctx context.Context, limit, offset int) (*dto.ConflictListResponse, error) {
	conflicts, total, err := uc.conflictRepo.GetPending(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending conflicts: %w", err)
	}

	items := make([]*dto.SyncConflictDTO, len(conflicts))
	for i, conflict := range conflicts {
		items[i] = dto.FromSyncConflict(conflict)
	}

	return &dto.ConflictListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetBySyncLogID retrieves conflicts for a specific sync log
func (uc *ConflictUseCase) GetBySyncLogID(ctx context.Context, syncLogID int64) ([]*dto.SyncConflictDTO, error) {
	conflicts, err := uc.conflictRepo.GetBySyncLogID(ctx, syncLogID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts by sync log: %w", err)
	}

	items := make([]*dto.SyncConflictDTO, len(conflicts))
	for i, conflict := range conflicts {
		items[i] = dto.FromSyncConflict(conflict)
	}

	return items, nil
}

// Resolve resolves a conflict
func (uc *ConflictUseCase) Resolve(ctx context.Context, id int64, userID int64, req *dto.ResolveConflictRequest) error {
	// Check if conflict exists
	conflict, err := uc.conflictRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get conflict: %w", err)
	}
	if conflict == nil {
		return fmt.Errorf("conflict not found")
	}

	// Check if already resolved
	if !conflict.IsPending() {
		return fmt.Errorf("conflict is already resolved")
	}

	// Resolve the conflict
	if err := uc.conflictRepo.Resolve(ctx, id, req.Resolution, userID, req.ResolvedData); err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	// Update notes if provided
	if req.Notes != "" {
		conflict.Notes = req.Notes
		if err := uc.conflictRepo.Update(ctx, conflict); err != nil {
			return fmt.Errorf("failed to update conflict notes: %w", err)
		}
	}

	return nil
}

// BulkResolve resolves multiple conflicts
func (uc *ConflictUseCase) BulkResolve(ctx context.Context, userID int64, req *dto.BulkResolveRequest) error {
	if len(req.IDs) == 0 {
		return fmt.Errorf("no conflict IDs provided")
	}

	if err := uc.conflictRepo.BulkResolve(ctx, req.IDs, req.Resolution, userID); err != nil {
		return fmt.Errorf("failed to bulk resolve conflicts: %w", err)
	}

	return nil
}

// GetStats retrieves conflict statistics
func (uc *ConflictUseCase) GetStats(ctx context.Context) (*dto.ConflictStatsDTO, error) {
	stats, err := uc.conflictRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict stats: %w", err)
	}
	return dto.FromConflictStats(stats), nil
}

// Delete deletes a conflict
func (uc *ConflictUseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.conflictRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete conflict: %w", err)
	}
	return nil
}
