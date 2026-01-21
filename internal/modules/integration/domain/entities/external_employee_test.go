package entities

import "testing"

func TestNewExternalEmployee(t *testing.T) {
	externalID := "ext-123"
	code := "EMP001"

	emp := NewExternalEmployee(externalID, code)

	if emp.ExternalID != externalID {
		t.Errorf("expected external ID %q, got %q", externalID, emp.ExternalID)
	}
	if emp.Code != code {
		t.Errorf("expected code %q, got %q", code, emp.Code)
	}
	if !emp.IsActive {
		t.Error("expected employee to be active")
	}
	if emp.LocalUserID != nil {
		t.Error("expected local user ID to be nil")
	}
	if emp.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestExternalEmployee_GetFullName(t *testing.T) {
	tests := []struct {
		name       string
		firstName  string
		lastName   string
		middleName string
		want       string
	}{
		{
			name:       "full name with middle name",
			firstName:  "Иван",
			lastName:   "Иванов",
			middleName: "Иванович",
			want:       "Иванов Иван Иванович",
		},
		{
			name:       "full name without middle name",
			firstName:  "John",
			lastName:   "Doe",
			middleName: "",
			want:       "Doe John",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emp := NewExternalEmployee("ext-1", "E001")
			emp.FirstName = tt.firstName
			emp.LastName = tt.lastName
			emp.MiddleName = tt.middleName

			got := emp.GetFullName()
			if got != tt.want {
				t.Errorf("GetFullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExternalEmployee_IsLinked(t *testing.T) {
	emp := NewExternalEmployee("ext-1", "E001")

	if emp.IsLinked() {
		t.Error("new employee should not be linked")
	}

	userID := int64(42)
	emp.LocalUserID = &userID

	if !emp.IsLinked() {
		t.Error("employee with user ID should be linked")
	}
}

func TestExternalEmployee_LinkToLocalUser(t *testing.T) {
	emp := NewExternalEmployee("ext-1", "E001")
	userID := int64(123)

	emp.LinkToLocalUser(userID)

	if emp.LocalUserID == nil || *emp.LocalUserID != userID {
		t.Errorf("expected local user ID %d, got %v", userID, emp.LocalUserID)
	}
}

func TestExternalEmployee_Unlink(t *testing.T) {
	emp := NewExternalEmployee("ext-1", "E001")
	userID := int64(123)
	emp.LocalUserID = &userID

	emp.Unlink()

	if emp.LocalUserID != nil {
		t.Error("expected local user ID to be nil after unlink")
	}
}

func TestExternalEmployee_MarkSynced(t *testing.T) {
	emp := NewExternalEmployee("ext-1", "E001")
	originalSyncTime := emp.LastSyncAt

	emp.MarkSynced()

	if !emp.LastSyncAt.After(originalSyncTime) && !emp.LastSyncAt.Equal(originalSyncTime) {
		t.Error("expected last sync time to be updated")
	}
}

func TestODataEmployee_ToExternalEmployee(t *testing.T) {
	odata := &ODataEmployee{
		RefKey:       "ref-123",
		Code:         "EMP001",
		FirstName:    "Иван",
		LastName:     "Иванов",
		MiddleName:   "Иванович",
		Email:        "ivan@example.com",
		Phone:        "+7999123456",
		Position:     "Developer",
		Department:   "IT",
		DeletionMark: false,
	}

	emp := odata.ToExternalEmployee()

	if emp.ExternalID != odata.RefKey {
		t.Errorf("expected external ID %q, got %q", odata.RefKey, emp.ExternalID)
	}
	if emp.Code != odata.Code {
		t.Errorf("expected code %q, got %q", odata.Code, emp.Code)
	}
	if emp.FirstName != odata.FirstName {
		t.Errorf("expected first name %q, got %q", odata.FirstName, emp.FirstName)
	}
	if emp.LastName != odata.LastName {
		t.Errorf("expected last name %q, got %q", odata.LastName, emp.LastName)
	}
	if emp.Email != odata.Email {
		t.Errorf("expected email %q, got %q", odata.Email, emp.Email)
	}
	if !emp.IsActive {
		t.Error("expected employee to be active when DeletionMark is false")
	}
}

func TestODataEmployee_ToExternalEmployee_DeletionMark(t *testing.T) {
	odata := &ODataEmployee{
		RefKey:       "ref-456",
		Code:         "EMP002",
		DeletionMark: true,
	}

	emp := odata.ToExternalEmployee()

	if emp.IsActive {
		t.Error("expected employee to be inactive when DeletionMark is true")
	}
}
