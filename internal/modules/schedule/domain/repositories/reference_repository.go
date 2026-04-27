package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type ReferenceRepository interface {
	ListStudentGroups(ctx context.Context, limit, offset int) ([]*entities.StudentGroup, error)
	ListDisciplines(ctx context.Context, limit, offset int) ([]*entities.Discipline, error)
	ListSemesters(ctx context.Context, activeOnly bool) ([]*entities.Semester, error)
	ListLessonTypes(ctx context.Context) ([]*entities.LessonType, error)
	GetActiveSemester(ctx context.Context) (*entities.Semester, error)
}
