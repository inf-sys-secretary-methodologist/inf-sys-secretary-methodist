package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// --- Fake AnalyticsRepository ---
//
// Implements the 11-method repositories.AnalyticsRepository interface.
// Only the methods recalculate actually invokes (GetAtRiskStudents /
// GetStudentsByRiskLevel / SaveRiskHistory) carry behavioral fakes;
// the rest return zero values so the type satisfies the interface.

type fakeAnalyticsRepo struct {
	mu sync.Mutex

	// GetAtRiskStudents: paginated batches + optional error
	atRiskBatches [][]entities.StudentRiskScore
	atRiskCall    int
	atRiskErr     error

	// GetStudentsByRiskLevel: per-level batches
	byLevelBatches map[entities.RiskLevel][][]entities.StudentRiskScore
	byLevelCalls   map[entities.RiskLevel]int
	byLevelErr     error

	// SaveRiskHistory: failure injection + spy
	saved        []*entities.RiskHistoryEntry
	saveErr      error
	saveFailIDs  map[int64]bool // student IDs that should fail
	saveCallback func()
}

func newFakeAnalyticsRepo() *fakeAnalyticsRepo {
	return &fakeAnalyticsRepo{
		byLevelBatches: map[entities.RiskLevel][][]entities.StudentRiskScore{},
		byLevelCalls:   map[entities.RiskLevel]int{},
		saveFailIDs:    map[int64]bool{},
	}
}

func (m *fakeAnalyticsRepo) GetAtRiskStudents(_ context.Context, _ *entities.TeacherScope, _, _ int) ([]entities.StudentRiskScore, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.atRiskErr != nil {
		return nil, 0, m.atRiskErr
	}
	if m.atRiskCall >= len(m.atRiskBatches) {
		return nil, 0, nil
	}
	batch := m.atRiskBatches[m.atRiskCall]
	m.atRiskCall++
	return batch, int64(len(batch)), nil
}

func (m *fakeAnalyticsRepo) GetStudentsByRiskLevel(_ context.Context, _ *entities.TeacherScope, level entities.RiskLevel, _, _ int) ([]entities.StudentRiskScore, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byLevelErr != nil {
		return nil, 0, m.byLevelErr
	}
	batches := m.byLevelBatches[level]
	idx := m.byLevelCalls[level]
	m.byLevelCalls[level] = idx + 1
	if idx >= len(batches) {
		return nil, 0, nil
	}
	return batches[idx], int64(len(batches[idx])), nil
}

func (m *fakeAnalyticsRepo) SaveRiskHistory(_ context.Context, entry *entities.RiskHistoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveCallback != nil {
		m.saveCallback()
	}
	if m.saveFailIDs[entry.StudentID] {
		return errors.New("forced save error for student")
	}
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, entry)
	return nil
}

// --- Unused interface methods (zero-value fakes so the interface is satisfied) ---

func (m *fakeAnalyticsRepo) GetStudentRisk(_ context.Context, _ int64) (*entities.StudentRiskScore, error) {
	return nil, nil
}
func (m *fakeAnalyticsRepo) GetGroupSummary(_ context.Context, _ string) (*entities.GroupAnalyticsSummary, error) {
	return nil, nil
}
func (m *fakeAnalyticsRepo) GetAllGroupsSummary(_ context.Context, _ *entities.TeacherScope) ([]entities.GroupAnalyticsSummary, error) {
	return nil, nil
}
func (m *fakeAnalyticsRepo) GetMonthlyAttendanceTrend(_ context.Context, _ int) ([]entities.MonthlyAttendanceTrend, error) {
	return nil, nil
}
func (m *fakeAnalyticsRepo) GetRiskWeightConfig(_ context.Context) (*entities.RiskWeightConfig, error) {
	return nil, nil
}
func (m *fakeAnalyticsRepo) UpdateRiskWeightConfig(_ context.Context, _ *entities.RiskWeightConfig) error {
	return nil
}
func (m *fakeAnalyticsRepo) GetStudentRiskHistory(_ context.Context, _ int64, _ int) ([]entities.RiskHistoryEntry, error) {
	return nil, nil
}

func mkRisk(id int64, score float64, level entities.RiskLevel) entities.StudentRiskScore {
	return entities.StudentRiskScore{StudentID: id, RiskScore: score, RiskLevel: level}
}

// --- Tests ---

func TestNewRiskRecalcScheduler_HappyPath(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)
	require.NotNil(t, rs)
	require.NotNil(t, rs.scheduler, "internal scheduler must be wired")
	assert.Nil(t, rs.alertFunc, "no alert func when variadic empty")
}

func TestNewRiskRecalcScheduler_WithAlertFunc(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	alertCalled := false
	alert := func(_ context.Context, _ entities.StudentRiskScore) { alertCalled = true }

	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"), alert)
	require.NoError(t, err)
	require.NotNil(t, rs.alertFunc, "alert func must be wired when provided")
	rs.alertFunc(context.Background(), entities.StudentRiskScore{})
	assert.True(t, alertCalled, "captured alert func must be the same provided")
}

func TestRiskRecalcScheduler_StartStop(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)

	rs.Start()
	// Give the underlying gocron a moment to fully boot.
	time.Sleep(20 * time.Millisecond)

	require.NoError(t, rs.Stop(), "graceful Shutdown must not error")
}

// TestRecalculate_EmptyDataset pins the early-break path: when
// GetAtRiskStudents returns an empty batch the outer loop breaks
// without invoking SaveRiskHistory; the by-level fallback then loops
// over Low + Medium and also gets empty results.
func TestRecalculate_EmptyDataset(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)

	rs.recalculate()

	assert.Empty(t, repo.saved, "no history must be saved on empty dataset")
}

// TestRecalculate_AtRiskBatchSaved pins the success path: each student
// in the batch yields a SaveRiskHistory call; batch < batchSize ends loop.
func TestRecalculate_AtRiskBatchSaved(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	repo.atRiskBatches = [][]entities.StudentRiskScore{
		{
			mkRisk(1, 85.0, entities.RiskLevelCritical),
			mkRisk(2, 75.0, entities.RiskLevelHigh),
		},
	}
	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)

	rs.recalculate()

	assert.Len(t, repo.saved, 2)
	assert.Equal(t, int64(1), repo.saved[0].StudentID)
	assert.Equal(t, 85.0, repo.saved[0].RiskScore)
}

// TestRecalculate_AlertFiresForHighRisk pins the alertFunc invariant:
// student с score >= 70 triggers the alert; score < 70 does not.
func TestRecalculate_AlertFiresForHighRisk(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	repo.atRiskBatches = [][]entities.StudentRiskScore{
		{
			mkRisk(1, 85.0, entities.RiskLevelCritical), // alert
			mkRisk(2, 50.0, entities.RiskLevelMedium),   // no alert
			mkRisk(3, 70.0, entities.RiskLevelHigh),     // alert (boundary)
		},
	}

	var alertedIDs []int64
	alert := func(_ context.Context, s entities.StudentRiskScore) {
		alertedIDs = append(alertedIDs, s.StudentID)
	}

	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"), alert)
	require.NoError(t, err)
	rs.recalculate()

	assert.ElementsMatch(t, []int64{1, 3}, alertedIDs,
		"alert must fire for risk score >= 70 only")
}

// TestRecalculate_SaveErrorContinuesIteration pins resilience: one bad
// save does NOT abort the loop; remaining students are still saved.
func TestRecalculate_SaveErrorContinuesIteration(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	repo.atRiskBatches = [][]entities.StudentRiskScore{
		{
			mkRisk(1, 80.0, entities.RiskLevelCritical),
			mkRisk(2, 75.0, entities.RiskLevelHigh), // this one fails
			mkRisk(3, 90.0, entities.RiskLevelCritical),
		},
	}
	repo.saveFailIDs = map[int64]bool{2: true}

	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)
	rs.recalculate()

	// SaveRiskHistory called 3 times, but only 1 + 3 saved (2 failed)
	savedIDs := []int64{}
	for _, e := range repo.saved {
		savedIDs = append(savedIDs, e.StudentID)
	}
	assert.ElementsMatch(t, []int64{1, 3}, savedIDs,
		"failure on student 2 must not abort iteration")
}

// TestRecalculate_FetchErrorBreaks pins the fetch-fail path: when
// GetAtRiskStudents returns an error, the at-risk loop breaks immediately
// (no SaveRiskHistory invocations). By-level fallback still runs.
func TestRecalculate_FetchErrorBreaks(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	repo.atRiskErr = errors.New("DB connection lost")

	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)
	rs.recalculate()

	assert.Empty(t, repo.saved, "fetch error must break loop before any save")
}

// TestRecalculate_ByLevelLowMediumProcessed pins the by-level branch:
// when at-risk batch is empty, Low + Medium loops still fire and save
// to history. This is the second half of recalculate (line 132+).
func TestRecalculate_ByLevelLowMediumProcessed(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	repo.byLevelBatches = map[entities.RiskLevel][][]entities.StudentRiskScore{
		entities.RiskLevelLow: {
			{mkRisk(10, 20.0, entities.RiskLevelLow)},
		},
		entities.RiskLevelMedium: {
			{mkRisk(20, 50.0, entities.RiskLevelMedium)},
		},
	}

	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)
	rs.recalculate()

	assert.Len(t, repo.saved, 2, "both Low + Medium students must be saved")
	savedIDs := []int64{}
	for _, e := range repo.saved {
		savedIDs = append(savedIDs, e.StudentID)
	}
	assert.ElementsMatch(t, []int64{10, 20}, savedIDs)
}

// TestRecalculate_PaginationContinuesUntilSmallerBatch pins the paging
// invariant: batchSize is 100; while a batch returns exactly batchSize
// the loop continues, breaks when smaller batch comes back.
func TestRecalculate_PaginationContinuesUntilSmallerBatch(t *testing.T) {
	repo := newFakeAnalyticsRepo()
	// First call returns a full batch of 100; second call returns 1 (< 100).
	fullBatch := make([]entities.StudentRiskScore, 100)
	for i := range fullBatch {
		fullBatch[i] = mkRisk(int64(i+1), 80.0, entities.RiskLevelCritical)
	}
	tail := []entities.StudentRiskScore{mkRisk(200, 85.0, entities.RiskLevelCritical)}
	repo.atRiskBatches = [][]entities.StudentRiskScore{fullBatch, tail}

	rs, err := NewRiskRecalcScheduler(repo, logging.NewLogger("debug"))
	require.NoError(t, err)
	rs.recalculate()

	assert.Len(t, repo.saved, 101, "100 + 1 paginated saves")
	assert.Equal(t, 2, repo.atRiskCall, "fetch must be called exactly twice")
}
