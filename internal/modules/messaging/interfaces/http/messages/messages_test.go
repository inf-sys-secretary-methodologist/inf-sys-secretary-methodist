package messages

import "testing"

// TestSystemMessages_Values pins the user-facing strings written by
// the messaging usecase on conversation lifecycle events. Drift on
// these texts is visible behavior — surfaced в conversation history
// UI for every participant — so a regression should fail CI before
// reaching the screenshot diff. Table-driven per CLAUDE.md gate
// (≥3-variant rule).
func TestSystemMessages_Values(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"group_created", SystemGroupCreated, "Group created"},
		{"user_joined", SystemUserJoined, "User joined the chat"},
		{"user_left", SystemUserLeft, "User left the chat"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("constant drift: got %q, want %q", tc.got, tc.want)
			}
		})
	}
}
