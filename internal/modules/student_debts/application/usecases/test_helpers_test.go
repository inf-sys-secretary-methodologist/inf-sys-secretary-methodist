package usecases_test

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// fakeDebtRepo is a function-backed double for the read ports. Unset
// fields panic if called, surfacing unexpected repo traffic in a test.
type fakeDebtRepo struct {
	getByID func(ctx context.Context, id int64) (*entities.StudentDebt, error)
	list    func(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error)
}

func (f *fakeDebtRepo) GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error) {
	return f.getByID(ctx, id)
}

func (f *fakeDebtRepo) List(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	return f.list(ctx, filter)
}

// fakeTeacherScope returns a fixed discipline set (or error) for any
// teacher id, and records the id it was asked about.
type fakeTeacherScope struct {
	ids       []int64
	err       error
	askedWith int64
	called    bool
}

func (f *fakeTeacherScope) DisciplineIDsForTeacher(_ context.Context, teacherID int64) ([]int64, error) {
	f.called = true
	f.askedWith = teacherID
	return f.ids, f.err
}

// recordingAudit captures emitted audit events for assertions.
type recordingAudit struct {
	events []auditEvent
}

type auditEvent struct {
	action   string
	resource string
	fields   map[string]any
}

func (a *recordingAudit) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	a.events = append(a.events, auditEvent{action: action, resource: resource, fields: fields})
}

// debtWith builds a persisted-shape debt with the given id, owning
// student and discipline (nil pointers stay unset). Status is open.
func debtWith(id int64, studentUserID, disciplineID *int64) *entities.StudentDebt {
	d, err := entities.NewStudentDebt("Иванов Иван", "ИВТ-21", "Базы данных", 3, entities.ControlFormExam)
	if err != nil {
		panic(err)
	}
	d.ID = id
	d.StudentUserID = studentUserID
	d.DisciplineID = disciplineID
	return d
}

func ptr[T any](v T) *T { return &v }
