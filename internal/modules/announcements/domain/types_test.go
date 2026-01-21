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
