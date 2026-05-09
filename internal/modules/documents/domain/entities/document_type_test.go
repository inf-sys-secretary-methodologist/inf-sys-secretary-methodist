package entities

import "testing"

// Pin the v0.126.0 contract: a document type marked methodist_only is
// hidden from teacher and student; non-methodist_only is visible to
// every role. Admin / methodist / secretary always see methodist-only
// templates because they need them for paperwork orchestration.
func TestDocumentType_CanAccessByRole(t *testing.T) {
	tests := []struct {
		name           string
		methodistOnly  bool
		role           string
		expectedAccess bool
	}{
		// Open templates: visible to every known role
		{"open / system_admin", false, "system_admin", true},
		{"open / methodist", false, "methodist", true},
		{"open / academic_secretary", false, "academic_secretary", true},
		{"open / teacher", false, "teacher", true},
		{"open / student", false, "student", true},
		// Methodist-only templates: visible only to staff who orchestrate paperwork
		{"methodist_only / system_admin", true, "system_admin", true},
		{"methodist_only / methodist", true, "methodist", true},
		{"methodist_only / academic_secretary", true, "academic_secretary", true},
		// Methodist-only templates: hidden from end-users (teacher fills, student receives)
		{"methodist_only / teacher", true, "teacher", false},
		{"methodist_only / student", true, "student", false},
		// Failure-closed: an unknown role is never allowed access
		{"open / unknown role", false, "unknown_role", true},
		{"methodist_only / unknown role", true, "unknown_role", false},
		// Empty role string is treated as unknown
		{"open / empty role", false, "", true},
		{"methodist_only / empty role", true, "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dt := &DocumentType{MethodistOnly: tc.methodistOnly}
			got := dt.CanAccessByRole(tc.role)
			if got != tc.expectedAccess {
				t.Errorf("CanAccessByRole(%q) with MethodistOnly=%v: got %v, want %v",
					tc.role, tc.methodistOnly, got, tc.expectedAccess)
			}
		})
	}
}
