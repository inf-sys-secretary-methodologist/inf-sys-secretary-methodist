package entities

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
)

func TestNewAnnouncement(t *testing.T) {
	title := "Test Title"
	content := "Test Content"
	authorID := int64(1)

	a := NewAnnouncement(title, content, authorID)

	if a.Title != title {
		t.Errorf("expected title %q, got %q", title, a.Title)
	}
	if a.Content != content {
		t.Errorf("expected content %q, got %q", content, a.Content)
	}
	if a.AuthorID != authorID {
		t.Errorf("expected author ID %d, got %d", authorID, a.AuthorID)
	}
	if a.Status != domain.AnnouncementStatusDraft {
		t.Errorf("expected status %q, got %q", domain.AnnouncementStatusDraft, a.Status)
	}
	if a.Priority != domain.AnnouncementPriorityNormal {
		t.Errorf("expected priority %q, got %q", domain.AnnouncementPriorityNormal, a.Priority)
	}
	if a.TargetAudience != domain.TargetAudienceAll {
		t.Errorf("expected audience %q, got %q", domain.TargetAudienceAll, a.TargetAudience)
	}
	if a.IsPinned {
		t.Error("expected not pinned")
	}
	if a.ViewCount != 0 {
		t.Errorf("expected view count 0, got %d", a.ViewCount)
	}
}

func TestAnnouncement_Publish(t *testing.T) {
	tests := []struct {
		name    string
		status  domain.AnnouncementStatus
		wantErr error
	}{
		{
			name:    "publish draft",
			status:  domain.AnnouncementStatusDraft,
			wantErr: nil,
		},
		{
			name:    "publish already published",
			status:  domain.AnnouncementStatusPublished,
			wantErr: ErrAnnouncementAlreadyPublished,
		},
		{
			name:    "publish archived",
			status:  domain.AnnouncementStatusArchived,
			wantErr: ErrAnnouncementArchived,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnnouncement("Test", "Content", 1)
			a.Status = tt.status

			err := a.Publish()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if a.Status != domain.AnnouncementStatusPublished {
				t.Errorf("expected status %q, got %q", domain.AnnouncementStatusPublished, a.Status)
			}

			if a.PublishAt == nil {
				t.Error("expected publish_at to be set")
			}
		})
	}
}

func TestAnnouncement_Archive(t *testing.T) {
	tests := []struct {
		name    string
		status  domain.AnnouncementStatus
		wantErr error
	}{
		{
			name:    "archive published",
			status:  domain.AnnouncementStatusPublished,
			wantErr: nil,
		},
		{
			name:    "archive draft",
			status:  domain.AnnouncementStatusDraft,
			wantErr: ErrAnnouncementNotPublished,
		},
		{
			name:    "archive already archived",
			status:  domain.AnnouncementStatusArchived,
			wantErr: ErrAnnouncementArchived,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnnouncement("Test", "Content", 1)
			a.Status = tt.status

			err := a.Archive()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if a.Status != domain.AnnouncementStatusArchived {
				t.Errorf("expected status %q, got %q", domain.AnnouncementStatusArchived, a.Status)
			}
		})
	}
}

func TestAnnouncement_CanEdit(t *testing.T) {
	tests := []struct {
		name     string
		authorID int64
		status   domain.AnnouncementStatus
		userID   int64
		isAdmin  bool
		want     bool
	}{
		{
			name:     "author can edit draft",
			authorID: 1,
			status:   domain.AnnouncementStatusDraft,
			userID:   1,
			isAdmin:  false,
			want:     true,
		},
		{
			name:     "non-author cannot edit",
			authorID: 1,
			status:   domain.AnnouncementStatusDraft,
			userID:   2,
			isAdmin:  false,
			want:     false,
		},
		{
			name:     "admin can edit any",
			authorID: 1,
			status:   domain.AnnouncementStatusDraft,
			userID:   2,
			isAdmin:  true,
			want:     true,
		},
		{
			name:     "cannot edit archived",
			authorID: 1,
			status:   domain.AnnouncementStatusArchived,
			userID:   1,
			isAdmin:  false,
			want:     false,
		},
		{
			name:     "admin can edit archived",
			authorID: 1,
			status:   domain.AnnouncementStatusArchived,
			userID:   2,
			isAdmin:  true,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnnouncement("Test", "Content", tt.authorID)
			a.Status = tt.status

			got := a.CanEdit(tt.userID, tt.isAdmin)
			if got != tt.want {
				t.Errorf("CanEdit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnnouncement_IsVisible(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		status    domain.AnnouncementStatus
		publishAt *time.Time
		expireAt  *time.Time
		want      bool
	}{
		{
			name:      "published without dates",
			status:    domain.AnnouncementStatusPublished,
			publishAt: nil,
			expireAt:  nil,
			want:      true,
		},
		{
			name:      "draft is not visible",
			status:    domain.AnnouncementStatusDraft,
			publishAt: nil,
			expireAt:  nil,
			want:      false,
		},
		{
			name:      "archived is not visible",
			status:    domain.AnnouncementStatusArchived,
			publishAt: nil,
			expireAt:  nil,
			want:      false,
		},
		{
			name:      "published in past",
			status:    domain.AnnouncementStatusPublished,
			publishAt: &past,
			expireAt:  nil,
			want:      true,
		},
		{
			name:      "scheduled for future",
			status:    domain.AnnouncementStatusPublished,
			publishAt: &future,
			expireAt:  nil,
			want:      false,
		},
		{
			name:      "expired",
			status:    domain.AnnouncementStatusPublished,
			publishAt: &past,
			expireAt:  &past,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnnouncement("Test", "Content", 1)
			a.Status = tt.status
			a.PublishAt = tt.publishAt
			a.ExpireAt = tt.expireAt

			got := a.IsVisible()
			if got != tt.want {
				t.Errorf("IsVisible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnnouncement_SetPriority(t *testing.T) {
	a := NewAnnouncement("Test", "Content", 1)

	err := a.SetPriority(domain.AnnouncementPriorityHigh)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if a.Priority != domain.AnnouncementPriorityHigh {
		t.Errorf("expected priority %q, got %q", domain.AnnouncementPriorityHigh, a.Priority)
	}

	err = a.SetPriority("invalid")
	if err == nil {
		t.Error("expected error for invalid priority")
	}
}

func TestAnnouncement_SetTargetAudience(t *testing.T) {
	a := NewAnnouncement("Test", "Content", 1)

	err := a.SetTargetAudience(domain.TargetAudienceStudents)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if a.TargetAudience != domain.TargetAudienceStudents {
		t.Errorf("expected audience %q, got %q", domain.TargetAudienceStudents, a.TargetAudience)
	}

	err = a.SetTargetAudience("invalid")
	if err == nil {
		t.Error("expected error for invalid audience")
	}
}
