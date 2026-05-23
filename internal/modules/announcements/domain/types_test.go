package domain

import "testing"

func TestAnnouncementStatus_IsValid(t *testing.T) {
	tests := []struct {
		status AnnouncementStatus
		want   bool
	}{
		{AnnouncementStatusDraft, true},
		{AnnouncementStatusPublished, true},
		{AnnouncementStatusArchived, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("AnnouncementStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnnouncementPriority_IsValid(t *testing.T) {
	tests := []struct {
		priority AnnouncementPriority
		want     bool
	}{
		{AnnouncementPriorityLow, true},
		{AnnouncementPriorityNormal, true},
		{AnnouncementPriorityHigh, true},
		{AnnouncementPriorityUrgent, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := tt.priority.IsValid(); got != tt.want {
				t.Errorf("AnnouncementPriority.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTargetAudience_IsValid(t *testing.T) {
	tests := []struct {
		audience TargetAudience
		want     bool
	}{
		{TargetAudienceAll, true},
		{TargetAudienceStudents, true},
		{TargetAudienceTeachers, true},
		{TargetAudienceStaff, true},
		{TargetAudienceAdmins, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.audience), func(t *testing.T) {
			if got := tt.audience.IsValid(); got != tt.want {
				t.Errorf("TargetAudience.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCanAccessAudience_RoleMatrix pins v0.163.0 ADR-2 (#303 TIER 0):
// access matrix between caller's role and announcement target_audience.
// Pre-fix the handler trusted the client's ?audience= param, so a
// student could request audience=admins to read admin-broadcasts.
func TestCanAccessAudience_RoleMatrix(t *testing.T) {
	tests := []struct {
		role     string
		audience TargetAudience
		want     bool
	}{
		{"student", TargetAudienceAll, true},
		{"student", TargetAudienceStudents, true},
		{"student", TargetAudienceAdmins, false},
		{"student", TargetAudienceStaff, false},
		{"student", TargetAudienceTeachers, false},
		{"teacher", TargetAudienceAll, true},
		{"teacher", TargetAudienceTeachers, true},
		{"teacher", TargetAudienceAdmins, false},
		{"teacher", TargetAudienceStudents, false},
		{"methodist", TargetAudienceStaff, true},
		{"methodist", TargetAudienceAdmins, false},
		{"academic_secretary", TargetAudienceStaff, true},
		{"academic_secretary", TargetAudienceStudents, false},
		{"system_admin", TargetAudienceAll, true},
		{"system_admin", TargetAudienceAdmins, true},
		{"system_admin", TargetAudienceStudents, true},
		{"system_admin", TargetAudienceTeachers, true},
		{"system_admin", TargetAudienceStaff, true},
		{"system_admin", "garbage", false},
		{"", TargetAudienceAll, true},
		{"", TargetAudienceAdmins, false},
		{"unknown_role", TargetAudienceStudents, false},
	}

	for _, tc := range tests {
		name := tc.role + "/" + string(tc.audience)
		t.Run(name, func(t *testing.T) {
			got := CanAccessAudience(tc.role, tc.audience)
			if got != tc.want {
				t.Errorf("CanAccessAudience(%q, %q) = %v, want %v",
					tc.role, tc.audience, got, tc.want)
			}
		})
	}
}
