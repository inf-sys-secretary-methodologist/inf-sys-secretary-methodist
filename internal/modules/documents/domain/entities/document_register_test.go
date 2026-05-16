package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestDocumentRegister_TableDriven pins the v0.149.0 Register transition:
//   - approved → registered (sets registration_number/date + registered_by)
//   - any non-approved → ErrCannotRegister
//   - empty/short number → ErrInvalidRegistrationNumber
//
// Issue: #230
func TestDocumentRegister_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC)
	registrarID := int64(7)

	cases := []struct {
		name       string
		startState DocumentStatus
		number     string
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "approved → registered", startState: DocumentStatusApproved, number: "01-2026/123", wantStatus: DocumentStatusRegistered},
		{name: "approved + ws-padded → registered (trims)", startState: DocumentStatusApproved, number: "  01-2026/124  ", wantStatus: DocumentStatusRegistered},
		{name: "draft → err", startState: DocumentStatusDraft, number: "01-2026/125", wantErr: ErrCannotRegister},
		{name: "approval → err", startState: DocumentStatusApproval, number: "01-2026/126", wantErr: ErrCannotRegister},
		{name: "registered → err (already)", startState: DocumentStatusRegistered, number: "01-2026/127", wantErr: ErrCannotRegister},
		{name: "rejected → err", startState: DocumentStatusRejected, number: "01-2026/128", wantErr: ErrCannotRegister},
		{name: "empty number → err", startState: DocumentStatusApproved, number: "", wantErr: ErrInvalidRegistrationNumber},
		{name: "ws-only number → err", startState: DocumentStatusApproved, number: "   ", wantErr: ErrInvalidRegistrationNumber},
		{name: "too short number → err", startState: DocumentStatusApproved, number: "ab", wantErr: ErrInvalidRegistrationNumber},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, 42)
			d.Status = tc.startState
			err := d.Register(tc.number, registrarID, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if d.Status != tc.startState {
					t.Errorf("status mutated on rejected transition: %q → %q", tc.startState, d.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, d.Status)
			}
			if d.RegisteredBy == nil || *d.RegisteredBy != registrarID {
				t.Errorf("expected RegisteredBy=%d, got %v", registrarID, d.RegisteredBy)
			}
			if d.RegistrationDate == nil || !d.RegistrationDate.Equal(now) {
				t.Errorf("expected RegistrationDate=%v, got %v", now, d.RegistrationDate)
			}
			if d.RegistrationNumber == nil || *d.RegistrationNumber != strings.TrimSpace(tc.number) {
				t.Errorf("expected RegistrationNumber=%q (trimmed), got %v", strings.TrimSpace(tc.number), d.RegistrationNumber)
			}
		})
	}
}
