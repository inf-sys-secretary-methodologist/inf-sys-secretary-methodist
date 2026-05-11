package auditlog_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/auditlog"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"

	"github.com/stretchr/testify/require"
)

// fakeReader records the AuditLogFilter it was called with and
// returns canned data. The recording variant is preferred over a
// testify Mock so the test reads like ordinary table-driven code.
type fakeReader struct {
	captured logging.AuditLogFilter
	calls    int
	items    []*logging.AuditLog
	total    int
	err      error
}

func (f *fakeReader) List(_ context.Context, filter logging.AuditLogFilter) (logging.AuditLogListResult, error) {
	f.captured = filter
	f.calls++
	if f.err != nil {
		return logging.AuditLogListResult{}, f.err
	}
	return logging.AuditLogListResult{Items: f.items, Total: f.total}, nil
}

func TestAdminAuditLogUseCase_List_LimitClamping(t *testing.T) {
	cases := []struct {
		name  string
		input int
		want  int
	}{
		{"zero falls back to DefaultLimit", 0, auditlog.DefaultLimit},
		{"negative falls back to DefaultLimit", -1, auditlog.DefaultLimit},
		{"value under MaxLimit passes through", 100, 100},
		{"value at MaxLimit passes through", auditlog.MaxLimit, auditlog.MaxLimit},
		{"value above MaxLimit clamps to MaxLimit", 1000, auditlog.MaxLimit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeReader{}
			uc := auditlog.NewAdminAuditLogUseCase(fake)

			_, err := uc.List(context.Background(), auditlog.ListInput{Limit: tc.input})
			require.NoError(t, err)
			require.Equal(t, tc.want, fake.captured.Limit)
		})
	}
}

func TestAdminAuditLogUseCase_List_NegativeOffsetClampsToZero(t *testing.T) {
	fake := &fakeReader{}
	uc := auditlog.NewAdminAuditLogUseCase(fake)

	_, err := uc.List(context.Background(), auditlog.ListInput{Offset: -5})
	require.NoError(t, err)
	require.Equal(t, 0, fake.captured.Offset)
}

func TestAdminAuditLogUseCase_List_PassesAllFiltersThrough(t *testing.T) {
	fake := &fakeReader{}
	uc := auditlog.NewAdminAuditLogUseCase(fake)

	userID := int64(42)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)

	_, err := uc.List(context.Background(), auditlog.ListInput{
		Action:   "curriculum.approved",
		Resource: "curriculum",
		UserID:   &userID,
		From:     &from,
		To:       &to,
		Limit:    25,
		Offset:   50,
	})
	require.NoError(t, err)

	require.Equal(t, "curriculum.approved", fake.captured.Action)
	require.Equal(t, "curriculum", fake.captured.Resource)
	require.NotNil(t, fake.captured.UserID)
	require.Equal(t, int64(42), *fake.captured.UserID)
	require.NotNil(t, fake.captured.From)
	require.Equal(t, from, *fake.captured.From)
	require.NotNil(t, fake.captured.To)
	require.Equal(t, to, *fake.captured.To)
	require.Equal(t, 25, fake.captured.Limit)
	require.Equal(t, 50, fake.captured.Offset)
}

func TestAdminAuditLogUseCase_List_InvalidTimeRangeReturnsSentinel(t *testing.T) {
	cases := []struct {
		name string
		from time.Time
		to   time.Time
	}{
		{
			"from equals to",
			time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"from after to",
			time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeReader{}
			uc := auditlog.NewAdminAuditLogUseCase(fake)

			_, err := uc.List(context.Background(), auditlog.ListInput{
				From: &tc.from,
				To:   &tc.to,
			})
			require.ErrorIs(t, err, auditlog.ErrInvalidTimeRange)
			require.Zero(t, fake.calls, "reader must not be invoked on invalid range")
		})
	}
}

func TestAdminAuditLogUseCase_List_PropagatesReaderError(t *testing.T) {
	wantErr := errors.New("db connection refused")
	fake := &fakeReader{err: wantErr}
	uc := auditlog.NewAdminAuditLogUseCase(fake)

	_, err := uc.List(context.Background(), auditlog.ListInput{})
	require.ErrorIs(t, err, wantErr)
}

func TestAdminAuditLogUseCase_List_ReturnsReaderResultUnchanged(t *testing.T) {
	actorID := int64(7)
	items := []*logging.AuditLog{
		{ID: 11, Action: "auth.login", Resource: "session", ActorUserID: &actorID},
		{ID: 12, Action: "auth.logout", Resource: "session", ActorUserID: &actorID},
	}
	fake := &fakeReader{items: items, total: 99}
	uc := auditlog.NewAdminAuditLogUseCase(fake)

	result, err := uc.List(context.Background(), auditlog.ListInput{})
	require.NoError(t, err)
	require.Equal(t, 99, result.Total)
	require.Len(t, result.Items, 2)
	require.Equal(t, int64(11), result.Items[0].ID)
	require.Equal(t, int64(12), result.Items[1].ID)
}

func TestNewAdminAuditLogUseCase_PanicsOnNilReader(t *testing.T) {
	require.PanicsWithValue(t, "auditlog: nil AuditLogReader", func() {
		auditlog.NewAdminAuditLogUseCase(nil)
	})
}
