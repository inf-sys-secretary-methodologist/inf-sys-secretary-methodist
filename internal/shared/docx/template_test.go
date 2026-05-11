package docx_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/docx"
)

const documentEntryName = "word/document.xml"

func buildTestDOCX(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func extractEntry(t *testing.T, docxBytes []byte, entryName string) string {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}
	for _, f := range r.File {
		if f.Name != entryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open entry %s: %v", entryName, err)
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read entry %s: %v", entryName, err)
		}
		return string(b)
	}
	t.Fatalf("entry %s not found", entryName)
	return ""
}

func TestSubstitute(t *testing.T) {
	t.Parallel()

	t.Run("empty replacements returns unchanged document.xml", func(t *testing.T) {
		t.Parallel()
		const original = "Year: {{year}} | Section: {{section}}"
		input := buildTestDOCX(t, map[string]string{documentEntryName: original})

		out, err := docx.Substitute(input, nil)
		if err != nil {
			t.Fatalf("Substitute: unexpected error: %v", err)
		}

		got := extractEntry(t, out, documentEntryName)
		if got != original {
			t.Errorf("document.xml mismatch:\n  want %q\n  got  %q", original, got)
		}
	})

	t.Run("single placeholder replaced", func(t *testing.T) {
		t.Parallel()
		input := buildTestDOCX(t, map[string]string{
			documentEntryName: "Report for {{year}}",
		})

		out, err := docx.Substitute(input, map[string]string{"{{year}}": "2026"})
		if err != nil {
			t.Fatalf("Substitute: unexpected error: %v", err)
		}

		got := extractEntry(t, out, documentEntryName)
		const want = "Report for 2026"
		if got != want {
			t.Errorf("document.xml mismatch:\n  want %q\n  got  %q", want, got)
		}
	})

	t.Run("multiple placeholders replaced", func(t *testing.T) {
		t.Parallel()
		input := buildTestDOCX(t, map[string]string{
			documentEntryName: "{{year}} | {{section_a}} | {{section_b}}",
		})

		out, err := docx.Substitute(input, map[string]string{
			"{{year}}":      "2026",
			"{{section_a}}": "Curricula",
			"{{section_b}}": "Grades",
		})
		if err != nil {
			t.Fatalf("Substitute: unexpected error: %v", err)
		}

		got := extractEntry(t, out, documentEntryName)
		const want = "2026 | Curricula | Grades"
		if got != want {
			t.Errorf("document.xml mismatch:\n  want %q\n  got  %q", want, got)
		}
	})

	t.Run("placeholder absent in template is no-op without error", func(t *testing.T) {
		t.Parallel()
		const original = "Static content with no placeholders"
		input := buildTestDOCX(t, map[string]string{documentEntryName: original})

		out, err := docx.Substitute(input, map[string]string{"{{year}}": "2026"})
		if err != nil {
			t.Fatalf("Substitute: unexpected error: %v", err)
		}

		got := extractEntry(t, out, documentEntryName)
		if got != original {
			t.Errorf("document.xml should be unchanged:\n  want %q\n  got  %q", original, got)
		}
	})

	t.Run("preserves sibling entries untouched", func(t *testing.T) {
		t.Parallel()
		const sibling = "<?xml version=\"1.0\"?><Types/>"
		input := buildTestDOCX(t, map[string]string{
			documentEntryName:     "Year: {{year}}",
			"[Content_Types].xml": sibling,
		})

		out, err := docx.Substitute(input, map[string]string{"{{year}}": "2026"})
		if err != nil {
			t.Fatalf("Substitute: unexpected error: %v", err)
		}

		gotDoc := extractEntry(t, out, documentEntryName)
		if gotDoc != "Year: 2026" {
			t.Errorf("document.xml mismatch: got %q", gotDoc)
		}

		gotSibling := extractEntry(t, out, "[Content_Types].xml")
		if gotSibling != sibling {
			t.Errorf("sibling mutated:\n  want %q\n  got  %q", sibling, gotSibling)
		}
	})

	t.Run("nil input returns ErrInvalidTemplate", func(t *testing.T) {
		t.Parallel()
		_, err := docx.Substitute(nil, map[string]string{"{{year}}": "2026"})
		if !errors.Is(err, docx.ErrInvalidTemplate) {
			t.Errorf("expected ErrInvalidTemplate, got %v", err)
		}
	})

	t.Run("random bytes return ErrInvalidTemplate", func(t *testing.T) {
		t.Parallel()
		_, err := docx.Substitute([]byte("not a zip archive at all"), nil)
		if !errors.Is(err, docx.ErrInvalidTemplate) {
			t.Errorf("expected ErrInvalidTemplate, got %v", err)
		}
	})

	t.Run("zip without document.xml returns ErrInvalidTemplate", func(t *testing.T) {
		t.Parallel()
		input := buildTestDOCX(t, map[string]string{
			"[Content_Types].xml": "<?xml version=\"1.0\"?><Types/>",
		})

		_, err := docx.Substitute(input, nil)
		if !errors.Is(err, docx.ErrInvalidTemplate) {
			t.Errorf("expected ErrInvalidTemplate, got %v", err)
		}
	})
}
