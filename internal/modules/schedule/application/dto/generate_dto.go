package dto

// GenerateRequest is the request body for previewing or applying a generated
// schedule. Days is optional; when empty the generator uses Monday-Saturday.
type GenerateRequest struct {
	SemesterID int64 `json:"semester_id"`
	Days       []int `json:"days,omitempty"`
}

// GeneratedLessonOutput is one placed lesson in a draft, ready for display.
type GeneratedLessonOutput struct {
	LoadID         int64  `json:"load_id"`
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	TeacherID      int64  `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	DisciplineID   int64  `json:"discipline_id"`
	DisciplineName string `json:"discipline_name"`
	LessonTypeID   int64  `json:"lesson_type_id"`
	LessonTypeName string `json:"lesson_type_name"`
	WeekType       string `json:"week_type"`
	DayOfWeek      int    `json:"day_of_week"`
	SlotNumber     int    `json:"slot_number"`
	TimeStart      string `json:"time_start"`
	TimeEnd        string `json:"time_end"`
	ClassroomID    int64  `json:"classroom_id"`
	ClassroomName  string `json:"classroom_name"`
}

// UnplacedLessonOutput is a load line the solver could not place.
type UnplacedLessonOutput struct {
	LoadID         int64  `json:"load_id"`
	GroupName      string `json:"group_name"`
	DisciplineName string `json:"discipline_name"`
	LessonTypeName string `json:"lesson_type_name"`
	WeekType       string `json:"week_type"`
}

// SchedulePreviewOutput is the draft returned by the preview endpoint.
type SchedulePreviewOutput struct {
	Lessons        []GeneratedLessonOutput `json:"lessons"`
	Unplaced       []UnplacedLessonOutput  `json:"unplaced"`
	TotalRequested int                     `json:"total_requested"`
	PlacedCount    int                     `json:"placed_count"`
	UnplacedCount  int                     `json:"unplaced_count"`
}

// ApplyResultOutput summarizes a persisted generation run.
type ApplyResultOutput struct {
	Created  int `json:"created"`
	Unplaced int `json:"unplaced"`
}
