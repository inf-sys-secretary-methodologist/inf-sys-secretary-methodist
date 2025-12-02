// Package entities contains announcement domain entities.
package entities

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
)

var (
	// ErrAnnouncementAlreadyPublished is returned when trying to publish an already published announcement.
	ErrAnnouncementAlreadyPublished = errors.New("announcement is already published")
	// ErrAnnouncementArchived is returned when trying to modify an archived announcement.
	ErrAnnouncementArchived = errors.New("announcement is archived")
	// ErrAnnouncementNotPublished is returned when trying to archive an unpublished announcement.
	ErrAnnouncementNotPublished = errors.New("announcement is not published")
	// ErrInvalidPublishDate is returned when publish date is in the past.
	ErrInvalidPublishDate = errors.New("publish date cannot be in the past")
	// ErrInvalidExpireDate is returned when expire date is before publish date.
	ErrInvalidExpireDate = errors.New("expire date must be after publish date")
)

// Announcement represents a news or announcement entity.
type Announcement struct {
	ID             int64                       `db:"id" json:"id"`
	Title          string                      `db:"title" json:"title"`
	Content        string                      `db:"content" json:"content"`
	Summary        *string                     `db:"summary" json:"summary,omitempty"`
	AuthorID       int64                       `db:"author_id" json:"author_id"`
	Status         domain.AnnouncementStatus   `db:"status" json:"status"`
	Priority       domain.AnnouncementPriority `db:"priority" json:"priority"`
	TargetAudience domain.TargetAudience       `db:"target_audience" json:"target_audience"`
	PublishAt      *time.Time                  `db:"publish_at" json:"publish_at,omitempty"`
	ExpireAt       *time.Time                  `db:"expire_at" json:"expire_at,omitempty"`
	IsPinned       bool                        `db:"is_pinned" json:"is_pinned"`
	ViewCount      int64                       `db:"view_count" json:"view_count"`
	Tags           []string                    `db:"tags" json:"tags,omitempty"`
	Metadata       json.RawMessage             `db:"metadata" json:"metadata,omitempty"`
	CreatedAt      time.Time                   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time                   `db:"updated_at" json:"updated_at"`

	// Associations (not stored in DB, loaded separately)
	Author      *AnnouncementAuthor      `db:"-" json:"author,omitempty"`
	Attachments []AnnouncementAttachment `db:"-" json:"attachments,omitempty"`
}

// AnnouncementAuthor represents the author of an announcement.
type AnnouncementAuthor struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AnnouncementAttachment represents a file attached to an announcement.
type AnnouncementAttachment struct {
	ID             int64     `db:"id" json:"id"`
	AnnouncementID int64     `db:"announcement_id" json:"announcement_id"`
	FileName       string    `db:"file_name" json:"file_name"`
	FilePath       string    `db:"file_path" json:"file_path"`
	FileSize       int64     `db:"file_size" json:"file_size"`
	MimeType       string    `db:"mime_type" json:"mime_type"`
	UploadedBy     int64     `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// NewAnnouncement creates a new announcement with default values.
func NewAnnouncement(title, content string, authorID int64) *Announcement {
	now := time.Now()
	return &Announcement{
		Title:          title,
		Content:        content,
		AuthorID:       authorID,
		Status:         domain.AnnouncementStatusDraft,
		Priority:       domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll,
		IsPinned:       false,
		ViewCount:      0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Publish publishes the announcement.
func (a *Announcement) Publish() error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}
	if a.Status == domain.AnnouncementStatusPublished {
		return ErrAnnouncementAlreadyPublished
	}

	now := time.Now()
	a.Status = domain.AnnouncementStatusPublished
	if a.PublishAt == nil {
		a.PublishAt = &now
	}
	a.UpdatedAt = now
	return nil
}

// Archive archives the announcement.
func (a *Announcement) Archive() error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}
	if a.Status != domain.AnnouncementStatusPublished {
		return ErrAnnouncementNotPublished
	}

	a.Status = domain.AnnouncementStatusArchived
	a.UpdatedAt = time.Now()
	return nil
}

// Unpublish moves the announcement back to draft status.
func (a *Announcement) Unpublish() error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}
	if a.Status != domain.AnnouncementStatusPublished {
		return ErrAnnouncementNotPublished
	}

	a.Status = domain.AnnouncementStatusDraft
	a.UpdatedAt = time.Now()
	return nil
}

// SetPriority sets the priority of the announcement.
func (a *Announcement) SetPriority(priority domain.AnnouncementPriority) error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}
	if !priority.IsValid() {
		return errors.New("invalid priority")
	}

	a.Priority = priority
	a.UpdatedAt = time.Now()
	return nil
}

// SetTargetAudience sets the target audience of the announcement.
func (a *Announcement) SetTargetAudience(audience domain.TargetAudience) error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}
	if !audience.IsValid() {
		return errors.New("invalid target audience")
	}

	a.TargetAudience = audience
	a.UpdatedAt = time.Now()
	return nil
}

// SetPublishSchedule sets the publish and expire dates.
func (a *Announcement) SetPublishSchedule(publishAt, expireAt *time.Time) error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}

	now := time.Now()
	if publishAt != nil && publishAt.Before(now) {
		return ErrInvalidPublishDate
	}
	if publishAt != nil && expireAt != nil && expireAt.Before(*publishAt) {
		return ErrInvalidExpireDate
	}

	a.PublishAt = publishAt
	a.ExpireAt = expireAt
	a.UpdatedAt = now
	return nil
}

// Pin pins the announcement to the top.
func (a *Announcement) Pin() error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}

	a.IsPinned = true
	a.UpdatedAt = time.Now()
	return nil
}

// Unpin removes the announcement from pinned.
func (a *Announcement) Unpin() error {
	if a.Status == domain.AnnouncementStatusArchived {
		return ErrAnnouncementArchived
	}

	a.IsPinned = false
	a.UpdatedAt = time.Now()
	return nil
}

// IncrementViewCount increments the view counter.
func (a *Announcement) IncrementViewCount() {
	a.ViewCount++
}

// IsExpired checks if the announcement has expired.
func (a *Announcement) IsExpired() bool {
	if a.ExpireAt == nil {
		return false
	}
	return time.Now().After(*a.ExpireAt)
}

// IsVisible checks if the announcement is visible to users.
func (a *Announcement) IsVisible() bool {
	if a.Status != domain.AnnouncementStatusPublished {
		return false
	}
	if a.IsExpired() {
		return false
	}
	if a.PublishAt != nil && time.Now().Before(*a.PublishAt) {
		return false
	}
	return true
}

// CanEdit checks if the user can edit this announcement.
func (a *Announcement) CanEdit(userID int64, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	return a.AuthorID == userID && a.Status != domain.AnnouncementStatusArchived
}
