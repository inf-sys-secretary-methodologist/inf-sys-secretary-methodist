package entities

import "testing"

func TestNewExternalStudent(t *testing.T) {
	externalID := "student-123"
	code := "STU001"

	student := NewExternalStudent(externalID, code)

	if student.ExternalID != externalID {
		t.Errorf("expected external ID %q, got %q", externalID, student.ExternalID)
	}
	if student.Code != code {
		t.Errorf("expected code %q, got %q", code, student.Code)
	}
	if student.Status != StudentStatusEnrolled {
		t.Errorf("expected status %q, got %q", StudentStatusEnrolled, student.Status)
	}
	if !student.IsActive {
		t.Error("expected student to be active")
	}
	if student.LocalUserID != nil {
		t.Error("expected local user ID to be nil")
	}
	if student.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestExternalStudent_GetFullName(t *testing.T) {
	tests := []struct {
		name       string
		firstName  string
		lastName   string
		middleName string
		want       string
	}{
		{
			name:       "full name with middle name",
			firstName:  "Мария",
			lastName:   "Петрова",
			middleName: "Ивановна",
			want:       "Петрова Мария Ивановна",
		},
		{
			name:       "full name without middle name",
			firstName:  "Jane",
			lastName:   "Smith",
			middleName: "",
			want:       "Smith Jane",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			student := NewExternalStudent("ext-1", "S001")
			student.FirstName = tt.firstName
			student.LastName = tt.lastName
			student.MiddleName = tt.middleName

			got := student.GetFullName()
			if got != tt.want {
				t.Errorf("GetFullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExternalStudent_IsLinked(t *testing.T) {
	student := NewExternalStudent("ext-1", "S001")

	if student.IsLinked() {
		t.Error("new student should not be linked")
	}

	userID := int64(42)
	student.LocalUserID = &userID

	if !student.IsLinked() {
		t.Error("student with user ID should be linked")
	}
}

func TestExternalStudent_LinkToLocalUser(t *testing.T) {
	student := NewExternalStudent("ext-1", "S001")
	userID := int64(123)

	student.LinkToLocalUser(userID)

	if student.LocalUserID == nil || *student.LocalUserID != userID {
		t.Errorf("expected local user ID %d, got %v", userID, student.LocalUserID)
	}
}

func TestExternalStudent_Unlink(t *testing.T) {
	student := NewExternalStudent("ext-1", "S001")
	userID := int64(123)
	student.LocalUserID = &userID

	student.Unlink()

	if student.LocalUserID != nil {
		t.Error("expected local user ID to be nil after unlink")
	}
}

func TestExternalStudent_MarkSynced(t *testing.T) {
	student := NewExternalStudent("ext-1", "S001")
	originalSyncTime := student.LastSyncAt

	student.MarkSynced()

	if !student.LastSyncAt.After(originalSyncTime) && !student.LastSyncAt.Equal(originalSyncTime) {
		t.Error("expected last sync time to be updated")
	}
}

func TestStudentStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"enrolled", StudentStatusEnrolled, "enrolled"},
		{"graduated", StudentStatusGraduated, "graduated"},
		{"expelled", StudentStatusExpelled, "expelled"},
		{"academic leave", StudentStatusAcademic, "academic_leave"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}

func TestODataStudent_ToExternalStudent(t *testing.T) {
	odata := &ODataStudent{
		RefKey:       "ref-456",
		Code:         "STU002",
		FirstName:    "Мария",
		LastName:     "Петрова",
		MiddleName:   "Ивановна",
		Email:        "maria@example.com",
		Phone:        "+79991234567",
		GroupName:    "ИТ-21",
		Faculty:      "Информатика",
		Specialty:    "Программная инженерия",
		Course:       2,
		StudyForm:    "очная",
		DeletionMark: false,
	}

	student := odata.ToExternalStudent()

	if student.ExternalID != odata.RefKey {
		t.Errorf("expected external ID %q, got %q", odata.RefKey, student.ExternalID)
	}
	if student.Code != odata.Code {
		t.Errorf("expected code %q, got %q", odata.Code, student.Code)
	}
	if student.FirstName != odata.FirstName {
		t.Errorf("expected first name %q, got %q", odata.FirstName, student.FirstName)
	}
	if student.LastName != odata.LastName {
		t.Errorf("expected last name %q, got %q", odata.LastName, student.LastName)
	}
	if student.MiddleName != odata.MiddleName {
		t.Errorf("expected middle name %q, got %q", odata.MiddleName, student.MiddleName)
	}
	if student.Email != odata.Email {
		t.Errorf("expected email %q, got %q", odata.Email, student.Email)
	}
	if student.GroupName != odata.GroupName {
		t.Errorf("expected group name %q, got %q", odata.GroupName, student.GroupName)
	}
	if student.Faculty != odata.Faculty {
		t.Errorf("expected faculty %q, got %q", odata.Faculty, student.Faculty)
	}
	if student.Course != odata.Course {
		t.Errorf("expected course %d, got %d", odata.Course, student.Course)
	}
	if !student.IsActive {
		t.Error("expected student to be active when DeletionMark is false")
	}
}

func TestODataStudent_ToExternalStudent_DeletionMark(t *testing.T) {
	odata := &ODataStudent{
		RefKey:       "ref-789",
		Code:         "STU003",
		DeletionMark: true,
	}

	student := odata.ToExternalStudent()

	if student.IsActive {
		t.Error("expected student to be inactive when DeletionMark is true")
	}
}

func TestODataStudent_ToExternalStudent_WithStatus(t *testing.T) {
	odata := &ODataStudent{
		RefKey: "ref-999",
		Code:   "STU004",
		Status: StudentStatusGraduated,
	}

	student := odata.ToExternalStudent()

	if student.Status != StudentStatusGraduated {
		t.Errorf("expected status %q, got %q", StudentStatusGraduated, student.Status)
	}
}

func TestODataStudent_ToExternalStudent_EmptyStatus(t *testing.T) {
	odata := &ODataStudent{
		RefKey: "ref-111",
		Code:   "STU005",
		Status: "",
	}

	student := odata.ToExternalStudent()

	if student.Status != StudentStatusEnrolled {
		t.Errorf("expected default status %q, got %q", StudentStatusEnrolled, student.Status)
	}
}
