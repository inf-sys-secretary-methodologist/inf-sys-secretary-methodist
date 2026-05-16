package usecases

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
)

// --- Fakes ---

type fakeBrandRepo struct {
	mu        sync.Mutex
	settings  *entities.BrandSettings
	getErr    error
	updateErr error
	updates   int
}

func (r *fakeBrandRepo) Get(_ context.Context) (*entities.BrandSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settings, r.getErr
}

func (r *fakeBrandRepo) Update(_ context.Context, s *entities.BrandSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.updateErr != nil {
		return r.updateErr
	}
	r.settings = s
	r.updates++
	return nil
}

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

type spyAudit struct {
	mu     sync.Mutex
	events []audited
	calls  int
}

type audited struct {
	action   string
	resource string
	fields   map[string]any
}

func (a *spyAudit) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls++
	a.events = append(a.events, audited{action: action, resource: resource, fields: fields})
}

func newSettings(t *testing.T, now time.Time) *entities.BrandSettings {
	t.Helper()
	s, err := entities.NewBrandSettings(
		"App", "Tagline", "https://logo", "https://favicon",
		"#aabbcc", "#112233", now,
	)
	require.NoError(t, err)
	return s
}

// --- SystemClock ---

func TestSystemClock_NowReturnsRecent(t *testing.T) {
	before := time.Now()
	got := SystemClock{}.Now()
	after := time.Now()
	assert.False(t, got.Before(before), "SystemClock.Now() before reference start")
	assert.False(t, got.After(after), "SystemClock.Now() after reference end")
}

// --- GetBrandingUseCase ---

func TestNewGetBrandingUseCase_PanicsOnNilRepo(t *testing.T) {
	assert.PanicsWithValue(t, "branding: nil BrandSettingsRepository", func() {
		NewGetBrandingUseCase(nil)
	})
}

func TestGetBrandingUseCase_Execute_ReturnsRepoResult(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	expected := newSettings(t, now)
	repo := &fakeBrandRepo{settings: expected}
	uc := NewGetBrandingUseCase(repo)

	got, err := uc.Execute(context.Background())
	require.NoError(t, err)
	assert.Same(t, expected, got)
}

func TestGetBrandingUseCase_Execute_PropagatesRepoError(t *testing.T) {
	repo := &fakeBrandRepo{getErr: errors.New("db down")}
	uc := NewGetBrandingUseCase(repo)

	got, err := uc.Execute(context.Background())
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "db down")
}

// --- UpdateBrandingUseCase ---

func TestNewUpdateBrandingUseCase_PanicsOnNilRepo(t *testing.T) {
	assert.PanicsWithValue(t, "branding: nil BrandSettingsRepository", func() {
		NewUpdateBrandingUseCase(nil, fakeClock{}, nil)
	})
}

func TestNewUpdateBrandingUseCase_DefaultsClockToSystem(t *testing.T) {
	repo := &fakeBrandRepo{}
	uc := NewUpdateBrandingUseCase(repo, nil, nil)
	_, ok := uc.clock.(SystemClock)
	assert.True(t, ok, "nil clock should default to SystemClock")
}

func TestUpdateBrandingUseCase_Execute_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	repo := &fakeBrandRepo{}
	audit := &spyAudit{}
	uc := NewUpdateBrandingUseCase(repo, fakeClock{now: now}, audit)

	got, err := uc.Execute(context.Background(), UpdateBrandingInput{
		AppName: "App", Tagline: "Tag", LogoURL: "https://logo",
		FaviconURL: "https://favicon", PrimaryColor: "#abcdef",
		SecondaryColor: "#fedcba", ActorUserID: 42,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "App", got.AppName())
	assert.Equal(t, now, got.UpdatedAt())
	assert.Equal(t, 1, repo.updates)
	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "brand.updated", ev.action)
	assert.Equal(t, "brand", ev.resource)
	assert.Equal(t, int64(42), ev.fields["actor_user_id"])
	assert.Equal(t, "App", ev.fields["app_name"])
}

func TestUpdateBrandingUseCase_Execute_DomainValidationError(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	repo := &fakeBrandRepo{}
	audit := &spyAudit{}
	uc := NewUpdateBrandingUseCase(repo, fakeClock{now: now}, audit)

	// AppName empty triggers domain invariant violation.
	got, err := uc.Execute(context.Background(), UpdateBrandingInput{
		AppName: "", Tagline: "Tag", PrimaryColor: "#abcdef",
		SecondaryColor: "#fedcba", ActorUserID: 1,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, 0, repo.updates, "repo Update should not be called on validation error")
	assert.Empty(t, audit.events, "no audit emission on validation error")
}

func TestUpdateBrandingUseCase_Execute_RepoError(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	repo := &fakeBrandRepo{updateErr: errors.New("constraint")}
	audit := &spyAudit{}
	uc := NewUpdateBrandingUseCase(repo, fakeClock{now: now}, audit)

	got, err := uc.Execute(context.Background(), UpdateBrandingInput{
		AppName: "App", Tagline: "Tag", LogoURL: "https://l",
		FaviconURL: "https://f", PrimaryColor: "#abcdef",
		SecondaryColor: "#fedcba", ActorUserID: 1,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "constraint")
	assert.Empty(t, audit.events, "no audit emission on repo error")
}

func TestUpdateBrandingUseCase_Execute_NilAuditSkips(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	repo := &fakeBrandRepo{}
	uc := NewUpdateBrandingUseCase(repo, fakeClock{now: now}, nil)

	got, err := uc.Execute(context.Background(), UpdateBrandingInput{
		AppName: "App", Tagline: "Tag", LogoURL: "https://l",
		FaviconURL: "https://f", PrimaryColor: "#abcdef",
		SecondaryColor: "#fedcba", ActorUserID: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, repo.updates)
}
