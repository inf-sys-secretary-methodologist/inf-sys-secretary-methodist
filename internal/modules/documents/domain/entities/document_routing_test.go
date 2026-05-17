package entities

import (
	"errors"
	"testing"
	"time"
)

// TestDocumentSendToRouting_TableDriven pins the v0.150.0 SendToRouting
// transition:
//   - registered → routing (sets RoutedBy + RoutedAt)
//   - any non-registered → ErrCannotRoute (no state mutation)
//
// Issue: #231
func TestDocumentSendToRouting_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 17, 9, 0, 0, 0, time.UTC)
	routerID := int64(7)

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "registered → routing", startState: DocumentStatusRegistered, wantStatus: DocumentStatusRouting},
		{name: "draft → err", startState: DocumentStatusDraft, wantErr: ErrCannotRoute},
		{name: "approval → err", startState: DocumentStatusApproval, wantErr: ErrCannotRoute},
		{name: "approved → err (must register first)", startState: DocumentStatusApproved, wantErr: ErrCannotRoute},
		{name: "routing → err (already)", startState: DocumentStatusRouting, wantErr: ErrCannotRoute},
		{name: "execution → err", startState: DocumentStatusExecution, wantErr: ErrCannotRoute},
		{name: "rejected → err", startState: DocumentStatusRejected, wantErr: ErrCannotRoute},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, 42)
			d.Status = tc.startState
			err := d.SendToRouting(routerID, now)

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
			if d.RoutedBy == nil || *d.RoutedBy != routerID {
				t.Errorf("expected RoutedBy=%d, got %v", routerID, d.RoutedBy)
			}
			if d.RoutedAt == nil || !d.RoutedAt.Equal(now) {
				t.Errorf("expected RoutedAt=%v, got %v", now, d.RoutedAt)
			}
		})
	}
}

// TestDocumentSignVisa_TableDriven pins the v0.150.0 SignVisa transition:
//   - routing → execution (sets VisaSignedBy + VisaSignedAt)
//   - any non-routing → ErrCannotSignVisa (no state mutation)
//
// Issue: #231
func TestDocumentSignVisa_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC)
	visaID := int64(9)

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "routing → execution", startState: DocumentStatusRouting, wantStatus: DocumentStatusExecution},
		{name: "draft → err", startState: DocumentStatusDraft, wantErr: ErrCannotSignVisa},
		{name: "approved → err", startState: DocumentStatusApproved, wantErr: ErrCannotSignVisa},
		{name: "registered → err (must route first)", startState: DocumentStatusRegistered, wantErr: ErrCannotSignVisa},
		{name: "execution → err (already past visa)", startState: DocumentStatusExecution, wantErr: ErrCannotSignVisa},
		{name: "executed → err", startState: DocumentStatusExecuted, wantErr: ErrCannotSignVisa},
		{name: "rejected → err", startState: DocumentStatusRejected, wantErr: ErrCannotSignVisa},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, 42)
			d.Status = tc.startState
			err := d.SignVisa(visaID, now)

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
			if d.VisaSignedBy == nil || *d.VisaSignedBy != visaID {
				t.Errorf("expected VisaSignedBy=%d, got %v", visaID, d.VisaSignedBy)
			}
			if d.VisaSignedAt == nil || !d.VisaSignedAt.Equal(now) {
				t.Errorf("expected VisaSignedAt=%v, got %v", now, d.VisaSignedAt)
			}
		})
	}
}
