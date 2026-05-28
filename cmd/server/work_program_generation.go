package main

import (
	"context"
	"fmt"

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
// the work_program usecases.DisciplineInfoProvider port — the cross-module
// glue that grounds РПД generation in the real учебный-план hour budget.
type disciplineInfoAdapter struct {
	reader disciplineItemReader
}

// compile-time check that the adapter satisfies the application port.
var _ wpUsecases.DisciplineInfoProvider = (*disciplineInfoAdapter)(nil)

// newDisciplineInfoAdapter wires the adapter over a discipline reader.
func newDisciplineInfoAdapter(reader disciplineItemReader) *disciplineInfoAdapter {
	return &disciplineInfoAdapter{reader: reader}
}

// GetDisciplineInfo loads the discipline item and maps it to the port DTO.
// The repository's not-found sentinel propagates unchanged.
func (a *disciplineInfoAdapter) GetDisciplineInfo(ctx context.Context, id int64) (wpUsecases.DisciplineInfo, error) {
	item, err := a.reader.GetByID(ctx, id)
	if err != nil {
		return wpUsecases.DisciplineInfo{}, fmt.Errorf("discipline lookup: %w", err)
	}
	return wpUsecases.DisciplineInfo{
		Name:           item.Title(),
		HoursLecture:   item.HoursLectures(),
		HoursPractice:  item.HoursPractice(),
		HoursLab:       item.HoursLab(),
		HoursSelfStudy: item.HoursSelf(),
		ControlForm:    controlFormLabel(item.ControlForm()),
	}, nil
}

// controlFormLabel renders the control-form enum as the Russian label the
// generation prompt expects. Unknown values fall back to the raw enum
// string so a new control form degrades gracefully rather than blanking.
func controlFormLabel(cf curEntities.ControlForm) string {
	switch cf {
	case curEntities.ControlFormExam:
		return "экзамен"
	case curEntities.ControlFormZachet:
		return "зачёт"
	case curEntities.ControlFormDifferentialZachet:
		return "дифференцированный зачёт"
	case curEntities.ControlFormCourseProject:
		return "курсовой проект"
	default:
		return cf.String()
	}
}
