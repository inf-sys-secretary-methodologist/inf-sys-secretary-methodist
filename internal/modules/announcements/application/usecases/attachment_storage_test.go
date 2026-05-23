package usecases

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAttachmentStorageKey_PathTraversalSanitized pins v0.163.0 ADR-4
// (#303 TIER 1): pre-fix attachmentStorageKey used path.Join which
// internally calls path.Clean — Clean collapses `..` components.
// A filename like `evil/../../escape.bin` would yield a key like
// `announcements/1/{uuid}-evil/../../escape.bin` that path.Clean
// then reduces to `escape.bin`, escaping the per-announcement prefix.
//
// After fix, the function must keep the key strictly under
// `announcements/{id}/` regardless of the filename's contents.
func TestAttachmentStorageKey_PathTraversalSanitized(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{"parent-dir traversal", "../../escape.bin"},
		{"embedded traversal", "evil/../../escape.bin"},
		{"absolute path", "/etc/passwd"},
		{"backslash traversal", "..\\..\\escape.bin"},
		{"current-dir prefix", "./escape.bin"},
		{"nested traversal", "a/b/../../../../escape.bin"},
		{"dot-only", ".."},
		{"trailing slashes", "evil///"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := attachmentStorageKey(42, tc.fileName)
			// Must stay strictly within announcements/42/
			assert.True(t, strings.HasPrefix(key, "announcements/42/"),
				"key %q must keep the announcements/42/ prefix", key)
			// No `..` components or absolute slashes that escape the prefix.
			assert.NotContains(t, key, "..",
				"key %q must not contain parent-dir components", key)
			assert.NotContains(t, strings.TrimPrefix(key, "announcements/42/"), "/",
				"sanitized filename must not introduce new directory levels: %q", key)
		})
	}
}

// TestAttachmentStorageKey_LegitimateFilenamePreserved verifies the
// sanitizer does NOT break normal filenames.
func TestAttachmentStorageKey_LegitimateFilenamePreserved(t *testing.T) {
	tests := []string{
		"report.pdf",
		"Отчёт_2026.docx",
		"image (1).png",
		"file with spaces.txt",
	}

	for _, fn := range tests {
		t.Run(fn, func(t *testing.T) {
			key := attachmentStorageKey(1, fn)
			assert.True(t, strings.HasPrefix(key, "announcements/1/"))
			assert.True(t, strings.HasSuffix(key, fn),
				"legitimate filename %q must be preserved in key tail", fn)
		})
	}
}
