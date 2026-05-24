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

// TestVisibleAudiences_RoleMatrix pins v0.163.1 ADR-2 polish (defense-in-depth):
// repo-layer SQL filter receives the precomputed list of audiences a caller
// of the given role can see. Companion of CanAccessAudience (boolean check
// for one audience) — VisibleAudiences returns the full set.
func TestVisibleAudiences_RoleMatrix(t *testing.T) {
	tests := []struct {
		role string
		want []TargetAudience
	}{
		{
			role: "student",
			want: []TargetAudience{TargetAudienceAll, TargetAudienceStudents},
		},
		{
			role: "teacher",
			want: []TargetAudience{TargetAudienceAll, TargetAudienceTeachers},
		},
		{
			role: "methodist",
			want: []TargetAudience{TargetAudienceAll, TargetAudienceStaff},
		},
		{
			role: "academic_secretary",
			want: []TargetAudience{TargetAudienceAll, TargetAudienceStaff},
		},
		{
			role: "system_admin",
			want: []TargetAudience{
				TargetAudienceAll, TargetAudienceStudents, TargetAudienceTeachers,
				TargetAudienceStaff, TargetAudienceAdmins,
			},
		},
		{
			role: "",
			want: []TargetAudience{TargetAudienceAll},
		},
		{
			role: "unknown_role",
			want: []TargetAudience{TargetAudienceAll},
		},
	}

	for _, tc := range tests {
		t.Run(tc.role, func(t *testing.T) {
			got := VisibleAudiences(tc.role)
			if len(got) != len(tc.want) {
				t.Fatalf("VisibleAudiences(%q) length = %d, want %d (got=%v)",
					tc.role, len(got), len(tc.want), got)
			}
			gotSet := make(map[TargetAudience]bool, len(got))
			for _, a := range got {
				gotSet[a] = true
			}
			for _, a := range tc.want {
				if !gotSet[a] {
					t.Errorf("VisibleAudiences(%q) missing %q (got=%v)",
						tc.role, a, got)
				}
			}
		})
	}
}

// TestVisibleAudiences_ConsistentWithCanAccess pins the invariant
// that VisibleAudiences and CanAccessAudience agree pairwise. Defends
// against drift if one matrix is updated without the other.
func TestVisibleAudiences_ConsistentWithCanAccess(t *testing.T) {
	roles := []string{"student", "teacher", "methodist", "academic_secretary", "system_admin", "", "unknown_role"}
	allAudiences := []TargetAudience{
		TargetAudienceAll, TargetAudienceStudents, TargetAudienceTeachers,
		TargetAudienceStaff, TargetAudienceAdmins,
	}

	for _, role := range roles {
		visible := VisibleAudiences(role)
		visibleSet := make(map[TargetAudience]bool, len(visible))
		for _, a := range visible {
			visibleSet[a] = true
		}
		for _, audience := range allAudiences {
			canAccess := CanAccessAudience(role, audience)
			inVisible := visibleSet[audience]
			if canAccess != inVisible {
				t.Errorf("drift role=%q audience=%q: CanAccessAudience=%v but VisibleAudiences includes=%v",
					role, audience, canAccess, inVisible)
			}
		}
	}
}
