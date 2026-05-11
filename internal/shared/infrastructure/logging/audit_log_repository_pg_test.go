package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// newAuditRepoMock spins a sqlmock-backed AuditLogRepositoryPG.
func newAuditRepoMock(t *testing.T) (*AuditLogRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewAuditLogRepositoryPG(db), mock
}

func TestAuditLogRepositoryPG_Write(t *testing.T) {
	actorID := int64(42)
	actorIP := "10.0.0.5"
	correlationID := "req-7c4f"

	cases := []struct {
		name       string
		log        *AuditLog
		wantSQL    string
		assertArgs func(t *testing.T, expect *sqlmock.ExpectedExec)
	}{
		{
			name: "all fields populated",
			log: &AuditLog{
				Action:        "curriculum.created",
				Resource:      "curriculum",
				ActorUserID:   &actorID,
				ActorIP:       &actorIP,
				CorrelationID: &correlationID,
				Fields:        map[string]any{"curriculum_id": int64(7), "title": "ИС-21"},
			},
			wantSQL: `INSERT INTO audit_logs`,
			assertArgs: func(t *testing.T, expect *sqlmock.ExpectedExec) {
				fieldsJSON, _ := json.Marshal(map[string]any{"curriculum_id": int64(7), "title": "ИС-21"})
				expect.WithArgs("curriculum.created", "curriculum", &actorID, &actorIP, &correlationID, fieldsJSON).
					WillReturnResult(sqlmock.NewResult(101, 1))
			},
		},
		{
			name: "nullable fields nil",
			log: &AuditLog{
				Action:   "auth.login",
				Resource: "session",
				Fields:   map[string]any{},
			},
			wantSQL: `INSERT INTO audit_logs`,
			assertArgs: func(t *testing.T, expect *sqlmock.ExpectedExec) {
				fieldsJSON, _ := json.Marshal(map[string]any{})
				expect.WithArgs("auth.login", "session", (*int64)(nil), (*string)(nil), (*string)(nil), fieldsJSON).
					WillReturnResult(sqlmock.NewResult(102, 1))
			},
		},
		{
			name: "fields with mixed types",
			log: &AuditLog{
				Action:        "document.deleted",
				Resource:      "document",
				ActorUserID:   &actorID,
				CorrelationID: &correlationID,
				Fields: map[string]any{
					"document_id": int64(99),
					"reason":      "duplicate",
					"is_public":   false,
				},
			},
			wantSQL: `INSERT INTO audit_logs`,
			assertArgs: func(t *testing.T, expect *sqlmock.ExpectedExec) {
				fieldsJSON, _ := json.Marshal(map[string]any{
					"document_id": int64(99),
					"reason":      "duplicate",
					"is_public":   false,
				})
				expect.WithArgs("document.deleted", "document", &actorID, (*string)(nil), &correlationID, fieldsJSON).
					WillReturnResult(sqlmock.NewResult(103, 1))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newAuditRepoMock(t)
			expect := mock.ExpectExec(regexp.QuoteMeta(tc.wantSQL))
			tc.assertArgs(t, expect)

			err := repo.Write(context.Background(), tc.log)
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("transport error propagates wrapped", func(t *testing.T) {
		repo, mock := newAuditRepoMock(t)

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO audit_logs`)).
			WillReturnError(fmt.Errorf("conn refused"))

		err := repo.Write(context.Background(), &AuditLog{
			Action:   "x",
			Resource: "y",
			Fields:   map[string]any{},
		})
		require.ErrorContains(t, err, "audit_logs: write")
		require.ErrorContains(t, err, "conn refused")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fields marshal error returns wrapped", func(t *testing.T) {
		repo, _ := newAuditRepoMock(t)

		// Functions cannot be JSON-marshaled — triggers json.Marshal error
		// before any DB exec; mock.ExpectExec deliberately not set.
		err := repo.Write(context.Background(), &AuditLog{
			Action:   "test.fail",
			Resource: "test",
			Fields:   map[string]any{"bad": func() {}},
		})
		require.ErrorContains(t, err, "audit_logs: marshal fields")
	})

	// Confirm CreatedAt is server-side default (DEFAULT CURRENT_TIMESTAMP)
	// — client does NOT pass it; this is also a structural pin so future
	// authors do not add a 7th positional arg by accident.
	t.Run("CreatedAt is not passed to INSERT", func(t *testing.T) {
		repo, mock := newAuditRepoMock(t)
		fieldsJSON, _ := json.Marshal(map[string]any{})
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO audit_logs`)).
			WithArgs("any", "any", (*int64)(nil), (*string)(nil), (*string)(nil), fieldsJSON).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Write(context.Background(), &AuditLog{
			Action:    "any",
			Resource:  "any",
			Fields:    map[string]any{},
			CreatedAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), // ignored
		})
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
