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
)

var (
	// ErrAnnouncementNotFound is returned when an announcement is not found.
	ErrAnnouncementNotFound = errors.New("announcement not found")
	// ErrUnauthorized is returned when user is not authorized for the operation.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")
)

// UserIDsProvider returns the active user IDs whose role makes them
// recipients for an announcement targeted к the given audience. The
// announcement broadcast fan-out scopes notifications к the audience —
// a student-targeted announcement only pushes к students, etc. — so
// the v0.163.0 Tier 1 cross-audience push leak doesn't reappear when
// the fan-out is wired (main.go).
type UserIDsProvider interface {
	GetUserIDsForAudience(ctx context.Context, audience domain.TargetAudience) ([]int64, error)
}

// SystemNotifier is the narrow port for sending a single system-
// originated notification к one user. Replaces the prior dependency
// on the concrete `*notifUsecases.NotificationUseCase` — the announcement
// usecase only ever calls SendSystemNotification, so the wide
// interface was leaking unused surface. The real
// notifications.NotificationUseCase satisfies it structurally; tests
// substitute a recording fake.
type SystemNotifier interface {
	SendSystemNotification(ctx context.Context, userID int64, title, summary string) error
}

// AnnouncementUseCase handles announcement business logic.
type AnnouncementUseCase struct {
	repo              AnnouncementRepository
	audit             AuditSink
	notifier          SystemNotifier
	userIDsProvider   UserIDsProvider
	attachmentStorage AttachmentStorage // optional; wired via SetAttachmentStorage
	// lifecycleCtx replaces context.Background() in the broadcast
	// fan-out goroutine. main.go passes a server-lifecycle ctx through
	// WithLifecycleContext so graceful shutdown cancels in-flight
	// notifications instead of leaking goroutines past server stop.
	lifecycleCtx context.Context
}

// NewAnnouncementUseCase creates a new AnnouncementUseCase. The audit
// arg is the narrow AuditSink port — pass *logging.AuditLogger in
// production, a recording fake in tests. Nil sink is treated as a
// silent no-op at every call site.
func NewAnnouncementUseCase(
	repo AnnouncementRepository,
	audit AuditSink,
	notifier SystemNotifier,
	userIDsProvider UserIDsProvider,
) *AnnouncementUseCase {
	return &AnnouncementUseCase{
		repo:            repo,
		audit:           audit,
		notifier:        notifier,
		userIDsProvider: userIDsProvider,
		lifecycleCtx:    context.Background(),
	}
}

// WithLifecycleContext registers the server-lifecycle ctx the broadcast
// fan-out should use instead of context.Background(). Chainable —
// returns the receiver. Pattern matches the optional-deps setter
// shape used elsewhere in the codebase (e.g. MFA verification wiring
// in auth usecase).
func (uc *AnnouncementUseCase) WithLifecycleContext(ctx context.Context) *AnnouncementUseCase {
	if ctx != nil {
		uc.lifecycleCtx = ctx
	}
	return uc
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

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "announcement.created", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"title":           announcement.Title,
			"author_id":       userID,
		})
	}

	return announcement, nil
}

// GetByID retrieves an announcement by ID, filtered к the caller's
// audience set. The handler computes `audiences` via
// domain.VisibleAudiences(role) — this method is the public read path,
// so the repo refuses anything outside the audience list (v0.163.1
// ADR-2 polish, defense-in-depth поверх handler-layer clamp from
// v0.163.0). Admin/author paths (Update / Delete / Publish / Archive)
// keep using uc.repo.GetByID directly without the audience filter.
func (uc *AnnouncementUseCase) GetByID(ctx context.Context, id int64, incrementView bool, audiences []domain.TargetAudience) (*entities.Announcement, error) {
	announcement, err := uc.repo.GetByIDForAudience(ctx, id, audiences)
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

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "announcement.updated", "announcement", map[string]interface{}{
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

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "announcement.deleted", "announcement", map[string]interface{}{
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

	filter := AnnouncementFilter{
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

// GetPinned retrieves pinned announcements visible к the caller's
// audience set. Handler passes audiences derived via
// domain.VisibleAudiences(role); the repo enforces SQL
// `target_audience = ANY($1)`.
func (uc *AnnouncementUseCase) GetPinned(ctx context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	return uc.repo.GetPinned(ctx, audiences, limit)
}

// GetRecent retrieves recent announcements visible к the caller's
// audience set. Same audience contract as GetPinned.
func (uc *AnnouncementUseCase) GetRecent(ctx context.Context, audiences []domain.TargetAudience, limit int) ([]*entities.Announcement, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	return uc.repo.GetRecent(ctx, audiences, limit)
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

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "announcement.published", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"published_by":    userID,
		})
	}

	// Broadcast push notification к users whose role matches the
	// announcement's target_audience. v0.163.1 polish: audience-scoped
	// (no cross-audience leak, closes v0.163.0 audit T1-7) and runs on
	// the server-lifecycle ctx so graceful shutdown can cancel
	// in-flight sends instead of orphaning the goroutine.
	if uc.notifier != nil && uc.userIDsProvider != nil {
		audience := announcement.TargetAudience
		go func() { // #nosec G118 -- fire-and-forget goroutine; cancellable via uc.lifecycleCtx
			userIDs, err := uc.userIDsProvider.GetUserIDsForAudience(uc.lifecycleCtx, audience)
			if err != nil {
				return
			}
			for _, uid := range userIDs {
				summary := ""
				if announcement.Summary != nil {
					summary = *announcement.Summary
				}
				_ = uc.notifier.SendSystemNotification(
					uc.lifecycleCtx,
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

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "announcement.unpublished", "announcement", map[string]interface{}{
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

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "announcement.archived", "announcement", map[string]interface{}{
			"announcement_id": announcement.ID,
			"archived_by":     userID,
		})
	}

	return announcement, nil
}
