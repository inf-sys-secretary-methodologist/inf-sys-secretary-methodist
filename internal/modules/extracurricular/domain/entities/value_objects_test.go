package entities

import "testing"

func TestCategory_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		input Category
		want  bool
	}{
		{name: "cultural", input: CategoryCultural, want: true},
		{name: "sports", input: CategorySports, want: true},
		{name: "recreational", input: CategoryRecreational, want: true},
		{name: "educational", input: CategoryEducational, want: true},
		{name: "other", input: CategoryOther, want: true},
		{name: "empty rejected", input: Category(""), want: false},
		{name: "uppercase rejected", input: Category("CULTURAL"), want: false},
		{name: "bogus rejected", input: Category("bogus"), want: false},
		{name: "whitespace rejected", input: Category(" cultural "), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.IsValid(); got != tt.want {
				t.Errorf("Category(%q).IsValid() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTargetAudience_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		input TargetAudience
		want  bool
	}{
		{name: "all", input: TargetAudienceAll, want: true},
		{name: "students", input: TargetAudienceStudents, want: true},
		{name: "teachers", input: TargetAudienceTeachers, want: true},
		{name: "staff", input: TargetAudienceStaff, want: true},
		{name: "empty rejected", input: TargetAudience(""), want: false},
		{name: "admins rejected — admins are not target", input: TargetAudience("admins"), want: false},
		{name: "uppercase rejected", input: TargetAudience("ALL"), want: false},
		{name: "bogus rejected", input: TargetAudience("bogus"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.IsValid(); got != tt.want {
				t.Errorf("TargetAudience(%q).IsValid() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		input Status
		want  bool
	}{
		{name: "draft", input: StatusDraft, want: true},
		{name: "published", input: StatusPublished, want: true},
		{name: "canceled", input: StatusCanceled, want: true},
		{name: "completed", input: StatusCompleted, want: true},
		{name: "empty rejected", input: Status(""), want: false},
		{name: "uppercase rejected", input: Status("DRAFT"), want: false},
		{name: "bogus rejected", input: Status("bogus"), want: false},
		{name: "BrE double-l form rejected — canonical is single-l", input: Status("can" + "celled"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.IsValid(); got != tt.want {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStatus_CanEdit(t *testing.T) {
	tests := []struct {
		input Status
		want  bool
	}{
		{StatusDraft, true},
		{StatusPublished, true},
		{StatusCanceled, false},
		{StatusCompleted, false},
		{Status("bogus"), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			if got := tt.input.CanEdit(); got != tt.want {
				t.Errorf("Status(%q).CanEdit() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStatus_CanRegister(t *testing.T) {
	tests := []struct {
		input Status
		want  bool
	}{
		{StatusDraft, false},
		{StatusPublished, true},
		{StatusCanceled, false},
		{StatusCompleted, false},
		{Status("bogus"), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			if got := tt.input.CanRegister(); got != tt.want {
				t.Errorf("Status(%q).CanRegister() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
