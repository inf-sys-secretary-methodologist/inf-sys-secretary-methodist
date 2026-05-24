package usecases

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

const testUpdatedTitle = "Updated"

// --- Helpers ---

// allAudiences enumerates every TargetAudience constant — the permissive
// "system_admin can see everything" set. Existing usecase tests use it
// for the GetByID / GetPinned / GetRecent audiences arg so they describe
// behavior independent of the v0.163.1 ADR-2 filter (which has its own
// dedicated tests).
var allAudiences = []domain.TargetAudience{
	domain.TargetAudienceAll,
	domain.TargetAudienceStudents,
	domain.TargetAudienceTeachers,
	domain.TargetAudienceStaff,
	domain.TargetAudienceAdmins,
}

func createTestAuditLogger() *logging.AuditLogger {
	logger := logging.NewLogger("error")
	return logging.NewAuditLogger(logger)
}

func createDefaultRequest() *dto.CreateAnnouncementRequest {
	return &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	}
}

// --- MockAnnouncementRepository ---

type MockAnnouncementRepository struct {
	announcements    map[int64]*entities.Announcement
	attachments      map[int64][]*entities.AnnouncementAttachment
	nextID           int64
	nextAttachmentID int64
}

func NewMockAnnouncementRepository() *MockAnnouncementRepository {
	return &MockAnnouncementRepository{
		announcements:    make(map[int64]*entities.Announcement),
		attachments:      make(map[int64][]*entities.AnnouncementAttachment),
		nextID:           1,
		nextAttachmentID: 1,
	}
}

func (m *MockAnnouncementRepository) Create(_ context.Context, announcement *entities.Announcement) error {
	announcement.ID = m.nextID
	m.nextID++
	m.announcements[announcement.ID] = announcement
	return nil
}

func (m *MockAnnouncementRepository) Save(_ context.Context, announcement *entities.Announcement) error {
	m.announcements[announcement.ID] = announcement
	return nil
}

func (m *MockAnnouncementRepository) GetByID(_ context.Context, id int64) (*entities.Announcement, error) {
	if a, exists := m.announcements[id]; exists {
		copiedAnn := *a
		return &copiedAnn, nil
	}
	return nil, nil
}

func (m *MockAnnouncementRepository) Delete(_ context.Context, id int64) error {
	delete(m.announcements, id)
	return nil
}

func (m *MockAnnouncementRepository) List(_ context.Context, _ AnnouncementFilter, limit, offset int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	i := 0
	for _, a := range m.announcements {
		if i >= offset && len(result) < limit {
			result = append(result, a)
		}
		i++
	}
	return result, nil
}

func (m *MockAnnouncementRepository) Count(_ context.Context, _ AnnouncementFilter) (int64, error) {
	return int64(len(m.announcements)), nil
}

func (m *MockAnnouncementRepository) GetByAuthor(_ context.Context, authorID int64, _, _ int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	for _, a := range m.announcements {
		if a.AuthorID == authorID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockAnnouncementRepository) GetPublished(_ context.Context, audience domain.TargetAudience, limit, offset int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	i := 0
	for _, a := range m.announcements {
		if a.Status == domain.AnnouncementStatusPublished && (a.TargetAudience == audience || a.TargetAudience == domain.TargetAudienceAll) {
			if i >= offset && len(result) < limit {
				result = append(result, a)
			}
			i++
		}
	}
	return result, nil
}

func (m *MockAnnouncementRepository) GetByIDForAudience(_ context.Context, id int64, audiences []domain.TargetAudience) (*entities.Announcement, error) {
	a, exists := m.announcements[id]
	if !exists {
		return nil, nil
	}
	if !audienceInSet(a.TargetAudience, audiences) {
		return nil, nil
	}
	copied := *a
	return &copied, nil
}

func (m *MockAnnouncementRepository) GetPinned(_ context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	for _, a := range m.announcements {
		if a.IsPinned && a.Status == domain.AnnouncementStatusPublished && audienceInSet(a.TargetAudience, audiences) {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAnnouncementRepository) GetRecent(_ context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	for _, a := range m.announcements {
		if a.Status == domain.AnnouncementStatusPublished && audienceInSet(a.TargetAudience, audiences) {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// audienceInSet reports whether `target` is one of `allowed`. Empty
// `allowed` returns false — explicit zero, matching the PG repo
// contract for an empty audience slice.
func audienceInSet(target domain.TargetAudience, allowed []domain.TargetAudience) bool {
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

func (m *MockAnnouncementRepository) IncrementViewCount(_ context.Context, id int64) error {
	if a, exists := m.announcements[id]; exists {
		a.ViewCount++
	}
	return nil
}

func (m *MockAnnouncementRepository) AddAttachment(_ context.Context, attachment *entities.AnnouncementAttachment) error {
	attachment.ID = m.nextAttachmentID
	m.nextAttachmentID++
	m.attachments[attachment.AnnouncementID] = append(m.attachments[attachment.AnnouncementID], attachment)
	return nil
}

func (m *MockAnnouncementRepository) RemoveAttachment(_ context.Context, attachmentID int64) error {
	for announcementID, atts := range m.attachments {
		for i, att := range atts {
			if att.ID == attachmentID {
				m.attachments[announcementID] = append(atts[:i], atts[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *MockAnnouncementRepository) GetAttachments(_ context.Context, announcementID int64) ([]*entities.AnnouncementAttachment, error) {
	return m.attachments[announcementID], nil
}

func (m *MockAnnouncementRepository) GetAttachmentByID(_ context.Context, attachmentID int64) (*entities.AnnouncementAttachment, error) {
	for _, atts := range m.attachments {
		for _, att := range atts {
			if att.ID == attachmentID {
				return att, nil
			}
		}
	}
	return nil, nil
}

// --- ErrorMockAnnouncementRepository ---

type ErrorMockAnnouncementRepository struct {
	MockAnnouncementRepository
	createErr         error
	saveErr           error
	getByIDErr        error
	deleteErr         error
	listErr           error
	countErr          error
	getPublishedErr   error
	getPinnedErr      error
	getRecentErr      error
	getAttachmentsErr error
}

func (m *ErrorMockAnnouncementRepository) Create(ctx context.Context, a *entities.Announcement) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.MockAnnouncementRepository.Create(ctx, a)
}

func (m *ErrorMockAnnouncementRepository) Save(ctx context.Context, a *entities.Announcement) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	return m.MockAnnouncementRepository.Save(ctx, a)
}

func (m *ErrorMockAnnouncementRepository) GetByID(ctx context.Context, id int64) (*entities.Announcement, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.MockAnnouncementRepository.GetByID(ctx, id)
}

func (m *ErrorMockAnnouncementRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.MockAnnouncementRepository.Delete(ctx, id)
}

func (m *ErrorMockAnnouncementRepository) List(ctx context.Context, f AnnouncementFilter, limit, offset int) ([]*entities.Announcement, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.MockAnnouncementRepository.List(ctx, f, limit, offset)
}

func (m *ErrorMockAnnouncementRepository) Count(ctx context.Context, f AnnouncementFilter) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.MockAnnouncementRepository.Count(ctx, f)
}

func (m *ErrorMockAnnouncementRepository) GetPublished(ctx context.Context, audience domain.TargetAudience, limit, offset int) ([]*entities.Announcement, error) {
	if m.getPublishedErr != nil {
		return nil, m.getPublishedErr
	}
	return m.MockAnnouncementRepository.GetPublished(ctx, audience, limit, offset)
}

func (m *ErrorMockAnnouncementRepository) GetByIDForAudience(ctx context.Context, id int64, audiences []domain.TargetAudience) (*entities.Announcement, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.MockAnnouncementRepository.GetByIDForAudience(ctx, id, audiences)
}

func (m *ErrorMockAnnouncementRepository) GetPinned(ctx context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error) {
	if m.getPinnedErr != nil {
		return nil, m.getPinnedErr
	}
	return m.MockAnnouncementRepository.GetPinned(ctx, audiences, limit)
}

func (m *ErrorMockAnnouncementRepository) GetRecent(ctx context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error) {
	if m.getRecentErr != nil {
		return nil, m.getRecentErr
	}
	return m.MockAnnouncementRepository.GetRecent(ctx, audiences, limit)
}

func (m *ErrorMockAnnouncementRepository) GetAttachments(ctx context.Context, announcementID int64) ([]*entities.AnnouncementAttachment, error) {
	if m.getAttachmentsErr != nil {
		return nil, m.getAttachmentsErr
	}
	return m.MockAnnouncementRepository.GetAttachments(ctx, announcementID)
}

// --- MockSystemNotifier ---

// MockSystemNotifier records every SendSystemNotification call so tests
// can assert audience-scoped fan-out hit each expected user with the
// expected title / summary.
type MockSystemNotifier struct {
	mu    sync.Mutex
	calls []notifierCall
	err   error
}

type notifierCall struct {
	ctx     context.Context
	userID  int64
	title   string
	summary string
}

func (m *MockSystemNotifier) SendSystemNotification(ctx context.Context, userID int64, title, summary string) error {
	m.mu.Lock()
	m.calls = append(m.calls, notifierCall{ctx: ctx, userID: userID, title: title, summary: summary})
	m.mu.Unlock()
	return m.err
}

func (m *MockSystemNotifier) callsSnapshot() []notifierCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]notifierCall, len(m.calls))
	copy(out, m.calls)
	return out
}

// --- MockUserIDsProvider ---

type MockUserIDsProvider struct {
	userIDs []int64
	err     error
	// Recorded args from the most recent call. Tests inspect these к
	// pin v0.163.1 polish guarantees: audience scoping (no
	// cross-audience push leak) and lifecycle-ctx propagation.
	lastAudience domain.TargetAudience
	lastCtx      context.Context
	mu           sync.Mutex
}

func (m *MockUserIDsProvider) GetUserIDsForAudience(ctx context.Context, audience domain.TargetAudience) ([]int64, error) {
	m.mu.Lock()
	m.lastAudience = audience
	m.lastCtx = ctx
	m.mu.Unlock()
	return m.userIDs, m.err
}

// snapshot returns the recorded last call args under the mutex.
func (m *MockUserIDsProvider) snapshot() (domain.TargetAudience, context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastAudience, m.lastCtx
}

// ============================
// Tests
// ============================

// --- Create ---

func TestAnnouncementUseCase_Create(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	req := &dto.CreateAnnouncementRequest{
		Title:          "Test Announcement",
		Content:        "Test Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	}

	announcement, err := uc.Create(ctx, 1, req)
	require.NoError(t, err)
	assert.NotZero(t, announcement.ID)
	assert.Equal(t, "Test Announcement", announcement.Title)
	assert.Equal(t, domain.AnnouncementStatusDraft, announcement.Status)
	assert.Equal(t, int64(1), announcement.AuthorID)
}

func TestAnnouncementUseCase_Create_InvalidPriority(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	req := &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriority("invalid"),
		TargetAudience: domain.TargetAudienceAll,
	}

	_, err := uc.Create(ctx, 1, req)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestAnnouncementUseCase_Create_InvalidAudience(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	req := &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudience("invalid"),
	}

	_, err := uc.Create(ctx, 1, req)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestAnnouncementUseCase_Create_WithOptionalFields(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	summary := "Test Summary"
	publishAt := time.Now().Add(24 * time.Hour)
	expireAt := time.Now().Add(7 * 24 * time.Hour)

	req := &dto.CreateAnnouncementRequest{
		Title:          "Test Announcement",
		Content:        "Test Content",
		Summary:        &summary,
		Priority:       domain.AnnouncementPriorityHigh,
		TargetAudience: domain.TargetAudienceTeachers,
		PublishAt:      &publishAt,
		ExpireAt:       &expireAt,
		IsPinned:       true,
		Tags:           []string{"important", "deadline"},
	}

	announcement, err := uc.Create(ctx, 1, req)
	require.NoError(t, err)
	require.NotNil(t, announcement.Summary)
	assert.Equal(t, summary, *announcement.Summary)
	assert.Equal(t, domain.AnnouncementPriorityHigh, announcement.Priority)
	assert.Equal(t, domain.TargetAudienceTeachers, announcement.TargetAudience)
	assert.True(t, announcement.IsPinned)
	assert.Len(t, announcement.Tags, 2)
}

func TestAnnouncementUseCase_Create_RepoError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		createErr:                  errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Create(ctx, 1, createDefaultRequest())
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_Create_WithAuditLogger(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	auditLogger := createTestAuditLogger()
	uc := NewAnnouncementUseCase(repo, auditLogger, nil, nil)
	ctx := context.Background()

	announcement, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)
	assert.NotZero(t, announcement.ID)
}

// --- GetByID ---

func TestAnnouncementUseCase_GetByID(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	announcement, err := uc.GetByID(ctx, created.ID, false, allAudiences)
	require.NoError(t, err)
	assert.Equal(t, created.ID, announcement.ID)
}

func TestAnnouncementUseCase_GetByID_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetByID(ctx, 999, false, allAudiences)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_GetByID_IncrementView(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	announcement, err := uc.GetByID(ctx, created.ID, true, allAudiences)
	require.NoError(t, err)
	assert.Equal(t, int64(1), announcement.ViewCount)
}

func TestAnnouncementUseCase_GetByID_NoIncrementForDraft(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	// Draft is not visible, so view should not increment even with incrementView=true
	announcement, err := uc.GetByID(ctx, created.ID, true, allAudiences)
	require.NoError(t, err)
	assert.Equal(t, int64(0), announcement.ViewCount)
}

func TestAnnouncementUseCase_GetByID_RepoError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getByIDErr:                 errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetByID(ctx, 1, false, allAudiences)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_GetByID_GetAttachmentsError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	// Create an announcement first
	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)

	// Now set the error so GetAttachments fails
	repo.getAttachmentsErr = errors.New("attachments error")

	_, err = uc.GetByID(ctx, created.ID, false, allAudiences)
	assert.Error(t, err)
	assert.Equal(t, "attachments error", err.Error())
}

func TestAnnouncementUseCase_GetByID_WithAttachments(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	// Add attachments directly to mock
	_ = repo.AddAttachment(ctx, &entities.AnnouncementAttachment{
		ID:             1,
		AnnouncementID: created.ID,
		FileName:       "test.pdf",
		FileSize:       1024,
		MimeType:       "application/pdf",
	})

	announcement, err := uc.GetByID(ctx, created.ID, false, allAudiences)
	require.NoError(t, err)
	assert.Len(t, announcement.Attachments, 1)
	assert.Equal(t, "test.pdf", announcement.Attachments[0].FileName)
}

// --- Update ---

func TestAnnouncementUseCase_Update(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Original Title",
		Content:        "Original Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	newTitle := "Updated Title"
	newContent := "Updated Content"
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{
		Title:   &newTitle,
		Content: &newContent,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "Updated Content", updated.Content)
}

func TestAnnouncementUseCase_Update_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	newTitle := "Test"
	_, err := uc.Update(ctx, 1, 999, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Update_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newTitle := testUpdatedTitle
	_, err := uc.Update(ctx, 2, created.ID, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAnnouncementUseCase_Update_AdminCanEdit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newTitle := "Updated by Admin"
	updated, err := uc.Update(ctx, 2, created.ID, true, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "Updated by Admin", updated.Title)
}

func TestAnnouncementUseCase_Update_Priority(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newPriority := domain.AnnouncementPriorityUrgent
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{Priority: &newPriority})
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementPriorityUrgent, updated.Priority)
}

func TestAnnouncementUseCase_Update_InvalidPriority(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	invalidPriority := domain.AnnouncementPriority("invalid")
	_, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{Priority: &invalidPriority})
	assert.Error(t, err)
}

func TestAnnouncementUseCase_Update_TargetAudience(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newAudience := domain.TargetAudienceStudents
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{TargetAudience: &newAudience})
	require.NoError(t, err)
	assert.Equal(t, domain.TargetAudienceStudents, updated.TargetAudience)
}

func TestAnnouncementUseCase_Update_InvalidTargetAudience(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	invalidAudience := domain.TargetAudience("invalid")
	_, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{TargetAudience: &invalidAudience})
	assert.Error(t, err)
}

func TestAnnouncementUseCase_Update_Summary(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newSummary := "New Summary"
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{Summary: &newSummary})
	require.NoError(t, err)
	require.NotNil(t, updated.Summary)
	assert.Equal(t, "New Summary", *updated.Summary)
}

func TestAnnouncementUseCase_Update_PublishSchedule(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	publishAt := time.Now().Add(24 * time.Hour)
	expireAt := time.Now().Add(7 * 24 * time.Hour)
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{
		PublishAt: &publishAt,
		ExpireAt:  &expireAt,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.PublishAt)
	require.NotNil(t, updated.ExpireAt)
}

func TestAnnouncementUseCase_Update_PublishAtOnly(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	publishAt := time.Now().Add(24 * time.Hour)
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{
		PublishAt: &publishAt,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.PublishAt)
}

func TestAnnouncementUseCase_Update_ExpireAtOnly(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	expireAt := time.Now().Add(7 * 24 * time.Hour)
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{
		ExpireAt: &expireAt,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.ExpireAt)
}

func TestAnnouncementUseCase_Update_Pin(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	// Pin
	isPinned := true
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{IsPinned: &isPinned})
	require.NoError(t, err)
	assert.True(t, updated.IsPinned)

	// Unpin
	isPinned = false
	updated, err = uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{IsPinned: &isPinned})
	require.NoError(t, err)
	assert.False(t, updated.IsPinned)
}

func TestAnnouncementUseCase_Update_Tags(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{
		Tags: []string{"tag1", "tag2", "tag3"},
	})
	require.NoError(t, err)
	assert.Len(t, updated.Tags, 3)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, updated.Tags)
}

func TestAnnouncementUseCase_Update_RepoGetByIDError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getByIDErr:                 errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	newTitle := testUpdatedTitle
	_, err := uc.Update(ctx, 1, 1, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_Update_RepoSaveError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")

	newTitle := testUpdatedTitle
	_, err = uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
}

func TestAnnouncementUseCase_Update_WithAuditLogger(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	auditLogger := createTestAuditLogger()
	uc := NewAnnouncementUseCase(repo, auditLogger, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newTitle := testUpdatedTitle
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, testUpdatedTitle, updated.Title)
}

func TestAnnouncementUseCase_Update_AllFieldsAtOnce(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	newTitle := "New Title"
	newContent := "New Content"
	newSummary := "New Summary"
	newPriority := domain.AnnouncementPriorityHigh
	newAudience := domain.TargetAudienceStaff
	publishAt := time.Now().Add(24 * time.Hour)
	expireAt := time.Now().Add(7 * 24 * time.Hour)
	isPinned := true
	tags := []string{"tag1"}

	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{
		Title:          &newTitle,
		Content:        &newContent,
		Summary:        &newSummary,
		Priority:       &newPriority,
		TargetAudience: &newAudience,
		PublishAt:      &publishAt,
		ExpireAt:       &expireAt,
		IsPinned:       &isPinned,
		Tags:           tags,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "New Content", updated.Content)
	require.NotNil(t, updated.Summary)
	assert.Equal(t, "New Summary", *updated.Summary)
	assert.Equal(t, domain.AnnouncementPriorityHigh, updated.Priority)
	assert.Equal(t, domain.TargetAudienceStaff, updated.TargetAudience)
	assert.True(t, updated.IsPinned)
	assert.Equal(t, []string{"tag1"}, updated.Tags)
}

// --- Delete ---

func TestAnnouncementUseCase_Delete(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	err := uc.Delete(ctx, 1, created.ID, false)
	require.NoError(t, err)

	_, err = uc.GetByID(ctx, created.ID, false, allAudiences)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Delete_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 999, false)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Delete_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	err := uc.Delete(ctx, 2, created.ID, false)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAnnouncementUseCase_Delete_AdminCanDelete(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	err := uc.Delete(ctx, 2, created.ID, true)
	require.NoError(t, err)

	_, err = uc.GetByID(ctx, created.ID, false, allAudiences)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Delete_RepoGetByIDError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getByIDErr:                 errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, false)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_Delete_RepoDeleteError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)

	repo.deleteErr = errors.New("delete error")

	err = uc.Delete(ctx, 1, created.ID, false)
	assert.Error(t, err)
	assert.Equal(t, "delete error", err.Error())
}

func TestAnnouncementUseCase_Delete_WithAuditLogger(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	auditLogger := createTestAuditLogger()
	uc := NewAnnouncementUseCase(repo, auditLogger, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	err := uc.Delete(ctx, 1, created.ID, false)
	require.NoError(t, err)
}

// --- List ---

func TestAnnouncementUseCase_List(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, _ = uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 1", Content: "C1", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	_, _ = uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 2", Content: "C2", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	_, _ = uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 3", Content: "C3", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})

	resp, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(3), resp.Total)
}

func TestAnnouncementUseCase_List_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	resp, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 0})
	require.NoError(t, err)
	assert.Equal(t, 20, resp.Limit)
}

func TestAnnouncementUseCase_List_NegativeLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	resp, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: -5})
	require.NoError(t, err)
	assert.Equal(t, 20, resp.Limit)
}

func TestAnnouncementUseCase_List_MaxLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	resp, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 500})
	require.NoError(t, err)
	assert.Equal(t, 100, resp.Limit)
}

func TestAnnouncementUseCase_List_RepoListError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		listErr:                    errors.New("list error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 10})
	assert.Error(t, err)
	assert.Equal(t, "list error", err.Error())
}

func TestAnnouncementUseCase_List_RepoCountError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		countErr:                   errors.New("count error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 10})
	assert.Error(t, err)
	assert.Equal(t, "count error", err.Error())
}

func TestAnnouncementUseCase_List_WithFilters(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	authorID := int64(1)
	status := domain.AnnouncementStatusDraft
	priority := domain.AnnouncementPriorityNormal
	audience := domain.TargetAudienceAll
	isPinned := false
	search := "test"

	_, _ = uc.Create(ctx, 1, createDefaultRequest())

	resp, err := uc.List(ctx, &dto.ListAnnouncementsRequest{
		AuthorID:       &authorID,
		Status:         &status,
		Priority:       &priority,
		TargetAudience: &audience,
		IsPinned:       &isPinned,
		Search:         &search,
		Tags:           []string{"tag1"},
		Limit:          10,
		Offset:         0,
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// --- GetPublished ---

func TestAnnouncementUseCase_GetPublished(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	a1, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 1", Content: "C1", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	a2, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 2", Content: "C2", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceTeachers})
	_, _ = uc.Publish(ctx, 1, a1.ID, false)
	_, _ = uc.Publish(ctx, 1, a2.ID, false)

	published, err := uc.GetPublished(ctx, domain.TargetAudienceAll, 10, 0)
	require.NoError(t, err)
	assert.Len(t, published, 1)
}

func TestAnnouncementUseCase_GetPublished_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetPublished(ctx, domain.TargetAudienceAll, 0, 0)
	require.NoError(t, err)
}

func TestAnnouncementUseCase_GetPublished_MaxLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	// With limit > 100, it should be clamped. We can't verify the clamped value directly
	// but at least it should not error.
	_, err := uc.GetPublished(ctx, domain.TargetAudienceAll, 200, 0)
	require.NoError(t, err)
}

func TestAnnouncementUseCase_GetPublished_RepoError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getPublishedErr:            errors.New("published error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetPublished(ctx, domain.TargetAudienceAll, 10, 0)
	assert.Error(t, err)
}

// --- GetPinned ---

func TestAnnouncementUseCase_GetPinned(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	a1, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Pinned",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
		IsPinned:       true,
	})
	_, _ = uc.Publish(ctx, 1, a1.ID, false)

	pinned, err := uc.GetPinned(ctx, allAudiences, 10)
	require.NoError(t, err)
	assert.Len(t, pinned, 1)
}

func TestAnnouncementUseCase_GetPinned_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetPinned(ctx, allAudiences, 0)
	require.NoError(t, err)
}

func TestAnnouncementUseCase_GetPinned_MaxLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetPinned(ctx, allAudiences, 50)
	require.NoError(t, err)
}

func TestAnnouncementUseCase_GetPinned_RepoError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getPinnedErr:               errors.New("pinned error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetPinned(ctx, allAudiences, 5)
	assert.Error(t, err)
}

// --- GetRecent ---

func TestAnnouncementUseCase_GetRecent(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	a1, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 1", Content: "C1", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	_, _ = uc.Publish(ctx, 1, a1.ID, false)

	recent, err := uc.GetRecent(ctx, allAudiences, 10)
	require.NoError(t, err)
	assert.Len(t, recent, 1)
}

func TestAnnouncementUseCase_GetRecent_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetRecent(ctx, allAudiences, 0)
	require.NoError(t, err)
}

func TestAnnouncementUseCase_GetRecent_MaxLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetRecent(ctx, allAudiences, 100)
	require.NoError(t, err)
}

func TestAnnouncementUseCase_GetRecent_RepoError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getRecentErr:               errors.New("recent error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.GetRecent(ctx, allAudiences, 10)
	assert.Error(t, err)
}

// --- Publish ---

func TestAnnouncementUseCase_Publish(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	published, err := uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusPublished, published.Status)
}

func TestAnnouncementUseCase_Publish_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Publish(ctx, 1, 999, false)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Publish_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	_, err := uc.Publish(ctx, 2, created.ID, false)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAnnouncementUseCase_Publish_AlreadyPublished(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	// Try to publish again
	_, err := uc.Publish(ctx, 1, created.ID, false)
	assert.Error(t, err)
}

func TestAnnouncementUseCase_Publish_RepoGetByIDError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getByIDErr:                 errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Publish(ctx, 1, 1, false)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_Publish_RepoSaveError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")

	_, err = uc.Publish(ctx, 1, created.ID, false)
	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
}

func TestAnnouncementUseCase_Publish_WithAuditLogger(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	auditLogger := createTestAuditLogger()
	uc := NewAnnouncementUseCase(repo, auditLogger, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	published, err := uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusPublished, published.Status)
}

func TestAnnouncementUseCase_Publish_NilNotificationUseCase(t *testing.T) {
	// When notificationUseCase is nil, publish should succeed without sending notifications
	repo := NewMockAnnouncementRepository()
	userProvider := &MockUserIDsProvider{
		userIDs: []int64{10, 20, 30},
	}
	// notificationUseCase is nil, userIDsProvider is set
	uc := NewAnnouncementUseCase(repo, nil, nil, userProvider)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	published, err := uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusPublished, published.Status)
}

func TestAnnouncementUseCase_Publish_NilUserIDsProvider(t *testing.T) {
	// When userIDsProvider is nil, publish should succeed without sending notifications
	repo := NewMockAnnouncementRepository()
	// Both notificationUseCase and userIDsProvider must be non-nil for notifications
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	published, err := uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusPublished, published.Status)
}

// TestAnnouncementUseCase_Publish_FanOutScopedToTargetAudience pins
// v0.163.1 polish guarantee #1: the broadcast fan-out provider is
// called with the announcement's target_audience (not "all users")
// so a student-targeted announcement only pushes к students. Closes
// v0.163.0 audit T1-7 cross-audience push leak.
func TestAnnouncementUseCase_Publish_FanOutScopedToTargetAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience domain.TargetAudience
	}{
		{"students-only", domain.TargetAudienceStudents},
		{"teachers-only", domain.TargetAudienceTeachers},
		{"admins-only", domain.TargetAudienceAdmins},
		{"staff-only", domain.TargetAudienceStaff},
		{"broadcast-all", domain.TargetAudienceAll},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockAnnouncementRepository()
			provider := &MockUserIDsProvider{userIDs: []int64{10}}
			notifier := &MockSystemNotifier{}
			uc := NewAnnouncementUseCase(repo, nil, notifier, provider)
			ctx := context.Background()

			req := createDefaultRequest()
			req.TargetAudience = tc.audience
			created, err := uc.Create(ctx, 1, req)
			require.NoError(t, err)

			_, err = uc.Publish(ctx, 1, created.ID, false)
			require.NoError(t, err)

			waitForFanOut(t, provider)
			gotAudience, _ := provider.snapshot()
			assert.Equalf(t, tc.audience, gotAudience,
				"expected provider к receive audience=%s, got %s",
				tc.audience, gotAudience)
		})
	}
}

// TestAnnouncementUseCase_Publish_FanOutUsesLifecycleContext pins
// v0.163.1 polish guarantee #2: the broadcast goroutine runs on the
// ctx registered via WithLifecycleContext, not context.Background().
// Graceful shutdown can therefore cancel in-flight sends rather than
// leak goroutines past server stop.
func TestAnnouncementUseCase_Publish_FanOutUsesLifecycleContext(t *testing.T) {
	type ctxKey struct{}
	sentinel := "v0.163.1-lifecycle-ctx"
	lifecycleCtx := context.WithValue(context.Background(), ctxKey{}, sentinel)

	repo := NewMockAnnouncementRepository()
	provider := &MockUserIDsProvider{userIDs: []int64{10}}
	notifier := &MockSystemNotifier{}
	uc := NewAnnouncementUseCase(repo, nil, notifier, provider).
		WithLifecycleContext(lifecycleCtx)
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)

	_, err = uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)

	waitForFanOut(t, provider)
	_, gotCtx := provider.snapshot()
	require.NotNil(t, gotCtx)
	got, ok := gotCtx.Value(ctxKey{}).(string)
	assert.True(t, ok, "ctx must carry sentinel value")
	assert.Equal(t, sentinel, got)

	// Notifier should also have received the lifecycle ctx for each
	// fanned-out recipient.
	calls := notifier.callsSnapshot()
	require.Len(t, calls, 1)
	notifGot, ok := calls[0].ctx.Value(ctxKey{}).(string)
	assert.True(t, ok, "notifier ctx must carry sentinel value")
	assert.Equal(t, sentinel, notifGot)
}

// waitForFanOut polls until the spy provider records a call or fails
// the test on timeout. Fan-out is fire-and-forget goroutine so the
// assertions need к wait briefly.
func waitForFanOut(t *testing.T, p *MockUserIDsProvider) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, ctx := p.snapshot()
		if ctx != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout waiting for broadcast fan-out goroutine к call provider")
}

func TestAnnouncementUseCase_Publish_AdminCanPublish(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	// Admin (user 2) can publish user 1's announcement
	published, err := uc.Publish(ctx, 2, created.ID, true)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusPublished, published.Status)
}

// --- Unpublish ---

func TestAnnouncementUseCase_Unpublish(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	unpublished, err := uc.Unpublish(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusDraft, unpublished.Status)
}

func TestAnnouncementUseCase_Unpublish_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Unpublish(ctx, 1, 999, false)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Unpublish_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	_, err := uc.Unpublish(ctx, 2, created.ID, false)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAnnouncementUseCase_Unpublish_NotPublished(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	// Draft cannot be unpublished
	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	_, err := uc.Unpublish(ctx, 1, created.ID, false)
	assert.Error(t, err)
}

func TestAnnouncementUseCase_Unpublish_RepoGetByIDError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getByIDErr:                 errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Unpublish(ctx, 1, 1, false)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_Unpublish_RepoSaveError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)
	_, err = uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")

	_, err = uc.Unpublish(ctx, 1, created.ID, false)
	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
}

func TestAnnouncementUseCase_Unpublish_WithAuditLogger(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	auditLogger := createTestAuditLogger()
	uc := NewAnnouncementUseCase(repo, auditLogger, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	unpublished, err := uc.Unpublish(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusDraft, unpublished.Status)
}

func TestAnnouncementUseCase_Unpublish_AdminCanUnpublish(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	unpublished, err := uc.Unpublish(ctx, 2, created.ID, true)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusDraft, unpublished.Status)
}

// --- Archive ---

func TestAnnouncementUseCase_Archive(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	archived, err := uc.Archive(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusArchived, archived.Status)
}

func TestAnnouncementUseCase_Archive_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Archive(ctx, 1, 999, false)
	assert.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementUseCase_Archive_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	_, err := uc.Archive(ctx, 2, created.ID, false)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAnnouncementUseCase_Archive_AdminCanArchive(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	archived, err := uc.Archive(ctx, 2, created.ID, true)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusArchived, archived.Status)
}

func TestAnnouncementUseCase_Archive_NotPublished(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	// Draft cannot be archived
	created, _ := uc.Create(ctx, 1, createDefaultRequest())

	_, err := uc.Archive(ctx, 1, created.ID, false)
	assert.Error(t, err)
}

func TestAnnouncementUseCase_Archive_RepoGetByIDError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
		getByIDErr:                 errors.New("db error"),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.Archive(ctx, 1, 1, false)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestAnnouncementUseCase_Archive_RepoSaveError(t *testing.T) {
	repo := &ErrorMockAnnouncementRepository{
		MockAnnouncementRepository: *NewMockAnnouncementRepository(),
	}
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, createDefaultRequest())
	require.NoError(t, err)
	_, err = uc.Publish(ctx, 1, created.ID, false)
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")

	_, err = uc.Archive(ctx, 1, created.ID, false)
	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
}

func TestAnnouncementUseCase_Archive_WithAuditLogger(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	auditLogger := createTestAuditLogger()
	uc := NewAnnouncementUseCase(repo, auditLogger, nil, nil)
	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, createDefaultRequest())
	_, _ = uc.Publish(ctx, 1, created.ID, false)

	archived, err := uc.Archive(ctx, 1, created.ID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.AnnouncementStatusArchived, archived.Status)
}
