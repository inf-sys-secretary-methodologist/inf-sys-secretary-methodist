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

// TestRegisterDocumentUseCase pins the v0.149.0 Register usecase
// contract: load → entity.Register → repo.Update + AuditSink emit
// on each outcome.
//
// Issue: #230
func TestRegisterDocumentUseCase(t *testing.T) {
	now := time.Date(2026, 5, 16, 16, 0, 0, 0, time.UTC)
	registrarID := int64(7)
	validNumber := "01-2026/123"
	tooShortNumber := "ab"

	cases := []struct {
		name      string
		startDoc  *entities.Document
		number    string
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproved), number: validNumber, wantAudit: "document.registered"},
		{name: "not-found", startDoc: nil, number: validNumber, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.register_denied", wantDeny: "not_found"},
		{name: "non-approved status rejected", startDoc: draftDoc(1, 42), number: validNumber, wantErr: entities.ErrCannotRegister, wantAudit: "document.register_denied", wantDeny: "not_approved"},
		{name: "short number rejected", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproved), number: tooShortNumber, wantErr: entities.ErrInvalidRegistrationNumber, wantAudit: "document.register_denied", wantDeny: "invalid_number"},
		{name: "empty number rejected", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproved), number: "", wantErr: entities.ErrInvalidRegistrationNumber, wantAudit: "document.register_denied", wantDeny: "invalid_number"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewRegisterDocumentUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), registrarID, usecases.RegisterDocumentInput{ID: 1, Number: tc.number})

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
			assert.Equal(t, entities.DocumentStatusRegistered, got.Status)
			assert.NotNil(t, got.RegisteredBy)
			assert.Equal(t, registrarID, *got.RegisteredBy)
			assert.NotNil(t, got.RegistrationNumber)
			assert.Equal(t, validNumber, *got.RegistrationNumber)
			assert.Equal(t, "document.registered", audit.Last().Action)
		})
	}
}
