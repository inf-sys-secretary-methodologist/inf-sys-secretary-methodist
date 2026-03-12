// Package entities содержит доменные сущности модуля files.
package entities

import "time"

// FileVersion представляет версию файла (для версионирования документов).
type FileVersion struct {
	ID             int64     `json:"id"`
	FileMetadataID int64     `json:"file_metadata_id"`  // Ссылка на основной файл
	VersionNumber  int       `json:"version_number"`    // Номер версии
	StorageKey     string    `json:"storage_key"`       // Ключ в S3/MinIO хранилище
	Size           int64     `json:"size"`              // Размер версии в байтах
	Checksum       string    `json:"checksum"`          // SHA-256 хеш версии
	Comment        string    `json:"comment,omitempty"` // Комментарий к версии
	CreatedBy      int64     `json:"created_by"`        // ID автора версии
	CreatedAt      time.Time `json:"created_at"`
}

// NewFileVersion создаёт новый экземпляр FileVersion.
func NewFileVersion(fileMetadataID int64, versionNumber int, storageKey, checksum, comment string, size int64, createdBy int64) *FileVersion {
	return &FileVersion{
		FileMetadataID: fileMetadataID,
		VersionNumber:  versionNumber,
		StorageKey:     storageKey,
		Size:           size,
		Checksum:       checksum,
		Comment:        comment,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
	}
}
