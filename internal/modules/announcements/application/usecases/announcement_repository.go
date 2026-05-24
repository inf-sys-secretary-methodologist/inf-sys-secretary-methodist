// Package usecases contains announcement business logic.
package usecases

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
	// GetByID returns the announcement without audience filtering. Used by
	// admin/author paths (Update / Delete / Publish / Archive) that need
	// full access regardless of caller role. Public-facing reads must use
	// GetByIDForAudience instead.
	GetByID(ctx context.Context, id int64) (*entities.Announcement, error)
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter AnnouncementFilter, limit, offset int) ([]*entities.Announcement, error)
	Count(ctx context.Context, filter AnnouncementFilter) (int64, error)
	GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Announcement, error)
	GetPublished(ctx context.Context, audience domain.TargetAudience, limit, offset int) ([]*entities.Announcement, error)

	// Audience-filtered public reads — v0.163.1 ADR-2 polish
	// (defense-in-depth поверх handler-layer clamp from v0.163.0).
	// All three apply `target_audience = ANY($N)` к the SQL so a usecase
	// caller that forgot к clamp at the handler boundary can't leak
	// announcements addressed к audiences the role can't see.

	// GetByIDForAudience returns the announcement only if its
	// target_audience falls within the provided list. Used by the
	// public read endpoint GET /api/announcements/:id.
	GetByIDForAudience(ctx context.Context, id int64, audiences []domain.TargetAudience) (*entities.Announcement, error)
	GetPinned(ctx context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error)
	GetRecent(ctx context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error)

	// View tracking
	IncrementViewCount(ctx context.Context, id int64) error

	// Attachments
	AddAttachment(ctx context.Context, attachment *entities.AnnouncementAttachment) error
	RemoveAttachment(ctx context.Context, attachmentID int64) error
	GetAttachments(ctx context.Context, announcementID int64) ([]*entities.AnnouncementAttachment, error)
	GetAttachmentByID(ctx context.Context, attachmentID int64) (*entities.AnnouncementAttachment, error)
}
