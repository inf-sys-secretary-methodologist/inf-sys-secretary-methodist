// Package scheduler contains background jobs for the analytics module.
package scheduler

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// RiskRecalcScheduler recalculates student risk scores daily and saves history.
type RiskRecalcScheduler struct {
	scheduler     gocron.Scheduler
	analyticsRepo repositories.AnalyticsRepository
	logger        *logging.Logger
}

// NewRiskRecalcScheduler creates a new risk recalculation scheduler.
func NewRiskRecalcScheduler(
	analyticsRepo repositories.AnalyticsRepository,
	logger *logging.Logger,
) (*RiskRecalcScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	rs := &RiskRecalcScheduler{
		scheduler:     s,
		analyticsRepo: analyticsRepo,
		logger:        logger,
	}

	// Run daily at 3:00 AM
	_, err = s.NewJob(
		gocron.CronJob("0 3 * * *", false),
		gocron.NewTask(rs.recalculate),
	)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// Start begins the scheduler.
func (rs *RiskRecalcScheduler) Start() {
	rs.scheduler.Start()
	rs.logger.Info("Risk recalculation scheduler started (daily at 3:00 AM)", nil)
}

// Stop gracefully stops the scheduler.
func (rs *RiskRecalcScheduler) Stop() error {
	return rs.scheduler.Shutdown()
}

func (rs *RiskRecalcScheduler) recalculate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rs.logger.Info("Starting daily risk score recalculation", nil)

	// Fetch all students with risk scores (paginated)
	offset := 0
	batchSize := 100
	totalSaved := 0

	for {
		students, _, err := rs.analyticsRepo.GetAtRiskStudents(ctx, batchSize, offset)
		if err != nil {
			rs.logger.Error("Risk recalc: failed to fetch students", map[string]any{
				"error":  err.Error(),
				"offset": offset,
			})
			break
		}

		if len(students) == 0 {
			break
		}

		for _, s := range students {
			entry := &entities.RiskHistoryEntry{
				StudentID:      s.StudentID,
				RiskScore:      s.RiskScore,
				RiskLevel:      s.RiskLevel,
				AttendanceRate: s.AttendanceRate,
				GradeAverage:   s.GradeAverage,
				RiskFactors:    s.RiskFactors,
				CalculatedAt:   time.Now(),
			}

			if err := rs.analyticsRepo.SaveRiskHistory(ctx, entry); err != nil {
				rs.logger.Error("Risk recalc: failed to save history", map[string]any{
					"error":      err.Error(),
					"student_id": s.StudentID,
				})
				continue
			}
			totalSaved++
		}

		offset += batchSize
		if len(students) < batchSize {
			break
		}
	}

	// Also process non-at-risk students by fetching from risk score view
	for _, level := range []entities.RiskLevel{entities.RiskLevelLow, entities.RiskLevelMedium} {
		offset = 0
		for {
			students, _, err := rs.analyticsRepo.GetStudentsByRiskLevel(ctx, level, batchSize, offset)
			if err != nil {
				break
			}
			if len(students) == 0 {
				break
			}
			for _, s := range students {
				entry := &entities.RiskHistoryEntry{
					StudentID:      s.StudentID,
					RiskScore:      s.RiskScore,
					RiskLevel:      s.RiskLevel,
					AttendanceRate: s.AttendanceRate,
					GradeAverage:   s.GradeAverage,
					RiskFactors:    s.RiskFactors,
					CalculatedAt:   time.Now(),
				}
				if err := rs.analyticsRepo.SaveRiskHistory(ctx, entry); err != nil {
					continue
				}
				totalSaved++
			}
			offset += batchSize
			if len(students) < batchSize {
				break
			}
		}
	}

	rs.logger.Info("Risk recalculation completed", map[string]any{
		"total_saved": totalSaved,
	})
}
