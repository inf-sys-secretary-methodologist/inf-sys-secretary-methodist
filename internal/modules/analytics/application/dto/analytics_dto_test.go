package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToStudentRiskResponse(t *testing.T) {
	group := "CS-101"
	attRate := 85.5
	gradeAvg := 3.8
	risk := &entities.StudentRiskScore{
		StudentID:      1,
		StudentName:    "John Doe",
		GroupName:      &group,
		AttendanceRate: &attRate,
		GradeAverage:   &gradeAvg,
		RiskLevel:      entities.RiskLevelMedium,
		RiskScore:      45.0,
		RiskFactors:    nil,
	}

	resp := ToStudentRiskResponse(risk)

	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.StudentID)
	assert.Equal(t, "John Doe", resp.StudentName)
	assert.Equal(t, &group, resp.GroupName)
	assert.Equal(t, &attRate, resp.AttendanceRate)
	assert.Equal(t, &gradeAvg, resp.GradeAverage)
	assert.Equal(t, "medium", resp.RiskLevel)
	assert.Equal(t, 45.0, resp.RiskScore)
	assert.Nil(t, resp.RiskFactors)
}

func TestToGroupSummaryResponse(t *testing.T) {
	summary := &entities.GroupAnalyticsSummary{
		GroupName:         "CS-102",
		TotalStudents:     30,
		AvgAttendanceRate: 90.0,
		AvgGrade:          4.0,
		CriticalRiskCount: 1,
		HighRiskCount:     2,
		MediumRiskCount:   5,
		LowRiskCount:      22,
		AtRiskPercentage:  10.0,
	}

	resp := ToGroupSummaryResponse(summary)

	require.NotNil(t, resp)
	assert.Equal(t, "CS-102", resp.GroupName)
	assert.Equal(t, 30, resp.TotalStudents)
	assert.Equal(t, 90.0, resp.AvgAttendanceRate)
	assert.Equal(t, 4.0, resp.AvgGrade)
	assert.Equal(t, 10.0, resp.AtRiskPercentage)
	assert.Equal(t, 1, resp.RiskDistribution.Critical)
	assert.Equal(t, 2, resp.RiskDistribution.High)
	assert.Equal(t, 5, resp.RiskDistribution.Medium)
	assert.Equal(t, 22, resp.RiskDistribution.Low)
}

func TestToMonthlyTrendResponse(t *testing.T) {
	month, _ := time.Parse("2006-01", "2024-03")
	trend := &entities.MonthlyAttendanceTrend{
		Month:          month,
		UniqueStudents: 25,
		TotalRecords:   100,
		PresentCount:   85,
		AbsentCount:    15,
		AttendanceRate: 85.0,
	}

	resp := ToMonthlyTrendResponse(trend)

	require.NotNil(t, resp)
	assert.Equal(t, "2024-03", resp.Month)
	assert.Equal(t, 25, resp.UniqueStudents)
	assert.Equal(t, 100, resp.TotalRecords)
	assert.Equal(t, 85, resp.PresentCount)
	assert.Equal(t, 15, resp.AbsentCount)
	assert.Equal(t, 85.0, resp.AttendanceRate)
}

func TestToAttendanceRecordResponse(t *testing.T) {
	now := time.Now()
	markedBy := int64(5)
	notes := "late arrival"
	lessonDate, _ := time.Parse("2006-01-02", "2024-03-15")
	record := &entities.AttendanceRecord{
		ID:         1,
		StudentID:  10,
		LessonID:   20,
		LessonDate: lessonDate,
		Status:     entities.AttendanceStatusLate,
		MarkedBy:   &markedBy,
		Notes:      &notes,
		CreatedAt:  now,
	}

	resp := ToAttendanceRecordResponse(record)

	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, int64(10), resp.StudentID)
	assert.Equal(t, int64(20), resp.LessonID)
	assert.Equal(t, "2024-03-15", resp.LessonDate)
	assert.Equal(t, "late", resp.Status)
	assert.Equal(t, &markedBy, resp.MarkedBy)
	assert.Equal(t, &notes, resp.Notes)
}

func TestRiskHistoryEntryFromEntity(t *testing.T) {
	now := time.Now()
	attRate := 80.0
	gradeAvg := 3.5
	subRate := 90.0
	entry := entities.RiskHistoryEntry{
		RiskScore:      55.0,
		RiskLevel:      entities.RiskLevelHigh,
		AttendanceRate: &attRate,
		GradeAverage:   &gradeAvg,
		SubmissionRate: &subRate,
		CalculatedAt:   now,
	}

	resp := RiskHistoryEntryFromEntity(entry)

	assert.Equal(t, 55.0, resp.RiskScore)
	assert.Equal(t, "high", resp.RiskLevel)
	assert.Equal(t, &attRate, resp.AttendanceRate)
	assert.Equal(t, &gradeAvg, resp.GradeAverage)
	assert.Equal(t, &subRate, resp.SubmissionRate)
	assert.Equal(t, now, resp.CalculatedAt)
}
