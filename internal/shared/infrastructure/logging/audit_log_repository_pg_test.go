package logging

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

// listSelectColumns mirrors the projection used by AuditLogRepositoryPG.List
// — kept here as a single source so test rows stay in sync with the
// production SELECT.
var listSelectColumns = []string{
	"id", "created_at", "action", "resource",
	"actor_user_id", "actor_ip", "correlation_id", "fields",
}

// makeListRow builds one *sqlmock.Rows row matching the projection;
// nullable cells take typed `sql.NullX` values so the scan path goes
// through the same conversions the production code performs.
func makeListRow(id int64, createdAt time.Time, action, resource string,
	userID *int64, ip, correlation *string, fields map[string]any,
) []driver.Value {
	var userVal driver.Value
	if userID != nil {
		userVal = *userID
	}
	var ipVal driver.Value
	if ip != nil {
		ipVal = *ip
	}
	var corrVal driver.Value
	if correlation != nil {
		corrVal = *correlation
	}
	fieldsJSON, _ := json.Marshal(fields)
	return []driver.Value{
		id, createdAt, action, resource, userVal, ipVal, corrVal, fieldsJSON,
	}
}

func TestAuditLogRepositoryPG_List(t *testing.T) {
	actorID := int64(42)
	actorIP := "10.0.0.5"
	corrID := "req-7c4f"
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 10, 12, 30, 0, 0, time.UTC)

	type expectedArgs struct {
		action   string
		resource string
		user     sql.NullInt64
		from     sql.NullTime
		to       sql.NullTime
	}

	cases := []struct {
		name      string
		filter    AuditLogFilter
		wantArgs  expectedArgs
		total     int
		rowsCount int
	}{
		{
			name:   "no filters — all dimensions empty/null",
			filter: AuditLogFilter{Limit: 50, Offset: 0},
			wantArgs: expectedArgs{
				action: "", resource: "",
				user: sql.NullInt64{Valid: false},
				from: sql.NullTime{Valid: false},
				to:   sql.NullTime{Valid: false},
			},
			total: 7, rowsCount: 2,
		},
		{
			name:   "action filter routes literally",
			filter: AuditLogFilter{Action: "curriculum.approved", Limit: 50},
			wantArgs: expectedArgs{
				action: "curriculum.approved", resource: "",
				user: sql.NullInt64{Valid: false},
				from: sql.NullTime{Valid: false},
				to:   sql.NullTime{Valid: false},
			},
			total: 3, rowsCount: 3,
		},
		{
			name:   "resource filter routes literally",
			filter: AuditLogFilter{Resource: "document", Limit: 50},
			wantArgs: expectedArgs{
				action: "", resource: "document",
				user: sql.NullInt64{Valid: false},
				from: sql.NullTime{Valid: false},
				to:   sql.NullTime{Valid: false},
			},
			total: 5, rowsCount: 1,
		},
		{
			name:   "user_id filter wraps in NullInt64",
			filter: AuditLogFilter{UserID: &actorID, Limit: 50},
			wantArgs: expectedArgs{
				action: "", resource: "",
				user: sql.NullInt64{Int64: actorID, Valid: true},
				from: sql.NullTime{Valid: false},
				to:   sql.NullTime{Valid: false},
			},
			total: 2, rowsCount: 2,
		},
		{
			name:   "from filter wraps in NullTime as inclusive lower bound",
			filter: AuditLogFilter{From: &from, Limit: 50},
			wantArgs: expectedArgs{
				action: "", resource: "",
				user: sql.NullInt64{Valid: false},
				from: sql.NullTime{Time: from, Valid: true},
				to:   sql.NullTime{Valid: false},
			},
			total: 4, rowsCount: 1,
		},
		{
			name:   "to filter wraps in NullTime as exclusive upper bound",
			filter: AuditLogFilter{To: &to, Limit: 50},
			wantArgs: expectedArgs{
				action: "", resource: "",
				user: sql.NullInt64{Valid: false},
				from: sql.NullTime{Valid: false},
				to:   sql.NullTime{Time: to, Valid: true},
			},
			total: 6, rowsCount: 1,
		},
		{
			name: "all filters combined honor sentinel-arg style",
			filter: AuditLogFilter{
				Action:   "auth.login",
				Resource: "session",
				UserID:   &actorID,
				From:     &from,
				To:       &to,
				Limit:    25,
				Offset:   50,
			},
			wantArgs: expectedArgs{
				action: "auth.login", resource: "session",
				user: sql.NullInt64{Int64: actorID, Valid: true},
				from: sql.NullTime{Time: from, Valid: true},
				to:   sql.NullTime{Time: to, Valid: true},
			},
			total: 1, rowsCount: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newAuditRepoMock(t)

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM audit_logs`)).
				WithArgs(tc.wantArgs.action, tc.wantArgs.resource, tc.wantArgs.user, tc.wantArgs.from, tc.wantArgs.to).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.total))

			rows := sqlmock.NewRows(listSelectColumns)
			for i := 0; i < tc.rowsCount; i++ {
				rows.AddRow(makeListRow(int64(100+i), createdAt, "curriculum.approved", "curriculum",
					&actorID, &actorIP, &corrID, map[string]any{"curriculum_id": int64(7)})...)
			}
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, action, resource`)).
				WithArgs(tc.wantArgs.action, tc.wantArgs.resource, tc.wantArgs.user, tc.wantArgs.from, tc.wantArgs.to,
					tc.filter.Limit, tc.filter.Offset).
				WillReturnRows(rows)

			result, err := repo.List(context.Background(), tc.filter)
			require.NoError(t, err)
			require.Equal(t, tc.total, result.Total)
			require.Len(t, result.Items, tc.rowsCount)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("count query error propagates wrapped", func(t *testing.T) {
		repo, mock := newAuditRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM audit_logs`)).
			WillReturnError(fmt.Errorf("conn refused"))

		_, err := repo.List(context.Background(), AuditLogFilter{Limit: 50})
		require.ErrorContains(t, err, "audit_logs: count")
		require.ErrorContains(t, err, "conn refused")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list query error propagates wrapped", func(t *testing.T) {
		repo, mock := newAuditRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM audit_logs`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, action, resource`)).
			WillReturnError(fmt.Errorf("query failed"))

		_, err := repo.List(context.Background(), AuditLogFilter{Limit: 50})
		require.ErrorContains(t, err, "audit_logs: list")
		require.ErrorContains(t, err, "query failed")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("row scan error propagates wrapped", func(t *testing.T) {
		repo, mock := newAuditRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM audit_logs`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		// id column receives non-int → Scan fails on first row
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, action, resource`)).
			WillReturnRows(sqlmock.NewRows(listSelectColumns).
				AddRow("bad-id", createdAt, "x", "y", nil, nil, nil, []byte("{}")))

		_, err := repo.List(context.Background(), AuditLogFilter{Limit: 50})
		require.ErrorContains(t, err, "audit_logs: list scan")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("results scan back into AuditLog DTO with nullable fields preserved", func(t *testing.T) {
		repo, mock := newAuditRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM audit_logs`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		row1 := makeListRow(11, createdAt, "auth.login", "session",
			&actorID, &actorIP, &corrID, map[string]any{"role": "methodist"})
		row2 := makeListRow(12, createdAt.Add(time.Minute), "auth.logout", "session",
			nil, nil, nil, map[string]any{}) // all nullable cells empty

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, action, resource`)).
			WillReturnRows(sqlmock.NewRows(listSelectColumns).AddRow(row1...).AddRow(row2...))

		result, err := repo.List(context.Background(), AuditLogFilter{Limit: 50})
		require.NoError(t, err)
		require.Len(t, result.Items, 2)

		require.Equal(t, int64(11), result.Items[0].ID)
		require.Equal(t, "auth.login", result.Items[0].Action)
		require.NotNil(t, result.Items[0].ActorUserID)
		require.Equal(t, actorID, *result.Items[0].ActorUserID)
		require.NotNil(t, result.Items[0].ActorIP)
		require.Equal(t, actorIP, *result.Items[0].ActorIP)
		require.NotNil(t, result.Items[0].CorrelationID)
		require.Equal(t, "methodist", result.Items[0].Fields["role"])

		require.Equal(t, int64(12), result.Items[1].ID)
		require.Nil(t, result.Items[1].ActorUserID)
		require.Nil(t, result.Items[1].ActorIP)
		require.Nil(t, result.Items[1].CorrelationID)
		require.Empty(t, result.Items[1].Fields)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
