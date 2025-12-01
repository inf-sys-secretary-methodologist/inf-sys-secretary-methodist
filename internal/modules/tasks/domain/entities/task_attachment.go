package entities

import "time"

// TaskAttachment represents a file attachment to a task.
type TaskAttachment struct {
	ID         int64     `db:"id" json:"id"`
	TaskID     int64     `db:"task_id" json:"task_id"`
	FileName   string    `db:"file_name" json:"file_name"`
	FilePath   string    `db:"file_path" json:"file_path"`
	FileSize   int64     `db:"file_size" json:"file_size"`
	MimeType   *string   `db:"mime_type" json:"mime_type,omitempty"`
	UploadedBy int64     `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// NewTaskAttachment creates a new task attachment.
func NewTaskAttachment(taskID int64, fileName, filePath string, fileSize int64, uploadedBy int64) *TaskAttachment {
	return &TaskAttachment{
		TaskID:     taskID,
		FileName:   fileName,
		FilePath:   filePath,
		FileSize:   fileSize,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}
}

// SetMimeType sets the MIME type of the attachment.
func (a *TaskAttachment) SetMimeType(mimeType string) {
	a.MimeType = &mimeType
}
