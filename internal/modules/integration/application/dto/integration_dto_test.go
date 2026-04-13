package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromExternalEmployee(t *testing.T) {
	now := time.Now()
	empDate := now.Add(-365 * 24 * time.Hour)
	localUserID := int64(100)
	emp := &entities.ExternalEmployee{
		ID:             1,
		ExternalID:     "EXT-001",
		Code:           "E001",
		FirstName:      "Ivan",
		LastName:       "Petrov",
		MiddleName:     "Ivanovich",
		Email:          "ivan@example.com",
		Phone:          "+7999",
		Position:       "Teacher",
		Department:     "CS",
		EmploymentDate: &empDate,
		IsActive:       true,
		LocalUserID:    &localUserID,
		LastSyncAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	dto := FromExternalEmployee(emp)

	require.NotNil(t, dto)
	assert.Equal(t, int64(1), dto.ID)
	assert.Equal(t, "EXT-001", dto.ExternalID)
	assert.Equal(t, "E001", dto.Code)
	assert.Equal(t, "Ivan", dto.FirstName)
	assert.Equal(t, "Petrov", dto.LastName)
	assert.Equal(t, "Ivanovich", dto.MiddleName)
	assert.Contains(t, dto.FullName, "Petrov")
	assert.Equal(t, "ivan@example.com", dto.Email)
	assert.Equal(t, "Teacher", dto.Position)
	assert.Equal(t, "CS", dto.Department)
	assert.True(t, dto.IsActive)
	assert.True(t, dto.IsLinked)
	assert.Equal(t, &localUserID, dto.LocalUserID)
}

func TestFromExternalEmployee_NotLinked(t *testing.T) {
	emp := &entities.ExternalEmployee{
		ID:         1,
		ExternalID: "EXT-002",
		Code:       "E002",
		FirstName:  "Anna",
		LastName:   "Sidorova",
		IsActive:   true,
		LastSyncAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	dto := FromExternalEmployee(emp)
	assert.False(t, dto.IsLinked)
	assert.Nil(t, dto.LocalUserID)
}

func TestFromExternalStudent(t *testing.T) {
	now := time.Now()
	enrollDate := now.Add(-180 * 24 * time.Hour)
	localUserID := int64(200)
	student := &entities.ExternalStudent{
		ID:             1,
		ExternalID:     "STU-001",
		Code:           "S001",
		FirstName:      "Dmitry",
		LastName:       "Kuznetsov",
		MiddleName:     "Petrovich",
		Email:          "dmitry@example.com",
		GroupName:      "CS-101",
		Faculty:        "IT",
		Specialty:      "Computer Science",
		Course:         2,
		StudyForm:      "full-time",
		EnrollmentDate: &enrollDate,
		Status:         "active",
		IsActive:       true,
		LocalUserID:    &localUserID,
		LastSyncAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	dto := FromExternalStudent(student)

	require.NotNil(t, dto)
	assert.Equal(t, int64(1), dto.ID)
	assert.Equal(t, "STU-001", dto.ExternalID)
	assert.Contains(t, dto.FullName, "Kuznetsov")
	assert.Equal(t, "CS-101", dto.GroupName)
	assert.Equal(t, "IT", dto.Faculty)
	assert.Equal(t, 2, dto.Course)
	assert.Equal(t, "active", dto.Status)
	assert.True(t, dto.IsActive)
	assert.True(t, dto.IsLinked)
}

func TestFromSyncLog(t *testing.T) {
	now := time.Now()
	completed := now.Add(time.Minute)
	log := &entities.SyncLog{
		ID:             1,
		EntityType:     entities.SyncEntityEmployee,
		Direction:      entities.SyncDirectionImport,
		Status:         entities.SyncStatusCompleted,
		StartedAt:      now,
		CompletedAt:    &completed,
		TotalRecords:   100,
		ProcessedCount: 100,
		SuccessCount:   95,
		ErrorCount:     5,
		ConflictCount:  2,
		CreatedAt:      now,
	}

	dto := FromSyncLog(log)

	require.NotNil(t, dto)
	assert.Equal(t, int64(1), dto.ID)
	assert.Equal(t, entities.SyncEntityEmployee, dto.EntityType)
	assert.Equal(t, entities.SyncDirectionImport, dto.Direction)
	assert.Equal(t, entities.SyncStatusCompleted, dto.Status)
	assert.Equal(t, 100, dto.TotalRecords)
	assert.Equal(t, 95, dto.SuccessCount)
	assert.Equal(t, 5, dto.ErrorCount)
	assert.Equal(t, 2, dto.ConflictCount)
}

func TestFromSyncStats(t *testing.T) {
	now := time.Now()
	stats := &entities.SyncStats{
		TotalSyncs:      50,
		SuccessfulSyncs: 45,
		FailedSyncs:     5,
		TotalRecords:    5000,
		TotalConflicts:  10,
		LastSyncAt:      now,
	}

	dto := FromSyncStats(stats)

	require.NotNil(t, dto)
	assert.Equal(t, int64(50), dto.TotalSyncs)
	assert.Equal(t, int64(45), dto.SuccessfulSyncs)
	assert.Equal(t, int64(5), dto.FailedSyncs)
	assert.Equal(t, int64(5000), dto.TotalRecords)
	assert.Equal(t, int64(10), dto.TotalConflicts)
}

func TestFromSyncConflict(t *testing.T) {
	now := time.Now()
	resolvedBy := int64(42)
	resolvedAt := now
	conflict := &entities.SyncConflict{
		ID:             1,
		SyncLogID:      10,
		EntityType:     entities.SyncEntityEmployee,
		EntityID:       "EMP-001",
		LocalData:      `{"name":"old"}`,
		ExternalData:   `{"name":"new"}`,
		ConflictType:   "field_mismatch",
		ConflictFields: []string{"name", "email"},
		Resolution:     entities.ConflictResolutionPending,
		ResolvedBy:     &resolvedBy,
		ResolvedAt:     &resolvedAt,
		Notes:          "needs review",
		CreatedAt:      now,
	}

	dto := FromSyncConflict(conflict)

	require.NotNil(t, dto)
	assert.Equal(t, int64(1), dto.ID)
	assert.Equal(t, int64(10), dto.SyncLogID)
	assert.Equal(t, entities.SyncEntityEmployee, dto.EntityType)
	assert.Equal(t, "EMP-001", dto.EntityID)
	assert.Equal(t, "field_mismatch", dto.ConflictType)
	assert.Equal(t, []string{"name", "email"}, dto.ConflictFields)
	assert.Equal(t, entities.ConflictResolutionPending, dto.Resolution)
	assert.Equal(t, "needs review", dto.Notes)
}

func TestFromConflictStats(t *testing.T) {
	stats := &entities.ConflictStats{
		TotalConflicts:    20,
		PendingConflicts:  5,
		ResolvedConflicts: 15,
		ByEntityType: map[entities.SyncEntityType]int64{
			entities.SyncEntityEmployee: 10,
			entities.SyncEntityStudent:  10,
		},
	}

	dto := FromConflictStats(stats)

	require.NotNil(t, dto)
	assert.Equal(t, int64(20), dto.TotalConflicts)
	assert.Equal(t, int64(5), dto.PendingConflicts)
	assert.Equal(t, int64(15), dto.ResolvedConflicts)
	assert.Len(t, dto.ByEntityType, 2)
}
