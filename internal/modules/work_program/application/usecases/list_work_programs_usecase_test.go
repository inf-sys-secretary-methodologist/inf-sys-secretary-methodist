package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// fakeListRepo captures the effective filter the use case dispatched.
type fakeListRepo struct {
	calledWith repositories.WorkProgramListFilter
	result     repositories.WorkProgramListResult
	err        error
	calls      int
}

func (f *fakeListRepo) List(_ context.Context, filter repositories.WorkProgramListFilter) (repositories.WorkProgramListResult, error) {
	f.calls++
	f.calledWith = filter
	return f.result, f.err
}

func TestNewListWorkProgramsUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewListWorkProgramsUseCase(nil, ...) did not panic")
		}
	}()
	NewListWorkProgramsUseCase(nil, &recordingAuditSink{})
}

func TestListWorkProgramsUseCase_AdminPassesFilterThrough(t *testing.T) {
	repo := &fakeListRepo{result: repositories.WorkProgramListResult{
		Items: []repositories.ListItem{{ID: 1}, {ID: 2}},
		Total: 2,
	}}
	uc := NewListWorkProgramsUseCase(repo, &recordingAuditSink{})

	statusFilter := "approved"
	res, err := uc.Execute(context.Background(), 99, "system_admin", ListWorkProgramsInput{
		Status:        &statusFilter,
		SpecialtyCode: "09.03.01",
		Limit:         100,
	})
	require.NoError(t, err)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, 2, res.Total)

	// Filter passed through verbatim — admin has no row-level restrictions.
	require.NotNil(t, repo.calledWith.Status)
	assert.Equal(t, domain.StatusApproved, *repo.calledWith.Status)
	assert.Equal(t, "09.03.01", repo.calledWith.SpecialtyCode)
	assert.Nil(t, repo.calledWith.AuthorID, "admin must not be author-restricted")
}

func TestListWorkProgramsUseCase_MethodistAndSecretaryPassThrough(t *testing.T) {
	for _, role := range []string{"methodist", "academic_secretary"} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeListRepo{}
			uc := NewListWorkProgramsUseCase(repo, &recordingAuditSink{})

			_, err := uc.Execute(context.Background(), 42, role, ListWorkProgramsInput{})
			require.NoError(t, err)
			assert.Nil(t, repo.calledWith.AuthorID,
				"role %q sees all WPs — no author restriction", role)
			assert.Nil(t, repo.calledWith.Status,
				"role %q sees every status — no status force", role)
		})
	}
}

func TestListWorkProgramsUseCase_TeacherIsForcedToOwn(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListWorkProgramsUseCase(repo, &recordingAuditSink{})

	// Teacher provided an explicit author_id of someone else — must be overridden.
	otherAuthor := int64(999)
	_, err := uc.Execute(context.Background(), 7, "teacher", ListWorkProgramsInput{
		AuthorID: &otherAuthor,
	})
	require.NoError(t, err)
	require.NotNil(t, repo.calledWith.AuthorID,
		"teacher must always have author filter forced to own id")
	assert.Equal(t, int64(7), *repo.calledWith.AuthorID,
		"teacher's author filter must equal actor id, not the requested value")
}

func TestListWorkProgramsUseCase_StudentIsForcedToApproved(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListWorkProgramsUseCase(repo, &recordingAuditSink{})

	// Student tried to query drafts — must be overridden to approved.
	drafts := "draft"
	_, err := uc.Execute(context.Background(), 200, "student", ListWorkProgramsInput{
		Status: &drafts,
	})
	require.NoError(t, err)
	require.NotNil(t, repo.calledWith.Status,
		"student must always have status forced to approved")
	assert.Equal(t, domain.StatusApproved, *repo.calledWith.Status)
}

func TestListWorkProgramsUseCase_UnknownRoleDenied(t *testing.T) {
	repo := &fakeListRepo{}
	audit := &recordingAuditSink{}
	uc := NewListWorkProgramsUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 1, "unknown_role", ListWorkProgramsInput{})
	assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
		"unknown role must be denied, got %v", err)
	assert.Zero(t, repo.calls, "repo.List must not be called on denied role")

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.list_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
}

func TestListWorkProgramsUseCase_PaginationDefaultsAndClamp(t *testing.T) {
	cases := []struct {
		name      string
		inLimit   int
		inOffset  int
		outLimit  int
		outOffset int
	}{
		{"zero_limit_defaults_to_50", 0, 0, defaultListLimit, 0},
		{"negative_limit_defaults_to_50", -10, 0, defaultListLimit, 0},
		{"negative_offset_clamped_to_0", 25, -5, 25, 0},
		{"over_max_limit_clamped_to_200", 500, 100, maxListLimit, 100},
		{"valid_values_passed_through", 75, 25, 75, 25},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeListRepo{}
			uc := NewListWorkProgramsUseCase(repo, &recordingAuditSink{})

			_, err := uc.Execute(context.Background(), 99, "system_admin",
				ListWorkProgramsInput{Limit: tc.inLimit, Offset: tc.inOffset})
			require.NoError(t, err)
			assert.Equal(t, tc.outLimit, repo.calledWith.Limit)
			assert.Equal(t, tc.outOffset, repo.calledWith.Offset)
		})
	}
}

func TestListWorkProgramsUseCase_RepoErrorPropagates(t *testing.T) {
	repo := &fakeListRepo{err: errors.New("conn refused")}
	uc := NewListWorkProgramsUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), 99, "system_admin", ListWorkProgramsInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
}
