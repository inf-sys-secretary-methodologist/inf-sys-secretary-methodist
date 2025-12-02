// Package repositories defines announcement repository interfaces.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
)

// AnnouncementFilter defines filtering options for announcements.
type AnnouncementFilter struct {
	AuthorID       *int64
	Status         *domain.AnnouncementStatus
	Priority       *domain.AnnouncementPriority
	TargetAudience *domain.TargetAudience
	IsPinned       *bool
	IsExpired      *bool
	Search         *string
	Tags           []string
}

// AnnouncementRepository defines the interface for announcement persistence.
type AnnouncementRepository interface {
	// CRUD operations
	Create(ctx context.Context, announcement *entities.Announcement) error
	Save(ctx context.Context, announcement *entities.Announcement) error
	GetByID(ctx context.Context, id int64) (*entities.Announcement, error)
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter AnnouncementFilter, limit, offset int) ([]*entities.Announcement, error)
	Count(ctx context.Context, filter AnnouncementFilter) (int64, error)
	GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Announcement, error)
	GetPublished(ctx context.Context, audience domain.TargetAudience, limit, offset int) ([]*entities.Announcement, error)
	GetPinned(ctx context.Context, limit int) ([]*entities.Announcement, error)
	GetRecent(ctx context.Context, limit int) ([]*entities.Announcement, error)

	// View tracking
	IncrementViewCount(ctx context.Context, id int64) error

	// Attachments
	AddAttachment(ctx context.Context, attachment *entities.AnnouncementAttachment) error
	RemoveAttachment(ctx context.Context, attachmentID int64) error
	GetAttachments(ctx context.Context, announcementID int64) ([]*entities.AnnouncementAttachment, error)
	GetAttachmentByID(ctx context.Context, attachmentID int64) (*entities.AnnouncementAttachment, error)
}
