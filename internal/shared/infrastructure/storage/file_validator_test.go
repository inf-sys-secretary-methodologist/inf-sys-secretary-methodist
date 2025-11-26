package storage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileValidator(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	assert.NotNil(t, v)
	assert.Equal(t, cfg.MaxFileSize, v.MaxFileSize())
}

func TestValidateFile_Size(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024, // 1KB
		AllowedMimeTypes:  []string{"text/plain"},
		AllowedExtensions: []string{".txt"},
	}
	v := NewFileValidator(cfg)

	tests := []struct {
		name      string
		fileSize  int64
		wantValid bool
	}{
		{
			name:      "valid size",
			fileSize:  500,
			wantValid: true,
		},
		{
			name:      "exact limit",
			fileSize:  1024,
			wantValid: true,
		},
		{
			name:      "exceeds limit",
			fileSize:  2048,
			wantValid: false,
		},
		{
			name:      "empty file",
			fileSize:  0,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.ValidateFile("test.txt", tt.fileSize, "text/plain", nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestValidateFile_Extension(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024 * 1024,
		AllowedMimeTypes:  []string{"application/pdf", "image/png"},
		AllowedExtensions: []string{".pdf", ".png"},
	}
	v := NewFileValidator(cfg)

	tests := []struct {
		name      string
		fileName  string
		wantValid bool
	}{
		{
			name:      "allowed extension pdf",
			fileName:  "document.pdf",
			wantValid: true,
		},
		{
			name:      "allowed extension png",
			fileName:  "image.png",
			wantValid: true,
		},
		{
			name:      "uppercase extension",
			fileName:  "document.PDF",
			wantValid: true,
		},
		{
			name:      "disallowed extension",
			fileName:  "script.exe",
			wantValid: false,
		},
		{
			name:      "no extension",
			fileName:  "noextension",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.ValidateFile(tt.fileName, 100, "application/pdf", nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestValidateFile_MimeType(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024 * 1024,
		AllowedMimeTypes:  []string{"application/pdf", "image/png"},
		AllowedExtensions: []string{".pdf", ".png"},
	}
	v := NewFileValidator(cfg)

	tests := []struct {
		name        string
		contentType string
		wantValid   bool
	}{
		{
			name:        "allowed mime type pdf",
			contentType: "application/pdf",
			wantValid:   true,
		},
		{
			name:        "allowed mime type png",
			contentType: "image/png",
			wantValid:   true,
		},
		{
			name:        "disallowed mime type",
			contentType: "application/x-executable",
			wantValid:   false,
		},
		{
			name:        "octet-stream is acceptable",
			contentType: "application/octet-stream",
			wantValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.ValidateFile("test.pdf", 100, tt.contentType, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestValidateFile_MagicBytes(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	tests := []struct {
		name        string
		fileName    string
		contentType string
		magicBytes  []byte
		wantValid   bool
		wantType    string
	}{
		{
			name:        "valid PDF magic bytes",
			fileName:    "document.pdf",
			contentType: "application/pdf",
			magicBytes:  []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}, // %PDF-1.4
			wantValid:   true,
			wantType:    "application/pdf",
		},
		{
			name:        "valid JPEG magic bytes",
			fileName:    "image.jpg",
			contentType: "image/jpeg",
			magicBytes:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			wantValid:   true,
			wantType:    "image/jpeg",
		},
		{
			name:        "valid PNG magic bytes",
			fileName:    "image.png",
			contentType: "image/png",
			magicBytes:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantValid:   true,
			wantType:    "image/png",
		},
		{
			name:        "mismatched magic bytes",
			fileName:    "fake.pdf",
			contentType: "application/pdf",
			magicBytes:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG magic in PDF
			wantValid:   false,
			wantType:    "image/png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.magicBytes)
			result, err := v.ValidateFile(tt.fileName, 100, tt.contentType, reader)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)
			if tt.wantType != "" {
				assert.Equal(t, tt.wantType, result.DetectedType)
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal filename",
			input:    "document.pdf",
			expected: "document.pdf",
		},
		{
			name:     "remove path traversal",
			input:    "../../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "remove null bytes",
			input:    "file\x00name.txt",
			expected: "filename.txt",
		},
		{
			name:     "remove special characters",
			input:    "file<name>:test.txt",
			expected: "filenametest.txt",
		},
		{
			name:     "preserve unicode",
			input:    "документ.pdf",
			expected: "документ.pdf",
		},
		{
			name:     "handle windows-style path on unix",
			input:    "C:\\Users\\file.txt",
			expected: "CUsersfile.txt", // backslashes removed on unix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, err := v.ValidateFileName(tt.input)
			if tt.expected == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, sanitized)
			}
		})
	}
}

func TestValidateFileName_Empty(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	_, err := v.ValidateFileName("")
	assert.Error(t, err)
}

func TestIsExtensionAllowed(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024,
		AllowedMimeTypes:  []string{"application/pdf"},
		AllowedExtensions: []string{".pdf", ".doc"},
	}
	v := NewFileValidator(cfg)

	tests := []struct {
		ext      string
		expected bool
	}{
		{".pdf", true},
		{".PDF", true},
		{".doc", true},
		{".exe", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			assert.Equal(t, tt.expected, v.IsExtensionAllowed(tt.ext))
		})
	}
}

func TestIsMimeTypeAllowed(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024,
		AllowedMimeTypes:  []string{"application/pdf", "image/png"},
		AllowedExtensions: []string{".pdf"},
	}
	v := NewFileValidator(cfg)

	tests := []struct {
		mimeType string
		expected bool
	}{
		{"application/pdf", true},
		{"image/png", true},
		{"application/x-executable", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			assert.Equal(t, tt.expected, v.IsMimeTypeAllowed(tt.mimeType))
		})
	}
}

func TestDefaultFileValidatorConfig(t *testing.T) {
	cfg := DefaultFileValidatorConfig()

	assert.Equal(t, int64(50*1024*1024), cfg.MaxFileSize)
	assert.NotEmpty(t, cfg.AllowedMimeTypes)
	assert.NotEmpty(t, cfg.AllowedExtensions)
	assert.Contains(t, cfg.AllowedMimeTypes, "application/pdf")
	assert.Contains(t, cfg.AllowedExtensions, ".pdf")
}

func TestAllowedExtensions(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024,
		AllowedMimeTypes:  []string{"application/pdf"},
		AllowedExtensions: []string{".pdf", ".doc"},
	}
	v := NewFileValidator(cfg)

	extensions := v.AllowedExtensions()
	assert.Len(t, extensions, 2)
}

func TestAllowedMimeTypes(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024,
		AllowedMimeTypes:  []string{"application/pdf", "image/png"},
		AllowedExtensions: []string{".pdf"},
	}
	v := NewFileValidator(cfg)

	types := v.AllowedMimeTypes()
	assert.Len(t, types, 2)
}
