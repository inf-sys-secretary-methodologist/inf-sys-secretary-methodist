package persistence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// newDocActivityReaderMock spins a sqlmock-backed DocumentActivityReaderPG.
// Mirrors newDocRepoMock for the parent DocumentRepositoryPG but builds
// the narrow reader directly — verifies the narrow port has its own
// construction path independent of the full repo.
func newDocActivityReaderMock(t *testing.T) (*DocumentActivityReaderPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDocumentActivityReaderPG(db), mock
}

func TestDocumentActivityReaderPG_AggregateActivityByType(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		rows *sqlmock.Rows
		want []usecases.DocumentActivityByTypeAgg
	}{
		{
			name: "empty range returns nil slice",
			rows: sqlmock.NewRows([]string{"name", "status", "count"}),
			want: nil,
		},
		{
			name: "multiple types and statuses grouped",
			rows: sqlmock.NewRows([]string{"name", "status", "count"}).
				AddRow("Приказ", "approved", 5).
				AddRow("Приказ", "draft", 2).
				AddRow("Письмо", "approved", 7),
			want: []usecases.DocumentActivityByTypeAgg{
				{TypeName: "Приказ", Status: entities.DocumentStatusApproved, Count: 5},
				{TypeName: "Приказ", Status: entities.DocumentStatusDraft, Count: 2},
				{TypeName: "Письмо", Status: entities.DocumentStatusApproved, Count: 7},
			},
		},
		{
			name: "single type single status",
			rows: sqlmock.NewRows([]string{"name", "status", "count"}).
				AddRow("Протокол", "registered", 1),
			want: []usecases.DocumentActivityByTypeAgg{
				{TypeName: "Протокол", Status: entities.DocumentStatusRegistered, Count: 1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader, mock := newDocActivityReaderMock(t)

			mock.ExpectQuery(`SELECT dt.name, d.status, COUNT\(\*\) FROM documents d\s+JOIN document_types dt ON dt.id = d.document_type_id\s+WHERE d.created_at >= \$1 AND d.created_at < \$2\s+GROUP BY dt.name, d.status`).
				WithArgs(from, to).
				WillReturnRows(tc.rows)

			got, err := reader.AggregateActivityByType(context.Background(), from, to)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("query error propagates wrapped", func(t *testing.T) {
		reader, mock := newDocActivityReaderMock(t)

		mock.ExpectQuery(`SELECT dt.name, d.status, COUNT\(\*\) FROM documents`).
			WithArgs(from, to).
			WillReturnError(fmt.Errorf("conn refused"))

		got, err := reader.AggregateActivityByType(context.Background(), from, to)
		require.ErrorContains(t, err, "documents: aggregate activity by type")
		require.ErrorContains(t, err, "conn refused")
		require.Nil(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
