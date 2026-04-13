package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToProjectOutput(t *testing.T) {
	now := time.Now()
	desc := "A project"
	startDate := now.Add(-7 * 24 * time.Hour)
	endDate := now.Add(30 * 24 * time.Hour)

	project := &entities.Project{
		ID:          1,
		Name:        "Project A",
		Description: &desc,
		OwnerID:     42,
		Status:      domain.ProjectStatusActive,
		StartDate:   &startDate,
		EndDate:     &endDate,
		Tasks:       []entities.Task{{}, {}, {}}, // 3 tasks
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	output := ToProjectOutput(project)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Project A", output.Name)
	assert.Equal(t, &desc, output.Description)
	assert.Equal(t, int64(42), output.OwnerID)
	assert.Equal(t, "active", output.Status)
	assert.Equal(t, &startDate, output.StartDate)
	assert.Equal(t, &endDate, output.EndDate)
	assert.Equal(t, 3, output.TaskCount)
}

func TestToProjectOutput_NoTasks(t *testing.T) {
	project := &entities.Project{
		ID:      1,
		Name:    "Empty",
		OwnerID: 1,
		Status:  domain.ProjectStatusPlanning,
	}

	output := ToProjectOutput(project)
	assert.Equal(t, 0, output.TaskCount)
}

func TestProjectFilterInput_ToProjectFilter(t *testing.T) {
	ownerID := int64(42)
	status := "active"
	search := "test"

	f := &ProjectFilterInput{
		OwnerID: &ownerID,
		Status:  &status,
		Search:  &search,
	}

	filter := f.ToProjectFilter()

	assert.Equal(t, &ownerID, filter.OwnerID)
	require.NotNil(t, filter.Status)
	assert.Equal(t, domain.ProjectStatus("active"), *filter.Status)
	assert.Equal(t, &search, filter.Search)
}

func TestProjectFilterInput_ToProjectFilter_NilStatus(t *testing.T) {
	f := &ProjectFilterInput{}
	filter := f.ToProjectFilter()
	assert.Nil(t, filter.Status)
}
