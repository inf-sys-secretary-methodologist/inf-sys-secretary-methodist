// Package entities содержит доменные сущности модуля files.
package entities

import "time"

// FileMetadata представляет метаданные загруженного файла в базе данных.
type FileMetadata struct {
	ID             int64      `json:"id"`
	OriginalName   string     `json:"original_name"`               // Оригинальное имя файла
	StorageKey     string     `json:"storage_key"`                 // Ключ в S3/MinIO хранилище
	Size           int64      `json:"size"`                        // Размер файла в байтах
	MimeType       string     `json:"mime_type"`                   // MIME тип файла
	Checksum       string     `json:"checksum"`                    // SHA-256 хеш файла
	UploadedBy     int64      `json:"uploaded_by"`                 // ID пользователя, загрузившего файл
	DocumentID     *int64     `json:"document_id,omitempty"`       // Связь с документом (опционально)
	TaskID         *int64     `json:"task_id,omitempty"`           // Связь с задачей (опционально)
	AnnouncementID *int64     `json:"announcement_id,omitempty"`   // Связь с объявлением (опционально)
	IsTemporary    bool       `json:"is_temporary"`                // Временный файл (до прикрепления)
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`        // Срок жизни временного файла
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`        // Мягкое удаление
}

// NewFileMetadata создаёт новый экземпляр FileMetadata.
func NewFileMetadata(originalName, storageKey, mimeType, checksum string, size int64, uploadedBy int64) *FileMetadata {
	now := time.Now()
	return &FileMetadata{
		OriginalName: originalName,
		StorageKey:   storageKey,
		Size:         size,
		MimeType:     mimeType,
		Checksum:     checksum,
		UploadedBy:   uploadedBy,
		IsTemporary:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AttachToDocument привязывает файл к документу.
func (f *FileMetadata) AttachToDocument(documentID int64) {
	f.DocumentID = &documentID
	f.IsTemporary = false
	f.ExpiresAt = nil
	f.UpdatedAt = time.Now()
}

// AttachToTask привязывает файл к задаче.
func (f *FileMetadata) AttachToTask(taskID int64) {
	f.TaskID = &taskID
	f.IsTemporary = false
	f.ExpiresAt = nil
	f.UpdatedAt = time.Now()
}

// AttachToAnnouncement привязывает файл к объявлению.
func (f *FileMetadata) AttachToAnnouncement(announcementID int64) {
	f.AnnouncementID = &announcementID
	f.IsTemporary = false
	f.ExpiresAt = nil
	f.UpdatedAt = time.Now()
}

// MarkAsDeleted выполняет мягкое удаление файла.
func (f *FileMetadata) MarkAsDeleted() {
	now := time.Now()
	f.DeletedAt = &now
	f.UpdatedAt = now
}

// IsDeleted проверяет, удалён ли файл (мягкое удаление).
func (f *FileMetadata) IsDeleted() bool {
	return f.DeletedAt != nil
}

// IsExpired проверяет, истёк ли срок жизни временного файла.
func (f *FileMetadata) IsExpired() bool {
	if !f.IsTemporary || f.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*f.ExpiresAt)
}
