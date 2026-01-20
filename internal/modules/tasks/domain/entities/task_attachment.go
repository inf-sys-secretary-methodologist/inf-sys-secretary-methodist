package entities

import "time"

// TaskAttachment represents a file attachment to a task.
type TaskAttachment struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	MimeType   *string   `json:"mime_type,omitempty"`
	UploadedBy int64     `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
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
