package usecases

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
)

// SetAttachmentStorage wires an AttachmentStorage implementation into the
// usecase. Called once during DI setup in main.go. Kept as a setter (not in
// the constructor) so the existing NewAnnouncementUseCase signature stays
// stable for callers that don't need attachments.
func (uc *AnnouncementUseCase) SetAttachmentStorage(s AttachmentStorage) {
	uc.attachmentStorage = s
}

// AddAttachment stores a file in object storage and links it to an
// announcement.
//
// Order of operations is deliberate:
//  1. Verify the announcement exists.
//  2. Upload to storage (so we never persist orphan rows).
//  3. Persist the metadata row in the repository.
//  4. If the repo write fails, delete the just-uploaded blob to avoid leaks.
func (uc *AnnouncementUseCase) AddAttachment(
	ctx context.Context,
	announcementID int64,
	fileName string,
	reader io.Reader,
	size int64,
	mimeType string,
	uploadedBy int64,
) (*entities.AnnouncementAttachment, error) {
	if uc.attachmentStorage == nil {
		return nil, ErrStorageNotConfigured
	}

	// 1. Announcement must exist.
	ann, err := uc.repo.GetByID(ctx, announcementID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup announcement: %w", err)
	}
	if ann == nil {
		return nil, ErrAnnouncementNotFound
	}

	// 2. Upload to storage. Key includes a UUID so concurrent uploads of
	// identically-named files don't collide.
	key := path.Join(
		"announcements",
		fmt.Sprintf("%d", announcementID),
		uuid.NewString()+"-"+fileName,
	)
	if _, err := uc.attachmentStorage.Upload(ctx, key, reader, size, mimeType); err != nil {
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}

	// 3. Persist metadata.
	att := &entities.AnnouncementAttachment{
		AnnouncementID: announcementID,
		FileName:       fileName,
		FilePath:       key,
		FileSize:       size,
		MimeType:       mimeType,
		UploadedBy:     uploadedBy,
	}
	if err := uc.repo.AddAttachment(ctx, att); err != nil {
		// 4. Best-effort rollback to avoid orphan blob.
		_ = uc.attachmentStorage.Delete(ctx, key)
		return nil, fmt.Errorf("failed to persist attachment: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.attachment.added", "announcement_attachment", map[string]interface{}{
			"announcement_id": announcementID,
			"attachment_id":   att.ID,
			"file_name":       fileName,
			"file_size":       size,
			"uploaded_by":     uploadedBy,
		})
	}

	return att, nil
}

// RemoveAttachment removes the storage blob and the metadata row.
func (uc *AnnouncementUseCase) RemoveAttachment(ctx context.Context, attachmentID int64) error {
	if uc.attachmentStorage == nil {
		return ErrStorageNotConfigured
	}

	att, err := uc.repo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to lookup attachment: %w", err)
	}
	if att == nil {
		return ErrAttachmentNotFound
	}

	if err := uc.attachmentStorage.Delete(ctx, att.FilePath); err != nil {
		return fmt.Errorf("failed to delete blob from storage: %w", err)
	}

	if err := uc.repo.RemoveAttachment(ctx, attachmentID); err != nil {
		return fmt.Errorf("failed to remove attachment row: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "announcement.attachment.removed", "announcement_attachment", map[string]interface{}{
			"announcement_id": att.AnnouncementID,
			"attachment_id":   attachmentID,
		})
	}

	return nil
}
