package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// TestStartRoutingUseCase pins the v0.150.0 StartRoutingUseCase contract:
// load → entity.SendToRouting → repo.Update + AuditSink emit on every
// outcome. Mirror к RegisterDocumentUseCase pattern.
//
// Issue: #231
func TestStartRoutingUseCase(t *testing.T) {
	now := time.Date(2026, 5, 17, 9, 0, 0, 0, time.UTC)
	routerID := int64(7)

	cases := []struct {
		name      string
		startDoc  *entities.Document
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusRegistered), wantAudit: "document.routed"},
		{name: "not-found", startDoc: nil, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.route_denied", wantDeny: "not_found"},
		{name: "non-registered status rejected", startDoc: draftDoc(1, 42), wantErr: entities.ErrCannotRoute, wantAudit: "document.route_denied", wantDeny: "not_registered"},
		{name: "approved status rejected (must register first)", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproved), wantErr: entities.ErrCannotRoute, wantAudit: "document.route_denied", wantDeny: "not_registered"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewStartRoutingUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), routerID, usecases.StartRoutingInput{ID: 1})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				rec := audit.Last()
				assert.Equal(t, tc.wantAudit, rec.Action)
				if tc.wantDeny != "" {
					assert.Equal(t, tc.wantDeny, rec.Fields["reason"])
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, entities.DocumentStatusRouting, got.Status)
			assert.NotNil(t, got.RoutedBy)
			assert.Equal(t, routerID, *got.RoutedBy)
			assert.Equal(t, "document.routed", audit.Last().Action)
		})
	}
}

// TestSignVisaUseCase pins the v0.150.0 SignVisaUseCase contract:
// load → entity.SignVisa → repo.Update + AuditSink emit on every
// outcome.
//
// Issue: #231
func TestSignVisaUseCase(t *testing.T) {
	now := time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC)
	visaID := int64(9)

	cases := []struct {
		name      string
		startDoc  *entities.Document
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusRouting), wantAudit: "document.visa_signed"},
		{name: "not-found", startDoc: nil, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.sign_visa_denied", wantDeny: "not_found"},
		{name: "non-routing status rejected", startDoc: draftDoc(1, 42), wantErr: entities.ErrCannotSignVisa, wantAudit: "document.sign_visa_denied", wantDeny: "not_routing"},
		{name: "registered status rejected (must route first)", startDoc: docAtStatus(1, 42, entities.DocumentStatusRegistered), wantErr: entities.ErrCannotSignVisa, wantAudit: "document.sign_visa_denied", wantDeny: "not_routing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewSignVisaUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), visaID, usecases.SignVisaInput{ID: 1})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				rec := audit.Last()
				assert.Equal(t, tc.wantAudit, rec.Action)
				if tc.wantDeny != "" {
					assert.Equal(t, tc.wantDeny, rec.Fields["reason"])
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, entities.DocumentStatusExecution, got.Status)
			assert.NotNil(t, got.VisaSignedBy)
			assert.Equal(t, visaID, *got.VisaSignedBy)
			assert.Equal(t, "document.visa_signed", audit.Last().Action)
		})
	}
}
