package usecases

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// ========== MockLessonRepository ==========

type MockLessonRepository struct {
	mu      sync.RWMutex
	lessons map[int64]*entities.Lesson
	nextID  atomic.Int64
}

func NewMockLessonRepository() *MockLessonRepository {
	return &MockLessonRepository{
		lessons: make(map[int64]*entities.Lesson),
	}
}

func (m *MockLessonRepository) Create(_ context.Context, lesson *entities.Lesson) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lesson.ID = m.nextID.Add(1)
	m.lessons[lesson.ID] = lesson
	return nil
}

func (m *MockLessonRepository) Save(_ context.Context, lesson *entities.Lesson) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lessons[lesson.ID] = lesson
	return nil
}

func (m *MockLessonRepository) GetByID(_ context.Context, id int64) (*entities.Lesson, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if l, ok := m.lessons[id]; ok {
		return l, nil
	}
	return nil, nil
}

func (m *MockLessonRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.lessons, id)
	return nil
}

func (m *MockLessonRepository) List(_ context.Context, _ repositories.LessonFilter, limit, offset int) ([]*entities.Lesson, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entities.Lesson
	for _, l := range m.lessons {
		all = append(all, l)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MockLessonRepository) Count(_ context.Context, _ repositories.LessonFilter) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.lessons)), nil
}

func (m *MockLessonRepository) GetTimetable(_ context.Context, _ repositories.LessonFilter) ([]*entities.Lesson, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entities.Lesson
	for _, l := range m.lessons {
		all = append(all, l)
	}
	return all, nil
}

// ========== MockClassroomRepository ==========

type MockClassroomRepository struct {
	mu         sync.RWMutex
	classrooms map[int64]*entities.Classroom
}

func NewMockClassroomRepository() *MockClassroomRepository {
	return &MockClassroomRepository{
		classrooms: make(map[int64]*entities.Classroom),
	}
}

func (m *MockClassroomRepository) GetByID(_ context.Context, id int64) (*entities.Classroom, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if c, ok := m.classrooms[id]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *MockClassroomRepository) List(_ context.Context, _ repositories.ClassroomFilter, limit, offset int) ([]*entities.Classroom, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entities.Classroom
	for _, c := range m.classrooms {
		all = append(all, c)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MockClassroomRepository) Count(_ context.Context, _ repositories.ClassroomFilter) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.classrooms)), nil
}

// ========== MockReferenceRepository ==========

type MockReferenceRepository struct {
	mu       sync.RWMutex
	groups   []*entities.StudentGroup
	discs    []*entities.Discipline
	sems     []*entities.Semester
	ltypes   []*entities.LessonType
	activeSm *entities.Semester
}

func NewMockReferenceRepository() *MockReferenceRepository {
	return &MockReferenceRepository{}
}

func (m *MockReferenceRepository) ListStudentGroups(_ context.Context, _, _ int) ([]*entities.StudentGroup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.groups, nil
}

func (m *MockReferenceRepository) ListDisciplines(_ context.Context, _, _ int) ([]*entities.Discipline, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.discs, nil
}

func (m *MockReferenceRepository) ListSemesters(_ context.Context, _ bool) ([]*entities.Semester, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sems, nil
}

func (m *MockReferenceRepository) ListLessonTypes(_ context.Context) ([]*entities.LessonType, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ltypes, nil
}

func (m *MockReferenceRepository) GetActiveSemester(_ context.Context) (*entities.Semester, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeSm, nil
}

// ========== MockScheduleChangeRepository ==========

type MockScheduleChangeRepository struct {
	mu      sync.RWMutex
	changes map[int64][]*entities.ScheduleChange
	nextID  atomic.Int64
}

func NewMockScheduleChangeRepository() *MockScheduleChangeRepository {
	return &MockScheduleChangeRepository{
		changes: make(map[int64][]*entities.ScheduleChange),
	}
}

func (m *MockScheduleChangeRepository) Create(_ context.Context, change *entities.ScheduleChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	change.ID = m.nextID.Add(1)
	m.changes[change.LessonID] = append(m.changes[change.LessonID], change)
	return nil
}

func (m *MockScheduleChangeRepository) GetByLessonID(_ context.Context, lessonID int64) ([]*entities.ScheduleChange, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.changes[lessonID], nil
}

func (m *MockScheduleChangeRepository) GetByDateRange(_ context.Context, _, _ time.Time) ([]*entities.ScheduleChange, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entities.ScheduleChange
	for _, changes := range m.changes {
		all = append(all, changes...)
	}
	return all, nil
}

// ========== Helpers ==========

func setupLessonUseCase() (*LessonUseCase, *MockLessonRepository) {
	lessonRepo := NewMockLessonRepository()
	classroomRepo := NewMockClassroomRepository()
	referenceRepo := NewMockReferenceRepository()
	changeRepo := NewMockScheduleChangeRepository()
	uc := NewLessonUseCase(lessonRepo, classroomRepo, referenceRepo, changeRepo, nil)
	return uc, lessonRepo
}

func setupLessonUseCaseAll() (*LessonUseCase, *MockLessonRepository, *MockClassroomRepository, *MockReferenceRepository, *MockScheduleChangeRepository) {
	lessonRepo := NewMockLessonRepository()
	classroomRepo := NewMockClassroomRepository()
	referenceRepo := NewMockReferenceRepository()
	changeRepo := NewMockScheduleChangeRepository()
	uc := NewLessonUseCase(lessonRepo, classroomRepo, referenceRepo, changeRepo, nil)
	return uc, lessonRepo, classroomRepo, referenceRepo, changeRepo
}

func validCreateLessonInput() CreateLessonInputForUC {
	return CreateLessonInputForUC{
		SemesterID:   1,
		DisciplineID: 2,
		LessonTypeID: 3,
		TeacherID:    4,
		GroupID:      5,
		ClassroomID:  6,
		DayOfWeek:    domain.Monday,
		TimeStart:    "09:00",
		TimeEnd:      "10:30",
		WeekType:     domain.WeekTypeAll,
	}
}
