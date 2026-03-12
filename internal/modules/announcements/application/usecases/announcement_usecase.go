// Package usecases contains announcement business logic.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/repositories"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

var (
	// ErrAnnouncementNotFound is returned when an announcement is not found.
	ErrAnnouncementNotFound = errors.New("announcement not found")
	// ErrUnauthorized is returned when user is not authorized for the operation.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")
)

// UserIDsProvider provides a list of user IDs for broadcast notifications.
type UserIDsProvider interface {
	GetActiveUserIDs(ctx context.Context) ([]int64, error)
}

// AnnouncementUseCase handles announcement business logic.
type AnnouncementUseCase struct {
	repo                repositories.AnnouncementRepository
	auditLogger         *logging.AuditLogger
	notificationUseCase *notifUsecases.NotificationUseCase
	userIDsProvider     UserIDsProvider
}

// NewAnnouncementUseCase creates a new AnnouncementUseCase.
func NewAnnouncementUseCase(
	repo repositories.AnnouncementRepository,
	auditLogger *logging.AuditLogger,
	notificationUseCase *notifUsecases.NotificationUseCase,
	userIDsProvider UserIDsProvider,
) *AnnouncementUseCase {
	return &AnnouncementUseCase{
		repo:                repo,
		auditLogger:         auditLogger,
		notificationUseCase: notificationUseCase,
		userIDsProvider:     userIDsProvider,
	}
}

// Create creates a new announcement.
func (uc *AnnouncementUseCase) Create(ctx context.Context, userID int64, req *dto.CreateAnnouncementRequest) (*entities.Announcement, error) {
	if !req.Priority.IsValid() {
		return nil, ErrInvalidInput
	}
	if !req.TargetAudience.IsValid() {
		return nil, ErrInvalidInput
	}

	announcement := entities.NewAnnouncement(req.Title, req.Content, userID)
	announcement.Summary = req.Summary
	announcement.Priority = req.Priority
	announcement.TargetAudience = req.TargetAudience
	announcement.PublishAt = req.PublishAt
	announcement.ExpireAt = req.ExpireAt
	announcement.IsPinned = req.IsPinned
	announcement.Tags = req.Tags

	if err := uc.repo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.created", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"title":           announcement.Title,
			"author_id":       userID,
		})
	}

	return announcement, nil
}

// GetByID retrieves an announcement by ID.
func (uc *AnnouncementUseCase) GetByID(ctx context.Context, id int64, incrementView bool) (*entities.Announcement, error) {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	// Load attachments
	attachments, err := uc.repo.GetAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	announcement.Attachments = make([]entities.AnnouncementAttachment, len(attachments))
	for i, att := range attachments {
		announcement.Attachments[i] = *att
	}

	// Increment view count if requested
	if incrementView && announcement.IsVisible() {
		_ = uc.repo.IncrementViewCount(ctx, id)
		announcement.ViewCount++
	}

	return announcement, nil
}

// Update updates an announcement.
func (uc *AnnouncementUseCase) Update(ctx context.Context, userID int64, id int64, isAdmin bool, req *dto.UpdateAnnouncementRequest) (*entities.Announcement, error) {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	if !announcement.CanEdit(userID, isAdmin) {
		return nil, ErrUnauthorized
	}

	if req.Title != nil {
		announcement.Title = *req.Title
	}
	if req.Content != nil {
		announcement.Content = *req.Content
	}
	if req.Summary != nil {
		announcement.Summary = req.Summary
	}
	if req.Priority != nil {
		if err := announcement.SetPriority(*req.Priority); err != nil {
			return nil, err
		}
	}
	if req.TargetAudience != nil {
		if err := announcement.SetTargetAudience(*req.TargetAudience); err != nil {
			return nil, err
		}
	}
	if req.PublishAt != nil || req.ExpireAt != nil {
		publishAt := announcement.PublishAt
		expireAt := announcement.ExpireAt
		if req.PublishAt != nil {
			publishAt = req.PublishAt
		}
		if req.ExpireAt != nil {
			expireAt = req.ExpireAt
		}
		if err := announcement.SetPublishSchedule(publishAt, expireAt); err != nil {
			return nil, err
		}
	}
	if req.IsPinned != nil {
		if *req.IsPinned {
			_ = announcement.Pin()
		} else {
			_ = announcement.Unpin()
		}
	}
	if req.Tags != nil {
		announcement.Tags = req.Tags
	}

	announcement.UpdatedAt = time.Now()

	if err := uc.repo.Save(ctx, announcement); err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.updated", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"updated_by":      userID,
		})
	}

	return announcement, nil
}

// Delete deletes an announcement.
func (uc *AnnouncementUseCase) Delete(ctx context.Context, userID int64, id int64, isAdmin bool) error {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if announcement == nil {
		return ErrAnnouncementNotFound
	}

	if !announcement.CanEdit(userID, isAdmin) {
		return ErrUnauthorized
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.deleted", "announcement", map[string]interface{}{
			"announcement_id": id,
			"deleted_by":      userID,
		})
	}

	return nil
}

// List lists announcements with filters.
func (uc *AnnouncementUseCase) List(ctx context.Context, req *dto.ListAnnouncementsRequest) (*dto.ListAnnouncementsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	filter := repositories.AnnouncementFilter{
		AuthorID:       req.AuthorID,
		Status:         req.Status,
		Priority:       req.Priority,
		TargetAudience: req.TargetAudience,
		IsPinned:       req.IsPinned,
		Search:         req.Search,
		Tags:           req.Tags,
	}

	announcements, err := uc.repo.List(ctx, filter, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &dto.ListAnnouncementsResponse{
		Announcements: dto.ToResponseList(announcements),
		Total:         total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}, nil
}

// GetPublished retrieves published announcements for a specific audience.
func (uc *AnnouncementUseCase) GetPublished(ctx context.Context, audience domain.TargetAudience, limit, offset int) ([]*entities.Announcement, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return uc.repo.GetPublished(ctx, audience, limit, offset)
}

// GetPinned retrieves pinned announcements.
func (uc *AnnouncementUseCase) GetPinned(ctx context.Context, limit int) ([]*entities.Announcement, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	return uc.repo.GetPinned(ctx, limit)
}

// GetRecent retrieves recent announcements.
func (uc *AnnouncementUseCase) GetRecent(ctx context.Context, limit int) ([]*entities.Announcement, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	return uc.repo.GetRecent(ctx, limit)
}

// Publish publishes an announcement.
func (uc *AnnouncementUseCase) Publish(ctx context.Context, userID int64, id int64, isAdmin bool) (*entities.Announcement, error) {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	if !announcement.CanEdit(userID, isAdmin) {
		return nil, ErrUnauthorized
	}

	if err := announcement.Publish(); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, announcement); err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.published", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"published_by":    userID,
		})
	}

	// Send broadcast notification to all users
	if uc.notificationUseCase != nil && uc.userIDsProvider != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			userIDs, err := uc.userIDsProvider.GetActiveUserIDs(context.Background())
			if err != nil {
				return
			}
			for _, uid := range userIDs {
				summary := ""
				if announcement.Summary != nil {
					summary = *announcement.Summary
				}
				_ = uc.notificationUseCase.SendSystemNotification(
					context.Background(),
					uid,
					fmt.Sprintf("Объявление: %s", announcement.Title),
					summary,
				)
			}
		}()
	}

	return announcement, nil
}

// Unpublish moves an announcement back to draft.
func (uc *AnnouncementUseCase) Unpublish(ctx context.Context, userID int64, id int64, isAdmin bool) (*entities.Announcement, error) {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	if !announcement.CanEdit(userID, isAdmin) {
		return nil, ErrUnauthorized
	}

	if err := announcement.Unpublish(); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, announcement); err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.unpublished", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"unpublished_by":  userID,
		})
	}

	return announcement, nil
}

// Archive archives an announcement.
func (uc *AnnouncementUseCase) Archive(ctx context.Context, userID int64, id int64, isAdmin bool) (*entities.Announcement, error) {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	if !isAdmin && announcement.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	if err := announcement.Archive(); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, announcement); err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.archived", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"archived_by":     userID,
		})
	}

	return announcement, nil
}
