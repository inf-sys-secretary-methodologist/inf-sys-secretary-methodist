package main

import (
	"context"
	"errors"

	curEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
)

// disciplineItemReader is the slice of the curriculum DisciplineItem
// repository the generation adapter needs. Declaring it here (composition
// root) keeps the work_program module free of any curriculum import.
type disciplineItemReader interface {
	GetByID(ctx context.Context, id int64) (*curEntities.DisciplineItem, error)
}

// disciplineInfoAdapter adapts the curriculum DisciplineItem repository to
// the work_program usecases.DisciplineInfoProvider port. STUB — behavior
// lands in the GREEN commit.
type disciplineInfoAdapter struct{}

// compile-time check that the adapter satisfies the application port.
var _ wpUsecases.DisciplineInfoProvider = (*disciplineInfoAdapter)(nil)

// newDisciplineInfoAdapter is the STUB constructor.
func newDisciplineInfoAdapter(reader disciplineItemReader) *disciplineInfoAdapter {
	return &disciplineInfoAdapter{}
}

// GetDisciplineInfo is the STUB entrypoint.
func (a *disciplineInfoAdapter) GetDisciplineInfo(ctx context.Context, id int64) (wpUsecases.DisciplineInfo, error) {
	return wpUsecases.DisciplineInfo{}, errors.New("cmd/server: GetDisciplineInfo not implemented")
}

// controlFormLabel is the STUB mapper.
func controlFormLabel(cf curEntities.ControlForm) string {
	return ""
}
