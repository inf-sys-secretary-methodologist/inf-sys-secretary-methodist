// Package entities содержит доменные сущности модуля files.
package entities

import "time"

// FileVersion представляет версию файла (для версионирования документов).
type FileVersion struct {
	ID             int64     `db:"id" json:"id"`
	FileMetadataID int64     `db:"file_metadata_id" json:"file_metadata_id"` // Ссылка на основной файл
	VersionNumber  int       `db:"version_number" json:"version_number"`     // Номер версии
	StorageKey     string    `db:"storage_key" json:"storage_key"`           // Ключ в S3/MinIO хранилище
	Size           int64     `db:"size" json:"size"`                         // Размер версии в байтах
	Checksum       string    `db:"checksum" json:"checksum"`                 // SHA-256 хеш версии
	Comment        string    `db:"comment" json:"comment,omitempty"`         // Комментарий к версии
	CreatedBy      int64     `db:"created_by" json:"created_by"`             // ID автора версии
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
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
