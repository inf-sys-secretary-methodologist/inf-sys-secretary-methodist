package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/repositories"
)

// MockAnnouncementRepository implements AnnouncementRepository for testing.
type MockAnnouncementRepository struct {
	announcements map[int64]*entities.Announcement
	attachments   map[int64][]*entities.AnnouncementAttachment
	nextID        int64
}

func NewMockAnnouncementRepository() *MockAnnouncementRepository {
	return &MockAnnouncementRepository{
		announcements: make(map[int64]*entities.Announcement),
		attachments:   make(map[int64][]*entities.AnnouncementAttachment),
		nextID:        1,
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
		// Return a copy to avoid pointer aliasing issues
		copy := *a
		return &copy, nil
	}
	return nil, nil
}

func (m *MockAnnouncementRepository) Delete(_ context.Context, id int64) error {
	delete(m.announcements, id)
	return nil
}

func (m *MockAnnouncementRepository) List(_ context.Context, _ repositories.AnnouncementFilter, limit, offset int) ([]*entities.Announcement, error) {
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

func (m *MockAnnouncementRepository) Count(_ context.Context, _ repositories.AnnouncementFilter) (int64, error) {
	return int64(len(m.announcements)), nil
}

func (m *MockAnnouncementRepository) GetByAuthor(_ context.Context, authorID int64, limit, offset int) ([]*entities.Announcement, error) {
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

func (m *MockAnnouncementRepository) GetPinned(_ context.Context, limit int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	for _, a := range m.announcements {
		if a.IsPinned && a.Status == domain.AnnouncementStatusPublished {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAnnouncementRepository) GetRecent(_ context.Context, limit int) ([]*entities.Announcement, error) {
	var result []*entities.Announcement
	for _, a := range m.announcements {
		if a.Status == domain.AnnouncementStatusPublished {
			result = append(result, a)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockAnnouncementRepository) IncrementViewCount(_ context.Context, id int64) error {
	if a, exists := m.announcements[id]; exists {
		a.ViewCount++
	}
	return nil
}

func (m *MockAnnouncementRepository) AddAttachment(_ context.Context, attachment *entities.AnnouncementAttachment) error {
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

// Tests

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
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if announcement.ID == 0 {
		t.Error("expected announcement ID to be set")
	}

	if announcement.Title != "Test Announcement" {
		t.Errorf("expected title 'Test Announcement', got '%s'", announcement.Title)
	}

	if announcement.Status != domain.AnnouncementStatusDraft {
		t.Errorf("expected status 'draft', got '%s'", announcement.Status)
	}

	if announcement.AuthorID != 1 {
		t.Errorf("expected author ID 1, got %d", announcement.AuthorID)
	}
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
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
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
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
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
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if announcement.Summary == nil || *announcement.Summary != summary {
		t.Errorf("expected summary '%s', got %v", summary, announcement.Summary)
	}

	if announcement.Priority != domain.AnnouncementPriorityHigh {
		t.Errorf("expected priority 'high', got '%s'", announcement.Priority)
	}

	if announcement.TargetAudience != domain.TargetAudienceTeachers {
		t.Errorf("expected audience 'teachers', got '%s'", announcement.TargetAudience)
	}

	if !announcement.IsPinned {
		t.Error("expected IsPinned to be true")
	}

	if len(announcement.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(announcement.Tags))
	}
}

func TestAnnouncementUseCase_GetByID(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create announcement
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Get by ID
	announcement, err := uc.GetByID(ctx, created.ID, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if announcement.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, announcement.ID)
	}
}

func TestAnnouncementUseCase_GetByID_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	_, err := uc.GetByID(ctx, 999, false)
	if err != ErrAnnouncementNotFound {
		t.Errorf("expected ErrAnnouncementNotFound, got %v", err)
	}
}

func TestAnnouncementUseCase_GetByID_IncrementView(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create and publish announcement
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})
	uc.Publish(ctx, 1, created.ID, false)

	// Get with view increment
	initialView := created.ViewCount
	announcement, _ := uc.GetByID(ctx, created.ID, true)

	if announcement.ViewCount != initialView+1 {
		t.Errorf("expected view count %d, got %d", initialView+1, announcement.ViewCount)
	}
}

func TestAnnouncementUseCase_Update(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create announcement
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Original Title",
		Content:        "Original Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Update
	newTitle := "Updated Title"
	newContent := "Updated Content"
	req := &dto.UpdateAnnouncementRequest{
		Title:   &newTitle,
		Content: &newContent,
	}

	updated, err := uc.Update(ctx, 1, created.ID, false, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
	}

	if updated.Content != "Updated Content" {
		t.Errorf("expected content 'Updated Content', got '%s'", updated.Content)
	}
}

func TestAnnouncementUseCase_Update_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()
	newTitle := "Test"

	_, err := uc.Update(ctx, 1, 999, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	if err != ErrAnnouncementNotFound {
		t.Errorf("expected ErrAnnouncementNotFound, got %v", err)
	}
}

func TestAnnouncementUseCase_Update_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create announcement by user 1
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Try to update by user 2 (not admin)
	newTitle := "Updated"
	_, err := uc.Update(ctx, 2, created.ID, false, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAnnouncementUseCase_Update_AdminCanEdit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create announcement by user 1
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Admin (user 2) can update
	newTitle := "Updated by Admin"
	updated, err := uc.Update(ctx, 2, created.ID, true, &dto.UpdateAnnouncementRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Title != "Updated by Admin" {
		t.Errorf("expected title 'Updated by Admin', got '%s'", updated.Title)
	}
}

func TestAnnouncementUseCase_Update_Priority(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	newPriority := domain.AnnouncementPriorityUrgent
	updated, err := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{Priority: &newPriority})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Priority != domain.AnnouncementPriorityUrgent {
		t.Errorf("expected priority 'urgent', got '%s'", updated.Priority)
	}
}

func TestAnnouncementUseCase_Update_Pin(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Pin
	isPinned := true
	updated, _ := uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{IsPinned: &isPinned})
	if !updated.IsPinned {
		t.Error("expected IsPinned to be true")
	}

	// Unpin
	isPinned = false
	updated, _ = uc.Update(ctx, 1, created.ID, false, &dto.UpdateAnnouncementRequest{IsPinned: &isPinned})
	if updated.IsPinned {
		t.Error("expected IsPinned to be false")
	}
}

func TestAnnouncementUseCase_Delete(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create announcement
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Delete
	err := uc.Delete(ctx, 1, created.ID, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = uc.GetByID(ctx, created.ID, false)
	if err != ErrAnnouncementNotFound {
		t.Error("expected announcement to be deleted")
	}
}

func TestAnnouncementUseCase_Delete_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	err := uc.Delete(ctx, 1, 999, false)
	if err != ErrAnnouncementNotFound {
		t.Errorf("expected ErrAnnouncementNotFound, got %v", err)
	}
}

func TestAnnouncementUseCase_Delete_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create by user 1
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Try to delete by user 2
	err := uc.Delete(ctx, 2, created.ID, false)
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAnnouncementUseCase_List(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create announcements
	uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 1", Content: "C1", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 2", Content: "C2", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 3", Content: "C3", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})

	// List
	resp, err := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 10})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("expected total 3, got %d", resp.Total)
	}
}

func TestAnnouncementUseCase_List_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	resp, _ := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 0})
	if resp.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", resp.Limit)
	}
}

func TestAnnouncementUseCase_List_MaxLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	resp, _ := uc.List(ctx, &dto.ListAnnouncementsRequest{Limit: 500})
	if resp.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", resp.Limit)
	}
}

func TestAnnouncementUseCase_GetPublished(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create and publish announcements
	a1, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 1", Content: "C1", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	a2, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 2", Content: "C2", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceTeachers})
	uc.Publish(ctx, 1, a1.ID, false)
	uc.Publish(ctx, 1, a2.ID, false)

	// Get published for all
	published, err := uc.GetPublished(ctx, domain.TargetAudienceAll, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(published) != 1 {
		t.Errorf("expected 1 published for all, got %d", len(published))
	}
}

func TestAnnouncementUseCase_GetPublished_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// This should not panic with limit 0
	_, err := uc.GetPublished(ctx, domain.TargetAudienceAll, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAnnouncementUseCase_GetPinned(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create, pin and publish
	a1, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Pinned",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
		IsPinned:       true,
	})
	uc.Publish(ctx, 1, a1.ID, false)

	// Get pinned
	pinned, err := uc.GetPinned(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(pinned) != 1 {
		t.Errorf("expected 1 pinned, got %d", len(pinned))
	}
}

func TestAnnouncementUseCase_GetPinned_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	_, err := uc.GetPinned(ctx, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAnnouncementUseCase_GetRecent(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create and publish
	a1, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{Title: "Test 1", Content: "C1", Priority: domain.AnnouncementPriorityNormal, TargetAudience: domain.TargetAudienceAll})
	uc.Publish(ctx, 1, a1.ID, false)

	// Get recent
	recent, err := uc.GetRecent(ctx, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(recent) != 1 {
		t.Errorf("expected 1 recent, got %d", len(recent))
	}
}

func TestAnnouncementUseCase_GetRecent_DefaultLimit(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	_, err := uc.GetRecent(ctx, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAnnouncementUseCase_Publish(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	// Publish
	published, err := uc.Publish(ctx, 1, created.ID, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if published.Status != domain.AnnouncementStatusPublished {
		t.Errorf("expected status 'published', got '%s'", published.Status)
	}
}

func TestAnnouncementUseCase_Publish_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	_, err := uc.Publish(ctx, 1, 999, false)
	if err != ErrAnnouncementNotFound {
		t.Errorf("expected ErrAnnouncementNotFound, got %v", err)
	}
}

func TestAnnouncementUseCase_Publish_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	_, err := uc.Publish(ctx, 2, created.ID, false)
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAnnouncementUseCase_Unpublish(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create and publish
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})
	uc.Publish(ctx, 1, created.ID, false)

	// Unpublish
	unpublished, err := uc.Unpublish(ctx, 1, created.ID, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if unpublished.Status != domain.AnnouncementStatusDraft {
		t.Errorf("expected status 'draft', got '%s'", unpublished.Status)
	}
}

func TestAnnouncementUseCase_Unpublish_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	_, err := uc.Unpublish(ctx, 1, 999, false)
	if err != ErrAnnouncementNotFound {
		t.Errorf("expected ErrAnnouncementNotFound, got %v", err)
	}
}

func TestAnnouncementUseCase_Archive(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	// Create and publish
	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})
	uc.Publish(ctx, 1, created.ID, false)

	// Archive
	archived, err := uc.Archive(ctx, 1, created.ID, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if archived.Status != domain.AnnouncementStatusArchived {
		t.Errorf("expected status 'archived', got '%s'", archived.Status)
	}
}

func TestAnnouncementUseCase_Archive_NotFound(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	_, err := uc.Archive(ctx, 1, 999, false)
	if err != ErrAnnouncementNotFound {
		t.Errorf("expected ErrAnnouncementNotFound, got %v", err)
	}
}

func TestAnnouncementUseCase_Archive_Unauthorized(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})

	_, err := uc.Archive(ctx, 2, created.ID, false)
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAnnouncementUseCase_Archive_AdminCanArchive(t *testing.T) {
	repo := NewMockAnnouncementRepository()
	uc := NewAnnouncementUseCase(repo, nil, nil, nil)

	ctx := context.Background()

	created, _ := uc.Create(ctx, 1, &dto.CreateAnnouncementRequest{
		Title:          "Test",
		Content:        "Content",
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
	})
	uc.Publish(ctx, 1, created.ID, false)

	// Admin can archive
	archived, err := uc.Archive(ctx, 2, created.ID, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if archived.Status != domain.AnnouncementStatusArchived {
		t.Errorf("expected status 'archived', got '%s'", archived.Status)
	}
}
