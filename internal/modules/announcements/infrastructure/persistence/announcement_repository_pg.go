// Package persistence provides PostgreSQL repository implementations for the announcements module.
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/repositories"
)

// AnnouncementRepositoryPG implements AnnouncementRepository using PostgreSQL.
type AnnouncementRepositoryPG struct {
	db *sql.DB
}

// NewAnnouncementRepositoryPG creates a new AnnouncementRepositoryPG.
func NewAnnouncementRepositoryPG(db *sql.DB) *AnnouncementRepositoryPG {
	return &AnnouncementRepositoryPG{db: db}
}

// Create creates a new announcement.
func (r *AnnouncementRepositoryPG) Create(ctx context.Context, announcement *entities.Announcement) error {
	query := `
		INSERT INTO announcements (
			title, content, summary, author_id, status, priority,
			target_audience, publish_at, expire_at, is_pinned, view_count,
			tags, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		announcement.Title, announcement.Content, announcement.Summary,
		announcement.AuthorID, announcement.Status, announcement.Priority,
		announcement.TargetAudience, announcement.PublishAt, announcement.ExpireAt,
		announcement.IsPinned, announcement.ViewCount,
		pq.Array(announcement.Tags), announcement.Metadata,
		announcement.CreatedAt, announcement.UpdatedAt,
	).Scan(&announcement.ID)
}

// Save updates an existing announcement.
func (r *AnnouncementRepositoryPG) Save(ctx context.Context, announcement *entities.Announcement) error {
	query := `
		UPDATE announcements SET
			title = $1, content = $2, summary = $3, status = $4, priority = $5,
			target_audience = $6, publish_at = $7, expire_at = $8, is_pinned = $9,
			tags = $10, metadata = $11, updated_at = $12
		WHERE id = $13`

	_, err := r.db.ExecContext(ctx, query,
		announcement.Title, announcement.Content, announcement.Summary,
		announcement.Status, announcement.Priority, announcement.TargetAudience,
		announcement.PublishAt, announcement.ExpireAt, announcement.IsPinned,
		pq.Array(announcement.Tags), announcement.Metadata,
		announcement.UpdatedAt, announcement.ID,
	)
	return err
}

// GetByID retrieves an announcement by ID.
func (r *AnnouncementRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Announcement, error) {
	query := `
		SELECT id, title, content, summary, author_id, status, priority,
			target_audience, publish_at, expire_at, is_pinned, view_count,
			tags, metadata, created_at, updated_at
		FROM announcements WHERE id = $1`

	announcement := &entities.Announcement{}
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&announcement.ID, &announcement.Title, &announcement.Content,
		&announcement.Summary, &announcement.AuthorID, &announcement.Status,
		&announcement.Priority, &announcement.TargetAudience,
		&announcement.PublishAt, &announcement.ExpireAt, &announcement.IsPinned,
		&announcement.ViewCount, &tags, &announcement.Metadata,
		&announcement.CreatedAt, &announcement.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	announcement.Tags = tags
	return announcement, nil
}

// Delete deletes an announcement.
func (r *AnnouncementRepositoryPG) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM announcements WHERE id = $1", id)
	return err
}

// List lists announcements with filters.
func (r *AnnouncementRepositoryPG) List(ctx context.Context, filter repositories.AnnouncementFilter, limit, offset int) ([]*entities.Announcement, error) {
	query, args := r.buildListQuery(filter, limit, offset, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanAnnouncements(rows)
}

// Count counts announcements with filters.
func (r *AnnouncementRepositoryPG) Count(ctx context.Context, filter repositories.AnnouncementFilter) (int64, error) {
	query, args := r.buildListQuery(filter, 0, 0, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// GetByAuthor retrieves announcements by author.
func (r *AnnouncementRepositoryPG) GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Announcement, error) {
	filter := repositories.AnnouncementFilter{AuthorID: &authorID}
	return r.List(ctx, filter, limit, offset)
}

// GetPublished retrieves published announcements for a specific audience.
func (r *AnnouncementRepositoryPG) GetPublished(ctx context.Context, audience domain.TargetAudience, limit, offset int) ([]*entities.Announcement, error) {
	status := domain.AnnouncementStatusPublished
	filter := repositories.AnnouncementFilter{
		Status:         &status,
		TargetAudience: &audience,
	}

	query := `
		SELECT id, title, content, summary, author_id, status, priority,
			target_audience, publish_at, expire_at, is_pinned, view_count,
			tags, metadata, created_at, updated_at
		FROM announcements
		WHERE status = $1
			AND (target_audience = $2 OR target_audience = 'all')
			AND (publish_at IS NULL OR publish_at <= NOW())
			AND (expire_at IS NULL OR expire_at > NOW())
		ORDER BY is_pinned DESC, publish_at DESC NULLS LAST, created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, filter.Status, audience, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanAnnouncements(rows)
}

// GetPinned retrieves pinned announcements.
func (r *AnnouncementRepositoryPG) GetPinned(ctx context.Context, limit int) ([]*entities.Announcement, error) {
	query := `
		SELECT id, title, content, summary, author_id, status, priority,
			target_audience, publish_at, expire_at, is_pinned, view_count,
			tags, metadata, created_at, updated_at
		FROM announcements
		WHERE is_pinned = true
			AND status = 'published'
			AND (publish_at IS NULL OR publish_at <= NOW())
			AND (expire_at IS NULL OR expire_at > NOW())
		ORDER BY priority DESC, created_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanAnnouncements(rows)
}

// GetRecent retrieves recent announcements.
func (r *AnnouncementRepositoryPG) GetRecent(ctx context.Context, limit int) ([]*entities.Announcement, error) {
	query := `
		SELECT id, title, content, summary, author_id, status, priority,
			target_audience, publish_at, expire_at, is_pinned, view_count,
			tags, metadata, created_at, updated_at
		FROM announcements
		WHERE status = 'published'
			AND (publish_at IS NULL OR publish_at <= NOW())
			AND (expire_at IS NULL OR expire_at > NOW())
		ORDER BY publish_at DESC NULLS LAST, created_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanAnnouncements(rows)
}

// IncrementViewCount increments the view counter for an announcement.
func (r *AnnouncementRepositoryPG) IncrementViewCount(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE announcements SET view_count = view_count + 1 WHERE id = $1", id)
	return err
}

// AddAttachment adds an attachment to an announcement.
func (r *AnnouncementRepositoryPG) AddAttachment(ctx context.Context, attachment *entities.AnnouncementAttachment) error {
	query := `
		INSERT INTO announcement_attachments (
			announcement_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		attachment.AnnouncementID, attachment.FileName, attachment.FilePath,
		attachment.FileSize, attachment.MimeType, attachment.UploadedBy,
		attachment.CreatedAt,
	).Scan(&attachment.ID)
}

// RemoveAttachment removes an attachment.
func (r *AnnouncementRepositoryPG) RemoveAttachment(ctx context.Context, attachmentID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM announcement_attachments WHERE id = $1", attachmentID)
	return err
}

// GetAttachments retrieves all attachments for an announcement.
func (r *AnnouncementRepositoryPG) GetAttachments(ctx context.Context, announcementID int64) ([]*entities.AnnouncementAttachment, error) {
	query := `
		SELECT id, announcement_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM announcement_attachments
		WHERE announcement_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, announcementID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var attachments []*entities.AnnouncementAttachment
	for rows.Next() {
		att := &entities.AnnouncementAttachment{}
		if err := rows.Scan(
			&att.ID, &att.AnnouncementID, &att.FileName, &att.FilePath,
			&att.FileSize, &att.MimeType, &att.UploadedBy, &att.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}

	return attachments, rows.Err()
}

// GetAttachmentByID retrieves an attachment by ID.
func (r *AnnouncementRepositoryPG) GetAttachmentByID(ctx context.Context, attachmentID int64) (*entities.AnnouncementAttachment, error) {
	query := `
		SELECT id, announcement_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM announcement_attachments
		WHERE id = $1`

	att := &entities.AnnouncementAttachment{}
	err := r.db.QueryRowContext(ctx, query, attachmentID).Scan(
		&att.ID, &att.AnnouncementID, &att.FileName, &att.FilePath,
		&att.FileSize, &att.MimeType, &att.UploadedBy, &att.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return att, nil
}

func (r *AnnouncementRepositoryPG) buildListQuery(filter repositories.AnnouncementFilter, limit, offset int, countOnly bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argNum))
		args = append(args, *filter.AuthorID)
		argNum++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argNum))
		args = append(args, *filter.Priority)
		argNum++
	}

	if filter.TargetAudience != nil {
		conditions = append(conditions, fmt.Sprintf("(target_audience = $%d OR target_audience = 'all')", argNum))
		args = append(args, *filter.TargetAudience)
		argNum++
	}

	if filter.IsPinned != nil {
		conditions = append(conditions, fmt.Sprintf("is_pinned = $%d", argNum))
		args = append(args, *filter.IsPinned)
		argNum++
	}

	if filter.IsExpired != nil {
		if *filter.IsExpired {
			conditions = append(conditions, "expire_at IS NOT NULL AND expire_at < NOW()")
		} else {
			conditions = append(conditions, "(expire_at IS NULL OR expire_at >= NOW())")
		}
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d)", argNum, argNum+1))
		searchPattern := "%" + *filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argNum += 2
	}

	if len(filter.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argNum))
		args = append(args, pq.Array(filter.Tags))
		argNum++
	}

	var query string
	if countOnly {
		query = "SELECT COUNT(*) FROM announcements"
	} else {
		query = `
			SELECT id, title, content, summary, author_id, status, priority,
				target_audience, publish_at, expire_at, is_pinned, view_count,
				tags, metadata, created_at, updated_at
			FROM announcements`
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if !countOnly {
		query += " ORDER BY is_pinned DESC, publish_at DESC NULLS LAST, created_at DESC"

		if limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
			args = append(args, limit, offset)
		}
	}

	return query, args
}

func (r *AnnouncementRepositoryPG) scanAnnouncements(rows *sql.Rows) ([]*entities.Announcement, error) {
	var announcements []*entities.Announcement

	for rows.Next() {
		announcement := &entities.Announcement{}
		var tags pq.StringArray

		if err := rows.Scan(
			&announcement.ID, &announcement.Title, &announcement.Content,
			&announcement.Summary, &announcement.AuthorID, &announcement.Status,
			&announcement.Priority, &announcement.TargetAudience,
			&announcement.PublishAt, &announcement.ExpireAt, &announcement.IsPinned,
			&announcement.ViewCount, &tags, &announcement.Metadata,
			&announcement.CreatedAt, &announcement.UpdatedAt,
		); err != nil {
			return nil, err
		}

		announcement.Tags = tags
		announcements = append(announcements, announcement)
	}

	return announcements, rows.Err()
}
