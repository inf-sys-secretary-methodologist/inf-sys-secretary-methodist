package entities

import (
	"time"
)

// ExternalStudent represents a student record from 1C system
type ExternalStudent struct {
	ID               int64      `json:"id"`
	ExternalID       string     `json:"external_id"`           // 1C GUID (Ref_Key)
	Code             string     `json:"code"`                  // 1C Code (зачетка)
	FirstName        string     `json:"first_name"`            // Имя
	LastName         string     `json:"last_name"`             // Фамилия
	MiddleName       string     `json:"middle_name,omitempty"` // Отчество
	Email            string     `json:"email,omitempty"`       // Email
	Phone            string     `json:"phone,omitempty"`       // Телефон
	GroupName        string     `json:"group_name,omitempty"`  // Группа
	Faculty          string     `json:"faculty,omitempty"`     // Факультет
	Specialty        string     `json:"specialty,omitempty"`   // Специальность
	Course           int        `json:"course,omitempty"`      // Курс
	StudyForm        string     `json:"study_form,omitempty"`  // Форма обучения
	EnrollmentDate   *time.Time `json:"enrollment_date,omitempty"`
	ExpulsionDate    *time.Time `json:"expulsion_date,omitempty"`
	GraduationDate   *time.Time `json:"graduation_date,omitempty"`
	Status           string     `json:"status"` // enrolled, graduated, expelled
	IsActive         bool       `json:"is_active"`
	LocalUserID      *int64     `json:"local_user_id,omitempty"` // Linked local user
	LastSyncAt       time.Time  `json:"last_sync_at"`
	ExternalDataHash string     `json:"external_data_hash"`
	RawData          string     `json:"raw_data,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Student status constants
const (
	StudentStatusEnrolled  = "enrolled"
	StudentStatusGraduated = "graduated"
	StudentStatusExpelled  = "expelled"
	StudentStatusAcademic  = "academic_leave"
)

// NewExternalStudent creates a new external student record
func NewExternalStudent(externalID, code string) *ExternalStudent {
	now := time.Now()
	return &ExternalStudent{
		ExternalID: externalID,
		Code:       code,
		Status:     StudentStatusEnrolled,
		IsActive:   true,
		LastSyncAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// GetFullName returns the full name of the student
func (s *ExternalStudent) GetFullName() string {
	if s.MiddleName != "" {
		return s.LastName + " " + s.FirstName + " " + s.MiddleName
	}
	return s.LastName + " " + s.FirstName
}

// IsLinked returns true if the student is linked to a local user
func (s *ExternalStudent) IsLinked() bool {
	return s.LocalUserID != nil
}

// LinkToLocalUser links the external student to a local user
func (s *ExternalStudent) LinkToLocalUser(userID int64) {
	s.LocalUserID = &userID
	s.UpdatedAt = time.Now()
}

// Unlink removes the link to a local user
func (s *ExternalStudent) Unlink() {
	s.LocalUserID = nil
	s.UpdatedAt = time.Now()
}

// MarkSynced updates the last sync timestamp
func (s *ExternalStudent) MarkSynced() {
	s.LastSyncAt = time.Now()
	s.UpdatedAt = time.Now()
}

// ExternalStudentFilter represents filter options for external students
type ExternalStudentFilter struct {
	Search    string `json:"search,omitempty"`
	GroupName string `json:"group_name,omitempty"`
	Faculty   string `json:"faculty,omitempty"`
	Course    *int   `json:"course,omitempty"`
	Status    string `json:"status,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
	IsLinked  *bool  `json:"is_linked,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// ODataStudent represents the student structure from 1C OData response
type ODataStudent struct {
	RefKey         string `json:"Ref_Key"`
	DataVersion    string `json:"DataVersion"`
	DeletionMark   bool   `json:"DeletionMark"`
	Code           string `json:"Code"`
	Description    string `json:"Description"`
	FirstName      string `json:"Имя,omitempty"`
	LastName       string `json:"Фамилия,omitempty"`
	MiddleName     string `json:"Отчество,omitempty"`
	Email          string `json:"Email,omitempty"`
	Phone          string `json:"Телефон,omitempty"`
	GroupName      string `json:"Группа,omitempty"`
	Faculty        string `json:"Факультет,omitempty"`
	Specialty      string `json:"Специальность,omitempty"`
	Course         int    `json:"Курс,omitempty"`
	StudyForm      string `json:"ФормаОбучения,omitempty"`
	EnrollmentDate string `json:"ДатаЗачисления,omitempty"`
	Status         string `json:"Статус,omitempty"`
}

// ToExternalStudent converts OData response to ExternalStudent entity
func (o *ODataStudent) ToExternalStudent() *ExternalStudent {
	student := NewExternalStudent(o.RefKey, o.Code)
	student.FirstName = o.FirstName
	student.LastName = o.LastName
	student.MiddleName = o.MiddleName
	student.Email = o.Email
	student.Phone = o.Phone
	student.GroupName = o.GroupName
	student.Faculty = o.Faculty
	student.Specialty = o.Specialty
	student.Course = o.Course
	student.StudyForm = o.StudyForm
	student.IsActive = !o.DeletionMark

	if o.Status != "" {
		student.Status = o.Status
	}

	return student
}
