package entities

import "testing"

func TestNewDepartment(t *testing.T) {
	tests := []struct {
		name        string
		deptName    string
		code        string
		description string
		parentID    *int64
	}{
		{
			name:        "create department without parent",
			deptName:    "IT Department",
			code:        "IT",
			description: "Information Technology",
			parentID:    nil,
		},
		{
			name:        "create department with parent",
			deptName:    "Frontend Team",
			code:        "FE",
			description: "Frontend development team",
			parentID:    ptrInt64(1),
		},
		{
			name:        "create department with empty description",
			deptName:    "HR",
			code:        "HR",
			description: "",
			parentID:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dept := NewDepartment(tt.deptName, tt.code, tt.description, tt.parentID)

			if dept.Name != tt.deptName {
				t.Errorf("expected name %q, got %q", tt.deptName, dept.Name)
			}
			if dept.Code != tt.code {
				t.Errorf("expected code %q, got %q", tt.code, dept.Code)
			}
			if dept.Description != tt.description {
				t.Errorf("expected description %q, got %q", tt.description, dept.Description)
			}
			if tt.parentID == nil && dept.ParentID != nil {
				t.Error("expected nil parent ID")
			}
			if tt.parentID != nil && (dept.ParentID == nil || *dept.ParentID != *tt.parentID) {
				t.Errorf("expected parent ID %d, got %v", *tt.parentID, dept.ParentID)
			}
			if !dept.IsActive {
				t.Error("expected department to be active")
			}
			if dept.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}
			if dept.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}
		})
	}
}

func ptrInt64(i int64) *int64 {
	return &i
}
