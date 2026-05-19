package persistence

// v0.153.7 Phase 6 backfill — closes uncovered rows.Err branches across
// DepartmentRepositoryPG.List / GetChildren, PositionRepositoryPG.List,
// UserProfileRepositoryPG.ListUsersWithOrg, plus the previously-uncovered
// `positionID.Valid` true branch inside GetProfileByID (existing test
// only exercised the deptID branch).
//
// All tests are sqlmock-driven и mirror the existing per-file conventions
// (regexp.QuoteMeta + WithArgs). No production change.

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepartmentList_RowsErrPropagates(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(deptCols).
		AddRow(int64(1), "Faculty", "FAC", "", nil, nil, true, now, now).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	_, err := repo.List(context.Background(), 10, 0, false)
	require.Error(t, err)
}

func TestDepartmentGetChildren_RowsErrPropagates(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(deptCols).
		AddRow(int64(2), "Child", "CHILD", "", int64(1), nil, true, now, now).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetChildren(context.Background(), 1)
	require.Error(t, err)
}

func TestPositionList_RowsErrPropagates(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(posCols).
		AddRow(int64(1), "Manager", "MGR", "", 1, true, now, now).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	_, err := repo.List(context.Background(), 10, 0, false)
	require.Error(t, err)
}

func TestProfileGetByID_PositionIDPopulated(t *testing.T) {
	// Covers `if positionID.Valid` branch in GetProfileByID — existing
	// happy-path test only set departmentID, leaving lines 72-74 uncovered.
	repo, mock := newProfileRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	posID := int64(7)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(profileCols).
			AddRow(int64(1), "u@test.com", "U", "user", "active",
				"", "", "",
				nil, "",
				posID, "Manager",
				now, now))

	user, err := repo.GetProfileByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, user.PositionID)
	assert.Equal(t, posID, *user.PositionID)
	assert.Nil(t, user.DepartmentID)
}

func TestProfileListUsersWithOrg_RowsErrPropagates(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(listProfileCols).
		AddRow(int64(1), "a@b.com", "A", "user", "active", "", "",
			nil, "", nil, "", now, now).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	_, err := repo.ListUsersWithOrg(context.Background(), nil, 10, 0)
	require.Error(t, err)
}
