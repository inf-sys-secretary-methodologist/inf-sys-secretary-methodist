// Package usecases contains business logic for the users module.
package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// PositionUseCase handles position management business logic.
type PositionUseCase struct {
	positionRepo repositories.PositionRepository
	auditLogger  *logging.AuditLogger
}

// NewPositionUseCase creates a new position use case.
func NewPositionUseCase(
	positionRepo repositories.PositionRepository,
	auditLogger *logging.AuditLogger,
) *PositionUseCase {
	return &PositionUseCase{
		positionRepo: positionRepo,
		auditLogger:  auditLogger,
	}
}

// CreatePosition creates a new position.
func (uc *PositionUseCase) CreatePosition(ctx context.Context, input *dto.CreatePositionInput) (*entities.Position, error) {
	position := entities.NewPosition(
		input.Name,
		input.Code,
		input.Description,
		input.Level,
	)

	err := uc.positionRepo.Create(ctx, position)
	if err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "position", map[string]interface{}{
			"position_id": position.ID,
			"name":        position.Name,
			"code":        position.Code,
		})
	}

	return position, nil
}

// GetPosition retrieves a position by ID.
func (uc *PositionUseCase) GetPosition(ctx context.Context, id int64) (*entities.Position, error) {
	return uc.positionRepo.GetByID(ctx, id)
}

// GetPositionByCode retrieves a position by code.
func (uc *PositionUseCase) GetPositionByCode(ctx context.Context, code string) (*entities.Position, error) {
	return uc.positionRepo.GetByCode(ctx, code)
}

// ListPositions retrieves a paginated list of positions.
func (uc *PositionUseCase) ListPositions(ctx context.Context, page, limit int, activeOnly bool) (*dto.PositionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	positions, err := uc.positionRepo.List(ctx, limit, offset, activeOnly)
	if err != nil {
		return nil, err
	}

	total, err := uc.positionRepo.Count(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.PositionListResponse{
		Positions:  positions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdatePosition updates an existing position.
func (uc *PositionUseCase) UpdatePosition(ctx context.Context, id int64, input *dto.UpdatePositionInput) (*entities.Position, error) {
	position, err := uc.positionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	position.Name = input.Name
	position.Code = input.Code
	position.Description = input.Description
	position.Level = input.Level
	if input.IsActive != nil {
		position.IsActive = *input.IsActive
	}
	position.UpdatedAt = time.Now()

	err = uc.positionRepo.Update(ctx, position)
	if err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "update", "position", map[string]interface{}{
			"position_id": id,
			"name":        position.Name,
			"code":        position.Code,
		})
	}

	return position, nil
}

// DeletePosition deletes a position.
func (uc *PositionUseCase) DeletePosition(ctx context.Context, id int64) error {
	// Verify position exists
	position, err := uc.positionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = uc.positionRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "delete", "position", map[string]interface{}{
			"position_id": id,
			"name":        position.Name,
		})
	}

	return nil
}
