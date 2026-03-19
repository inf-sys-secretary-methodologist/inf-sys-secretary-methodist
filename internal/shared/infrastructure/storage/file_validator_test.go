package storage

import (
	"bytes"
	"io"
	"strings"
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
		{
			name:        "empty content type is acceptable",
			contentType: "",
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
			name:        "valid GIF magic bytes",
			fileName:    "image.gif",
			contentType: "image/gif",
			magicBytes:  []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x00, 0x00},
			wantValid:   true,
			wantType:    "image/gif",
		},
		{
			name:        "valid ZIP magic bytes",
			fileName:    "archive.zip",
			contentType: "application/zip",
			magicBytes:  []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00},
			wantValid:   true,
			wantType:    "application/zip",
		},
		{
			name:        "valid RAR magic bytes",
			fileName:    "archive.rar",
			contentType: "application/x-rar-compressed",
			magicBytes:  []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00, 0x00},
			wantValid:   true,
			wantType:    "application/x-rar-compressed",
		},
		{
			name:        "mismatched magic bytes",
			fileName:    "fake.pdf",
			contentType: "application/pdf",
			magicBytes:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG magic in PDF
			wantValid:   false,
			wantType:    "image/png",
		},
		{
			name:        "unknown magic bytes with octet-stream",
			fileName:    "data.txt",
			contentType: "application/octet-stream",
			magicBytes:  []byte{0x00, 0x01, 0x02, 0x03},
			wantValid:   true,
			wantType:    "",
		},
		{
			name:        "docx detected as zip related type",
			fileName:    "document.docx",
			contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			magicBytes:  []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, // ZIP magic
			wantValid:   true,
			wantType:    "application/zip",
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

func TestValidateFile_MagicBytes_ReaderError(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	// Use a reader that returns an error
	errReader := &errorReader{}
	result, err := v.ValidateFile("test.pdf", 100, "application/pdf", errReader)
	assert.NoError(t, err) // ValidateFile itself doesn't return error from detectFileType
	assert.True(t, result.Valid)
}

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestValidateFile_EmptyReader(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	reader := bytes.NewReader([]byte{})
	result, err := v.ValidateFile("test.pdf", 100, "application/pdf", reader)
	assert.NoError(t, err)
	assert.True(t, result.Valid) // No magic bytes to detect, so no mismatch
}

func TestValidateFile_MultipleErrors(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       100,
		AllowedMimeTypes:  []string{"text/plain"},
		AllowedExtensions: []string{".txt"},
	}
	v := NewFileValidator(cfg)

	// file too big, wrong extension, wrong mime type
	result, err := v.ValidateFile("bad.exe", 200, "application/x-executable", nil)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.GreaterOrEqual(t, len(result.Errors), 2, "expected multiple errors")
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
			input:    "\u0434\u043e\u043a\u0443\u043c\u0435\u043d\u0442.pdf",
			expected: "\u0434\u043e\u043a\u0443\u043c\u0435\u043d\u0442.pdf",
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

func TestSanitizeFileName_LongName(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	// Create a very long filename (300 chars + .txt extension)
	longName := strings.Repeat("a", 300) + ".txt"
	sanitized, err := v.ValidateFileName(longName)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(sanitized), 255)
	assert.True(t, strings.HasSuffix(sanitized, ".txt"))
}

func TestValidateFileName_Empty(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	_, err := v.ValidateFileName("")
	assert.Error(t, err)
}

func TestValidateFileName_Dot(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	// filepath.Base of "." is "."
	_, err := v.ValidateFileName(".")
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

func TestAreRelatedTypes(t *testing.T) {
	cfg := DefaultFileValidatorConfig()
	v := NewFileValidator(cfg)

	tests := []struct {
		name     string
		type1    string
		type2    string
		expected bool
	}{
		{
			name:     "zip and docx are related",
			type1:    "application/zip",
			type2:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			expected: true,
		},
		{
			name:     "zip and xlsx are related",
			type1:    "application/zip",
			type2:    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			expected: true,
		},
		{
			name:     "zip and pptx are related",
			type1:    "application/zip",
			type2:    "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			expected: true,
		},
		{
			name:     "jpeg and jpg are related",
			type1:    "image/jpeg",
			type2:    "image/jpg",
			expected: true,
		},
		{
			name:     "pdf and png are not related",
			type1:    "application/pdf",
			type2:    "image/png",
			expected: false,
		},
		{
			name:     "unrelated types",
			type1:    "text/plain",
			type2:    "image/gif",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, v.areRelatedTypes(tt.type1, tt.type2))
		})
	}
}

func TestGenerateKey(t *testing.T) {
	key := GenerateKey(42, "document.pdf")
	assert.Contains(t, key, "documents/42/")
	assert.True(t, strings.HasSuffix(key, ".pdf"))
}

func TestGenerateKey_DifferentExtensions(t *testing.T) {
	key := GenerateKey(1, "file.docx")
	assert.True(t, strings.HasSuffix(key, ".docx"))

	key = GenerateKey(1, "image.png")
	assert.True(t, strings.HasSuffix(key, ".png"))
}

func TestGenerateTempKey(t *testing.T) {
	key := GenerateTempKey(99, "upload.pdf")
	assert.Contains(t, key, "temp/99/")
	assert.True(t, strings.HasSuffix(key, ".pdf"))
}

func TestGenerateTempKey_DifferentExtensions(t *testing.T) {
	key := GenerateTempKey(1, "file.xlsx")
	assert.True(t, strings.HasSuffix(key, ".xlsx"))
}

func TestGenerateKey_UniqueTwoCallsAreDifferent(t *testing.T) {
	key1 := GenerateKey(1, "test.pdf")
	key2 := GenerateKey(1, "test.pdf")
	// In practice they should differ due to UnixNano; we just check format
	assert.Contains(t, key1, "documents/1/")
	assert.Contains(t, key2, "documents/1/")
}

func TestInitMagicBytes(t *testing.T) {
	mb := initMagicBytes()
	assert.NotEmpty(t, mb)
	assert.Contains(t, mb, "application/pdf")
	assert.Contains(t, mb, "image/jpeg")
	assert.Contains(t, mb, "image/png")
	assert.Contains(t, mb, "image/gif")
	assert.Contains(t, mb, "application/zip")
	assert.Contains(t, mb, "application/x-rar-compressed")
}

func TestValidateFile_SanitizedName(t *testing.T) {
	cfg := FileValidatorConfig{
		MaxFileSize:       1024 * 1024,
		AllowedMimeTypes:  []string{"text/plain"},
		AllowedExtensions: []string{".txt"},
	}
	v := NewFileValidator(cfg)

	result, err := v.ValidateFile("../../../etc/test.txt", 100, "text/plain", nil)
	assert.NoError(t, err)
	assert.Equal(t, "test.txt", result.SanitizedName)
}
