package entities

import (
	"time"
)

// ExternalEmployee represents an employee record from 1C system
type ExternalEmployee struct {
	ID               int64      `db:"id" json:"id"`
	ExternalID       string     `db:"external_id" json:"external_id"`               // 1C GUID (Ref_Key)
	Code             string     `db:"code" json:"code"`                             // 1C Code
	FirstName        string     `db:"first_name" json:"first_name"`                 // Имя
	LastName         string     `db:"last_name" json:"last_name"`                   // Фамилия
	MiddleName       string     `db:"middle_name" json:"middle_name,omitempty"`     // Отчество
	Email            string     `db:"email" json:"email,omitempty"`                 // Email
	Phone            string     `db:"phone" json:"phone,omitempty"`                 // Телефон
	Position         string     `db:"position" json:"position,omitempty"`           // Должность
	Department       string     `db:"department" json:"department,omitempty"`       // Подразделение
	EmploymentDate   *time.Time `db:"employment_date" json:"employment_date,omitempty"`
	DismissalDate    *time.Time `db:"dismissal_date" json:"dismissal_date,omitempty"`
	IsActive         bool       `db:"is_active" json:"is_active"`
	LocalUserID      *int64     `db:"local_user_id" json:"local_user_id,omitempty"` // Linked local user
	LastSyncAt       time.Time  `db:"last_sync_at" json:"last_sync_at"`
	ExternalDataHash string     `db:"external_data_hash" json:"external_data_hash"` // Hash for change detection
	RawData          string     `db:"raw_data" json:"raw_data,omitempty"`           // Original JSON from 1C
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// NewExternalEmployee creates a new external employee record
func NewExternalEmployee(externalID, code string) *ExternalEmployee {
	now := time.Now()
	return &ExternalEmployee{
		ExternalID: externalID,
		Code:       code,
		IsActive:   true,
		LastSyncAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// GetFullName returns the full name of the employee
func (e *ExternalEmployee) GetFullName() string {
	if e.MiddleName != "" {
		return e.LastName + " " + e.FirstName + " " + e.MiddleName
	}
	return e.LastName + " " + e.FirstName
}

// IsLinked returns true if the employee is linked to a local user
func (e *ExternalEmployee) IsLinked() bool {
	return e.LocalUserID != nil
}

// LinkToLocalUser links the external employee to a local user
func (e *ExternalEmployee) LinkToLocalUser(userID int64) {
	e.LocalUserID = &userID
	e.UpdatedAt = time.Now()
}

// Unlink removes the link to a local user
func (e *ExternalEmployee) Unlink() {
	e.LocalUserID = nil
	e.UpdatedAt = time.Now()
}

// MarkSynced updates the last sync timestamp
func (e *ExternalEmployee) MarkSynced() {
	e.LastSyncAt = time.Now()
	e.UpdatedAt = time.Now()
}

// ExternalEmployeeFilter represents filter options for external employees
type ExternalEmployeeFilter struct {
	Search       string `json:"search,omitempty"`
	Department   string `json:"department,omitempty"`
	Position     string `json:"position,omitempty"`
	IsActive     *bool  `json:"is_active,omitempty"`
	IsLinked     *bool  `json:"is_linked,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}

// ODataEmployee represents the employee structure from 1C OData response
type ODataEmployee struct {
	RefKey         string  `json:"Ref_Key"`
	DataVersion    string  `json:"DataVersion"`
	DeletionMark   bool    `json:"DeletionMark"`
	Code           string  `json:"Code"`
	Description    string  `json:"Description"`
	FirstName      string  `json:"ФизическоеЛицо_Имя,omitempty"`
	LastName       string  `json:"ФизическоеЛицо_Фамилия,omitempty"`
	MiddleName     string  `json:"ФизическоеЛицо_Отчество,omitempty"`
	Email          string  `json:"КонтактнаяИнформация_АдресЭП,omitempty"`
	Phone          string  `json:"КонтактнаяИнформация_Телефон,omitempty"`
	Position       string  `json:"Должность,omitempty"`
	Department     string  `json:"Подразделение,omitempty"`
	EmploymentDate string  `json:"ДатаПриема,omitempty"`
	DismissalDate  string  `json:"ДатаУвольнения,omitempty"`
}

// ToExternalEmployee converts OData response to ExternalEmployee entity
func (o *ODataEmployee) ToExternalEmployee() *ExternalEmployee {
	emp := NewExternalEmployee(o.RefKey, o.Code)
	emp.FirstName = o.FirstName
	emp.LastName = o.LastName
	emp.MiddleName = o.MiddleName
	emp.Email = o.Email
	emp.Phone = o.Phone
	emp.Position = o.Position
	emp.Department = o.Department
	emp.IsActive = !o.DeletionMark

	// Parse dates if present
	// Note: 1C typically uses ISO 8601 format
	// Date parsing should be handled in the infrastructure layer

	return emp
}
