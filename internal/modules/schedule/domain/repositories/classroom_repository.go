package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// ClassroomFilter narrows ClassroomRepository.List/Count results.
type ClassroomFilter struct {
	Building    *string
	Type        *string
	MinCapacity *int
	IsAvailable *bool
}

// ClassroomRepository persists and queries Classroom aggregates.
type ClassroomRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.Classroom, error)
	List(ctx context.Context, filter ClassroomFilter, limit, offset int) ([]*entities.Classroom, error)
	Count(ctx context.Context, filter ClassroomFilter) (int64, error)
}
