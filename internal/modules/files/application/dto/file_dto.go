// Package dto содержит объекты передачи данных модуля files.
package dto

import "time"

// UploadFileInput представляет входные данные для загрузки файла.
type UploadFileInput struct {
	OriginalName string `json:"original_name" validate:"required,min=1,max=500"`
	MimeType     string `json:"mime_type" validate:"required"`
	Size         int64  `json:"size" validate:"required,min=1"`
	UserID       int64  `json:"-"` // Заполняется из контекста авторизации
}

// AttachFileInput представляет входные данные для прикрепления файла.
type AttachFileInput struct {
	FileID         int64  `json:"file_id" validate:"required"`
	DocumentID     *int64 `json:"document_id"`
	TaskID         *int64 `json:"task_id"`
	AnnouncementID *int64 `json:"announcement_id"`
}

// CreateVersionInput представляет входные данные для создания версии файла.
type CreateVersionInput struct {
	FileID  int64  `json:"file_id" validate:"required"`
	Comment string `json:"comment" validate:"omitempty,max=500"`
	UserID  int64  `json:"-"` // Заполняется из контекста авторизации
}

// FileResponse представляет ответ с информацией о файле.
type FileResponse struct {
	ID             int64     `json:"id"`
	OriginalName   string    `json:"original_name"`
	Size           int64     `json:"size"`
	MimeType       string    `json:"mime_type"`
	Checksum       string    `json:"checksum"`
	UploadedBy     int64     `json:"uploaded_by"`
	DocumentID     *int64    `json:"document_id,omitempty"`
	TaskID         *int64    `json:"task_id,omitempty"`
	AnnouncementID *int64    `json:"announcement_id,omitempty"`
	IsTemporary    bool      `json:"is_temporary"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DownloadURL    string    `json:"download_url,omitempty"`
}

// FileVersionResponse представляет ответ с информацией о версии файла.
type FileVersionResponse struct {
	ID            int64     `json:"id"`
	VersionNumber int       `json:"version_number"`
	Size          int64     `json:"size"`
	Checksum      string    `json:"checksum"`
	Comment       string    `json:"comment,omitempty"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	DownloadURL   string    `json:"download_url,omitempty"`
}

// FileListResponse представляет ответ со списком файлов с пагинацией.
type FileListResponse struct {
	Files      interface{} `json:"files"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// UploadResponse представляет ответ после успешной загрузки файла.
type UploadResponse struct {
	FileID       int64  `json:"file_id"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	Checksum     string `json:"checksum"`
}

// DownloadResponse представляет ответ для скачивания файла.
type DownloadResponse struct {
	PresignedURL string `json:"presigned_url"`
	FileName     string `json:"file_name"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
}
