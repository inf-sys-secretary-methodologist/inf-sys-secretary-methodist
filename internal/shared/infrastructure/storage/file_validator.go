// Package storage provides file storage and validation utilities.
package storage

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// FileValidator validates uploaded files
type FileValidator struct {
	maxFileSize      int64
	allowedMimeTypes map[string]bool
	allowedExtensions map[string]bool
	magicBytes       map[string][]byte
}

// FileValidatorConfig contains configuration for file validation
type FileValidatorConfig struct {
	MaxFileSize       int64
	AllowedMimeTypes  []string
	AllowedExtensions []string
}

// DefaultFileValidatorConfig returns default validation configuration
func DefaultFileValidatorConfig() FileValidatorConfig {
	return FileValidatorConfig{
		MaxFileSize: 50 * 1024 * 1024, // 50MB
		AllowedMimeTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-powerpoint",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"text/plain",
			"text/csv",
			"application/zip",
			"application/x-rar-compressed",
			"application/x-7z-compressed",
		},
		AllowedExtensions: []string{
			".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
			".jpg", ".jpeg", ".png", ".gif", ".webp",
			".txt", ".csv",
			".zip", ".rar", ".7z",
		},
	}
}

// NewFileValidator creates a new file validator
func NewFileValidator(cfg FileValidatorConfig) *FileValidator {
	allowedMimeTypes := make(map[string]bool)
	for _, mt := range cfg.AllowedMimeTypes {
		allowedMimeTypes[mt] = true
	}

	allowedExtensions := make(map[string]bool)
	for _, ext := range cfg.AllowedExtensions {
		allowedExtensions[strings.ToLower(ext)] = true
	}

	return &FileValidator{
		maxFileSize:       cfg.MaxFileSize,
		allowedMimeTypes:  allowedMimeTypes,
		allowedExtensions: allowedExtensions,
		magicBytes:        initMagicBytes(),
	}
}

// initMagicBytes initializes magic bytes for common file types
func initMagicBytes() map[string][]byte {
	return map[string][]byte{
		"application/pdf":  {0x25, 0x50, 0x44, 0x46}, // %PDF
		"image/jpeg":       {0xFF, 0xD8, 0xFF},
		"image/png":        {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		"image/gif":        {0x47, 0x49, 0x46, 0x38}, // GIF8
		"application/zip":  {0x50, 0x4B, 0x03, 0x04}, // PK..
		"application/x-rar-compressed": {0x52, 0x61, 0x72, 0x21}, // Rar!
	}
}

// ValidationResult contains the result of file validation
type ValidationResult struct {
	Valid        bool
	Errors       []string
	DetectedType string
	SanitizedName string
}

// ValidateFile validates a file based on configured rules
func (v *FileValidator) ValidateFile(fileName string, fileSize int64, contentType string, reader io.Reader) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Sanitize filename
	result.SanitizedName = v.sanitizeFileName(fileName)

	// Validate file size
	if fileSize > v.maxFileSize {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"Размер файла (%d байт) превышает максимально допустимый (%d байт)",
			fileSize, v.maxFileSize,
		))
	}

	if fileSize == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Файл пуст")
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "Файл должен иметь расширение")
	} else if !v.allowedExtensions[ext] {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"Расширение файла '%s' не разрешено", ext,
		))
	}

	// Validate MIME type
	if contentType != "" && contentType != "application/octet-stream" {
		if !v.allowedMimeTypes[contentType] {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Тип файла '%s' не разрешен", contentType,
			))
		}
	}

	// Validate magic bytes if reader is provided
	if reader != nil {
		detectedType, err := v.detectFileType(reader)
		if err == nil && detectedType != "" {
			result.DetectedType = detectedType
			// Check if detected type matches content type
			if contentType != "" && contentType != "application/octet-stream" && detectedType != contentType {
				// Allow some flexibility for related types
				if !v.areRelatedTypes(contentType, detectedType) {
					result.Valid = false
					result.Errors = append(result.Errors, fmt.Sprintf(
						"Содержимое файла не соответствует заявленному типу (заявлено: %s, обнаружено: %s)",
						contentType, detectedType,
					))
				}
			}
		}
	}

	return result, nil
}

// ValidateFileName validates and sanitizes a filename
func (v *FileValidator) ValidateFileName(fileName string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("имя файла не может быть пустым")
	}

	sanitized := v.sanitizeFileName(fileName)

	if sanitized == "" || sanitized == "." {
		return "", fmt.Errorf("недопустимое имя файла")
	}

	return sanitized, nil
}

// sanitizeFileName removes potentially dangerous characters from filename
func (v *FileValidator) sanitizeFileName(fileName string) string {
	// Get base name (remove path)
	fileName = filepath.Base(fileName)

	// Remove null bytes and control characters
	var sanitized strings.Builder
	for _, r := range fileName {
		if r > 31 && r != 127 && r != '/' && r != '\\' && r != ':' && r != '*' && r != '?' && r != '"' && r != '<' && r != '>' && r != '|' {
			sanitized.WriteRune(r)
		}
	}

	result := sanitized.String()

	// Limit length
	if len(result) > 255 {
		ext := filepath.Ext(result)
		name := strings.TrimSuffix(result, ext)
		maxNameLen := 255 - len(ext)
		if maxNameLen > 0 && len(name) > maxNameLen {
			name = name[:maxNameLen]
		}
		result = name + ext
	}

	return result
}

// detectFileType detects file type based on magic bytes
func (v *FileValidator) detectFileType(reader io.Reader) (string, error) {
	// Read first 8 bytes for magic number detection
	header := make([]byte, 8)
	n, err := reader.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	header = header[:n]

	for mimeType, magic := range v.magicBytes {
		if len(header) >= len(magic) && bytes.HasPrefix(header, magic) {
			return mimeType, nil
		}
	}

	return "", nil
}

// areRelatedTypes checks if two MIME types are related (e.g., variants of same format)
func (v *FileValidator) areRelatedTypes(type1, type2 string) bool {
	relatedGroups := [][]string{
		{"application/zip", "application/x-zip-compressed", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{"image/jpeg", "image/jpg"},
	}

	for _, group := range relatedGroups {
		has1, has2 := false, false
		for _, t := range group {
			if t == type1 {
				has1 = true
			}
			if t == type2 {
				has2 = true
			}
		}
		if has1 && has2 {
			return true
		}
	}

	return false
}

// MaxFileSize returns the maximum allowed file size
func (v *FileValidator) MaxFileSize() int64 {
	return v.maxFileSize
}

// IsExtensionAllowed checks if a file extension is allowed
func (v *FileValidator) IsExtensionAllowed(ext string) bool {
	return v.allowedExtensions[strings.ToLower(ext)]
}

// IsMimeTypeAllowed checks if a MIME type is allowed
func (v *FileValidator) IsMimeTypeAllowed(mimeType string) bool {
	return v.allowedMimeTypes[mimeType]
}

// AllowedExtensions returns list of allowed extensions
func (v *FileValidator) AllowedExtensions() []string {
	extensions := make([]string, 0, len(v.allowedExtensions))
	for ext := range v.allowedExtensions {
		extensions = append(extensions, ext)
	}
	return extensions
}

// AllowedMimeTypes returns list of allowed MIME types
func (v *FileValidator) AllowedMimeTypes() []string {
	types := make([]string, 0, len(v.allowedMimeTypes))
	for t := range v.allowedMimeTypes {
		types = append(types, t)
	}
	return types
}
