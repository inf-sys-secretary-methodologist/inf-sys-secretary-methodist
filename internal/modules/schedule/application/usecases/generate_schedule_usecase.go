package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/solver"
)

// Generation errors surfaced by Apply.
var (
	// ErrScheduleAlreadyExists is returned by Apply when the target semester
	// already has lessons — the caller must clear them before regenerating, so a
	// re-apply never silently duplicates the timetable.
	ErrScheduleAlreadyExists = errors.New("schedule already exists for this semester")
	// ErrGenerateNotWritable is returned when Apply is called on a use case built
	// without its write dependencies.
	ErrGenerateNotWritable = errors.New("generate use case is not configured for apply")
	// ErrSemesterNotFound is returned when the target semester cannot be resolved.
	ErrSemesterNotFound = errors.New("semester not found")
)

// generateRoomLimit bounds the classroom fetch for a generation run; a single
// institution never has more available rooms than this.
const generateRoomLimit = 1000

// defaultTeachingDays is the working week the generator fills when the caller
// does not specify days: Monday through Saturday.
var defaultTeachingDays = []domain.DayOfWeek{
	domain.Monday, domain.Tuesday, domain.Wednesday, domain.Thursday, domain.Friday, domain.Saturday,
}

// The generator consumes only these narrow read surfaces (DIP): the concrete
// PG repositories already satisfy them.
type (
	generateLoadLister interface {
		List(ctx context.Context, filter TeachingLoadFilter) ([]*entities.TeachingLoad, error)
	}
	generateSlotLister interface {
		List(ctx context.Context) ([]*entities.LessonSlot, error)
	}
	generateRoomLister interface {
		List(ctx context.Context, filter ClassroomFilter, limit, offset int) ([]*entities.Classroom, error)
	}
	// generateLessonWriter is the write surface used by Apply.
	generateLessonWriter interface {
		Count(ctx context.Context, filter LessonFilter) (int64, error)
		Create(ctx context.Context, lesson *entities.Lesson) error
	}
	// generateSemesterLister resolves semester dates for the applied lessons.
	generateSemesterLister interface {
		ListSemesters(ctx context.Context, activeOnly bool) ([]*entities.Semester, error)
	}
)

// GenerateScheduleUseCase turns the planned teaching load into a draft timetable
// by unfolding it into CSP solver variables and running the (pure) solver. It
// never persists on its own — Preview computes a draft; applying it is a
// separate, explicit step.
type GenerateScheduleUseCase struct {
	loads     generateLoadLister
	slots     generateSlotLister
	rooms     generateRoomLister
	lessons   generateLessonWriter
	semesters generateSemesterLister
	weights   solver.SoftWeights
	now       func() time.Time
}

// GenerateOption overrides an optional dependency; Apply requires the write
// dependencies supplied via WithApplyWriter and WithSemesters.
type GenerateOption func(*GenerateScheduleUseCase)

// WithApplyWriter supplies the lesson writer used by Apply.
func WithApplyWriter(lessons generateLessonWriter) GenerateOption {
	return func(uc *GenerateScheduleUseCase) { uc.lessons = lessons }
}

// WithSemesters supplies the semester reader used by Apply for lesson dates.
func WithSemesters(semesters generateSemesterLister) GenerateOption {
	return func(uc *GenerateScheduleUseCase) { uc.semesters = semesters }
}

// WithGenerateClock overrides the time source (tests).
func WithGenerateClock(fn func() time.Time) GenerateOption {
	return func(uc *GenerateScheduleUseCase) { uc.now = fn }
}

// NewGenerateScheduleUseCase wires the use case with its read dependencies and
// the default soft-preference weights. Apply additionally requires the write
// dependencies passed as options.
func NewGenerateScheduleUseCase(loads generateLoadLister, slots generateSlotLister, rooms generateRoomLister, opts ...GenerateOption) *GenerateScheduleUseCase {
	uc := &GenerateScheduleUseCase{
		loads:   loads,
		slots:   slots,
		rooms:   rooms,
		weights: solver.NewDefaultWeights(),
		now:     time.Now,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// ApplyResult summarizes a persisted generation run.
type ApplyResult struct {
	Created  int
	Unplaced int
}

// Apply generates the schedule for the semester and persists the placed lessons.
// It refuses to run when the semester already has lessons (ErrScheduleAlreadyExists)
// so re-applying never silently duplicates the timetable.
func (uc *GenerateScheduleUseCase) Apply(ctx context.Context, params GenerateParams) (*ApplyResult, error) {
	if uc.lessons == nil || uc.semesters == nil {
		return nil, ErrGenerateNotWritable
	}
	if params.SemesterID <= 0 {
		return nil, ErrInvalidInput
	}
	semesterID := params.SemesterID

	existing, err := uc.lessons.Count(ctx, LessonFilter{SemesterID: &semesterID})
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, ErrScheduleAlreadyExists
	}

	semester, err := uc.semesterByID(ctx, semesterID)
	if err != nil {
		return nil, err
	}

	plan, err := uc.plan(ctx, params)
	if err != nil {
		return nil, err
	}

	now := uc.now()
	created := 0
	for _, a := range plan.result.Assignments {
		slot := plan.slotByNum[a.Value.Slot]
		if slot == nil {
			continue
		}
		lesson := entities.NewLesson(
			semesterID,
			a.Variable.DisciplineID,
			a.Variable.LessonTypeID,
			a.Variable.TeacherID,
			a.Variable.GroupID,
			a.Value.RoomID,
			a.Value.Day,
			slot.TimeStart,
			slot.TimeEnd,
			a.Variable.WeekType,
			semester.StartDate,
			semester.EndDate,
			now,
		)
		if err := lesson.Validate(); err != nil {
			return nil, err
		}
		if err := uc.lessons.Create(ctx, lesson); err != nil {
			return nil, err
		}
		created++
	}

	return &ApplyResult{Created: created, Unplaced: len(plan.result.Unplaced)}, nil
}

// semesterByID resolves a semester's dates from the reference catalog.
func (uc *GenerateScheduleUseCase) semesterByID(ctx context.Context, id int64) (*entities.Semester, error) {
	semesters, err := uc.semesters.ListSemesters(ctx, false)
	if err != nil {
		return nil, err
	}
	for _, s := range semesters {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, ErrSemesterNotFound
}

// GenerateParams is the request for a draft schedule.
type GenerateParams struct {
	SemesterID int64
	Days       []domain.DayOfWeek // optional; defaults to Mon-Sat
}

// GeneratedLesson is one placed lesson in a draft, with names resolved for display.
type GeneratedLesson struct {
	LoadID         int64
	GroupID        int64
	GroupName      string
	TeacherID      int64
	TeacherName    string
	DisciplineID   int64
	DisciplineName string
	LessonTypeID   int64
	LessonTypeName string
	WeekType       string
	DayOfWeek      int
	SlotNumber     int
	TimeStart      string
	TimeEnd        string
	ClassroomID    int64
	ClassroomName  string
}

// UnplacedLesson is a load line the solver could not place (best-effort).
type UnplacedLesson struct {
	LoadID         int64
	GroupName      string
	DisciplineName string
	LessonTypeName string
	WeekType       string
}

// SchedulePreview is the draft returned by Preview: placed lessons plus anything
// left unplaced, with summary counts.
type SchedulePreview struct {
	Lessons        []GeneratedLesson
	Unplaced       []UnplacedLesson
	TotalRequested int
	PlacedCount    int
	UnplacedCount  int
}

// Preview assembles the solver input for the semester, runs the solver, and
// returns the resulting draft without persisting anything.
func (uc *GenerateScheduleUseCase) Preview(ctx context.Context, params GenerateParams) (*SchedulePreview, error) {
	plan, err := uc.plan(ctx, params)
	if err != nil {
		return nil, err
	}
	return plan.toPreview(), nil
}

// generationPlan holds a solved run plus the lookups needed to resolve names,
// slot times, and rooms. It is shared by Preview and (later) Apply.
type generationPlan struct {
	result    solver.Result
	loadByID  map[int64]*entities.TeachingLoad
	slotByNum map[int]*entities.LessonSlot
	roomByID  map[int64]*entities.Classroom
}

// plan fetches catalog data, unfolds the load into variables, and solves.
func (uc *GenerateScheduleUseCase) plan(ctx context.Context, params GenerateParams) (*generationPlan, error) {
	if params.SemesterID <= 0 {
		return nil, ErrInvalidInput
	}
	semesterID := params.SemesterID

	loads, err := uc.loads.List(ctx, TeachingLoadFilter{SemesterID: &semesterID})
	if err != nil {
		return nil, err
	}

	slots, err := uc.slots.List(ctx)
	if err != nil {
		return nil, err
	}

	available := true
	rooms, err := uc.rooms.List(ctx, ClassroomFilter{IsAvailable: &available}, generateRoomLimit, 0)
	if err != nil {
		return nil, err
	}

	days := params.Days
	if len(days) == 0 {
		days = defaultTeachingDays
	}

	slotByNum := make(map[int]*entities.LessonSlot, len(slots))
	slotNumbers := make([]int, 0, len(slots))
	for _, s := range slots {
		slotByNum[s.Number] = s
		slotNumbers = append(slotNumbers, s.Number)
	}

	roomByID := make(map[int64]*entities.Classroom, len(rooms))
	solverRooms := make([]solver.Room, 0, len(rooms))
	for _, r := range rooms {
		roomByID[r.ID] = r
		roomType := ""
		if r.Type != nil {
			roomType = *r.Type
		}
		solverRooms = append(solverRooms, solver.Room{
			ID:        r.ID,
			Capacity:  r.Capacity,
			Type:      roomType,
			Available: r.IsAvailable,
		})
	}

	loadByID := make(map[int64]*entities.TeachingLoad, len(loads))
	var variables []solver.Variable
	// Loads with an un-hydrated group have an unknown size; scheduling them would
	// silently disable the room-capacity constraint (H4), so they are forced into
	// the unplaced list instead of being fed to the solver.
	var forcedUnplaced []solver.Variable
	varID := 0
	for _, load := range loads {
		loadByID[load.ID] = load
		var allowed []string
		if load.LessonType != nil {
			allowed = domain.AllowedRoomTypesForLesson(load.LessonType.ShortName)
		}
		for range load.PairsPerWeek {
			varID++
			variable := solver.Variable{
				ID:               varID,
				LoadID:           load.ID,
				GroupID:          load.GroupID,
				TeacherID:        load.TeacherID,
				DisciplineID:     load.DisciplineID,
				LessonTypeID:     load.LessonTypeID,
				AllowedRoomTypes: allowed,
				WeekType:         load.WeekType,
			}
			if load.Group == nil {
				forcedUnplaced = append(forcedUnplaced, variable)
				continue
			}
			variable.GroupSize = load.Group.Capacity
			variables = append(variables, variable)
		}
	}

	result := solver.Solve(solver.Input{
		Variables: variables,
		Days:      days,
		Slots:     slotNumbers,
		Rooms:     solverRooms,
		Weights:   uc.weights,
	})
	result.Unplaced = append(result.Unplaced, forcedUnplaced...)

	return &generationPlan{
		result:    result,
		loadByID:  loadByID,
		slotByNum: slotByNum,
		roomByID:  roomByID,
	}, nil
}

// toPreview maps a solved plan into the display-ready preview.
func (p *generationPlan) toPreview() *SchedulePreview {
	preview := &SchedulePreview{}

	for _, a := range p.result.Assignments {
		load := p.loadByID[a.Variable.LoadID]
		lesson := GeneratedLesson{
			LoadID:       a.Variable.LoadID,
			GroupID:      a.Variable.GroupID,
			TeacherID:    a.Variable.TeacherID,
			DisciplineID: a.Variable.DisciplineID,
			LessonTypeID: a.Variable.LessonTypeID,
			WeekType:     string(a.Variable.WeekType),
			DayOfWeek:    int(a.Value.Day),
			SlotNumber:   a.Value.Slot,
			ClassroomID:  a.Value.RoomID,
		}
		if load != nil {
			lesson.GroupName = groupName(load)
			lesson.TeacherName = teacherName(load)
			lesson.DisciplineName = disciplineName(load)
			lesson.LessonTypeName = lessonTypeName(load)
		}
		if slot := p.slotByNum[a.Value.Slot]; slot != nil {
			lesson.TimeStart = slot.TimeStart
			lesson.TimeEnd = slot.TimeEnd
		}
		if room := p.roomByID[a.Value.RoomID]; room != nil {
			lesson.ClassroomName = classroomName(room)
		}
		preview.Lessons = append(preview.Lessons, lesson)
	}

	for _, v := range p.result.Unplaced {
		load := p.loadByID[v.LoadID]
		unplaced := UnplacedLesson{
			LoadID:   v.LoadID,
			WeekType: string(v.WeekType),
		}
		if load != nil {
			unplaced.GroupName = groupName(load)
			unplaced.DisciplineName = disciplineName(load)
			unplaced.LessonTypeName = lessonTypeName(load)
		}
		preview.Unplaced = append(preview.Unplaced, unplaced)
	}

	preview.PlacedCount = len(preview.Lessons)
	preview.UnplacedCount = len(preview.Unplaced)
	preview.TotalRequested = preview.PlacedCount + preview.UnplacedCount
	return preview
}

func groupName(l *entities.TeachingLoad) string {
	if l.Group != nil {
		return l.Group.Name
	}
	return ""
}

func teacherName(l *entities.TeachingLoad) string {
	if l.Teacher != nil {
		return l.Teacher.Name
	}
	return ""
}

func disciplineName(l *entities.TeachingLoad) string {
	if l.Discipline != nil {
		return l.Discipline.Name
	}
	return ""
}

func lessonTypeName(l *entities.TeachingLoad) string {
	if l.LessonType != nil {
		return l.LessonType.Name
	}
	return ""
}

// classroomName prefers an explicit room name, falling back to "Building-Number".
func classroomName(r *entities.Classroom) string {
	if r.Name != nil && *r.Name != "" {
		return *r.Name
	}
	return r.Building + "-" + r.Number
}
