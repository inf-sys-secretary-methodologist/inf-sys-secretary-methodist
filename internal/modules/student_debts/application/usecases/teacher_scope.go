package usecases

import "context"

// TeacherScopeResolver is the narrow port that answers "which disciplines
// does this teacher own?" — the basis for scoping a teacher's view of the
// debt registry to their own disciplines.
//
// The student_debts module must not import curriculum/schedule directly
// (cross-module impls are forbidden); main.go wires a concrete adapter
// backed by whichever module owns the teacher↔discipline mapping. The
// returned ids are curriculum_section_items ids matching
// StudentDebt.DisciplineID. An empty slice (teacher owns no disciplines)
// is not an error — it yields an empty debt page.
type TeacherScopeResolver interface {
	DisciplineIDsForTeacher(ctx context.Context, teacherID int64) ([]int64, error)
}
