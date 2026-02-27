// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	analyticsEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	analyticsRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/repositories"
	dashboardRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
)

const moodCacheKey = "metodych:mood"
const moodCacheTTL = 5 * time.Minute

// MoodUseCase handles mood computation for the Metodych character
type MoodUseCase struct {
	dashboardRepo dashboardRepos.DashboardRepository
	analyticsRepo analyticsRepos.AnalyticsRepository
	cache         *cache.RedisCache
	personality   *services.PersonalityService
}

// NewMoodUseCase creates a new MoodUseCase
func NewMoodUseCase(
	dashboardRepo dashboardRepos.DashboardRepository,
	analyticsRepo analyticsRepos.AnalyticsRepository,
	cache *cache.RedisCache,
	personality *services.PersonalityService,
) *MoodUseCase {
	return &MoodUseCase{
		dashboardRepo: dashboardRepo,
		analyticsRepo: analyticsRepo,
		cache:         cache,
		personality:   personality,
	}
}

// GetCurrentMood returns the current mood from cache or computes it
func (uc *MoodUseCase) GetCurrentMood(ctx context.Context) (*dto.MoodResponse, error) {
	// Try cache first
	if uc.cache != nil {
		var cached dto.MoodResponse
		found, err := uc.cache.Get(ctx, moodCacheKey, &cached)
		if err == nil && found {
			return &cached, nil
		}
	}

	// Compute mood
	mood, err := uc.ComputeMood(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compute mood: %w", err)
	}

	// Build response
	message := uc.personality.GetMoodComment(*mood)
	greeting := uc.personality.GetGreeting(mood.TimeOfDay)
	response := dto.ToMoodResponse(mood, message, greeting)

	// Cache the response
	if uc.cache != nil {
		cacheData, _ := json.Marshal(response)
		var cacheResponse dto.MoodResponse
		json.Unmarshal(cacheData, &cacheResponse)
		uc.cache.Set(ctx, moodCacheKey, response, moodCacheTTL)
	}

	return response, nil
}

// ComputeMood calculates the current mood based on system metrics
func (uc *MoodUseCase) ComputeMood(ctx context.Context) (*entities.MoodContext, error) {
	now := time.Now()
	timeOfDay := entities.GetTimeOfDay(now.Hour())
	dayOfWeek := now.Weekday().String()

	mood := &entities.MoodContext{
		State:      entities.MoodContent, // default
		Intensity:  0.5,
		TimeOfDay:  timeOfDay,
		DayOfWeek:  dayOfWeek,
		ComputedAt: now,
	}

	// Get overdue documents count (documents created but not processed in 30 days)
	docCount, err := uc.dashboardRepo.GetDocumentsCount(ctx, 30)
	if err == nil && docCount != nil {
		// Use the difference as a rough approximation of overdue
		overdue := max(int(docCount.Total-docCount.PreviousTotal), 0)
		mood.OverdueDocuments = overdue
	}

	// Get at-risk students
	atRiskStudents, totalAtRisk, err := uc.analyticsRepo.GetAtRiskStudents(ctx, 100, 0)
	if err == nil {
		mood.AtRiskStudents = int(totalAtRisk)

		// Count critical risk students
		criticalCount := 0
		for _, s := range atRiskStudents {
			if s.RiskLevel == analyticsEntities.RiskLevelCritical {
				criticalCount++
			}
		}

		// Determine attendance trend from analytics
		trends, trendErr := uc.analyticsRepo.GetMonthlyAttendanceTrend(ctx, 3)
		if trendErr == nil && len(trends) >= 2 {
			latest := trends[len(trends)-1].AttendanceRate
			previous := trends[len(trends)-2].AttendanceRate
			if latest > previous+2 {
				mood.AttendanceTrend = "improving"
			} else if latest < previous-2 {
				mood.AttendanceTrend = "declining"
			} else {
				mood.AttendanceTrend = "stable"
			}
		}

		// Apply mood rules
		switch {
		case mood.OverdueDocuments > 10 || criticalCount > 5:
			mood.State = entities.MoodPanicking
			mood.Intensity = 1.0
			mood.Reason = fmt.Sprintf("Критическая ситуация: %d просроченных документов, %d студентов в критической зоне", mood.OverdueDocuments, criticalCount)
		case mood.OverdueDocuments > 5 || criticalCount > 3:
			mood.State = entities.MoodStressed
			mood.Intensity = 0.8
			mood.Reason = fmt.Sprintf("Много дел: %d просроченных документов, %d студентов в зоне риска", mood.OverdueDocuments, criticalCount)
		case mood.AttendanceTrend == "improving" && mood.OverdueDocuments == 0:
			mood.State = entities.MoodInspired
			mood.Intensity = 0.9
			mood.Reason = "Посещаемость растёт, просрочек нет — всё отлично!"
		case mood.OverdueDocuments == 0 && int(totalAtRisk) == 0:
			mood.State = entities.MoodHappy
			mood.Intensity = 0.8
			mood.Reason = "Всё под контролем! Просрочек нет, студенты в порядке."
		}
	}

	// Time-based mood overrides (lower priority)
	isWeekend := now.Weekday() == time.Saturday || now.Weekday() == time.Sunday
	isEvening := now.Hour() >= 18
	isMondayMorning := now.Weekday() == time.Monday && now.Hour() < 12

	if mood.State == entities.MoodContent {
		switch {
		case isWeekend || isEvening:
			mood.State = entities.MoodRelaxed
			mood.Intensity = 0.4
			mood.Reason = "Вечер/выходной — можно немного расслабиться"
		case isMondayMorning:
			mood.State = entities.MoodWorried
			mood.Intensity = 0.6
			mood.Reason = "Понедельник утро — новая неделя, новые вызовы"
		}
	}

	return mood, nil
}
