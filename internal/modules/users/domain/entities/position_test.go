package entities

import "testing"

func TestNewPosition(t *testing.T) {
	tests := []struct {
		name        string
		posName     string
		code        string
		description string
		level       int
	}{
		{
			name:        "create manager position",
			posName:     "Manager",
			code:        "MGR",
			description: "Team manager",
			level:       3,
		},
		{
			name:        "create developer position",
			posName:     "Software Developer",
			code:        "DEV",
			description: "Software development",
			level:       2,
		},
		{
			name:        "create junior position",
			posName:     "Junior Developer",
			code:        "JR_DEV",
			description: "",
			level:       1,
		},
		{
			name:        "create ceo position",
			posName:     "CEO",
			code:        "CEO",
			description: "Chief Executive Officer",
			level:       10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := NewPosition(tt.posName, tt.code, tt.description, tt.level)

			if pos.Name != tt.posName {
				t.Errorf("expected name %q, got %q", tt.posName, pos.Name)
			}
			if pos.Code != tt.code {
				t.Errorf("expected code %q, got %q", tt.code, pos.Code)
			}
			if pos.Description != tt.description {
				t.Errorf("expected description %q, got %q", tt.description, pos.Description)
			}
			if pos.Level != tt.level {
				t.Errorf("expected level %d, got %d", tt.level, pos.Level)
			}
			if !pos.IsActive {
				t.Error("expected position to be active")
			}
			if pos.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}
			if pos.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}
		})
	}
}
