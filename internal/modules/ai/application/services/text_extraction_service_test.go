package services

import (
	"archive/zip"
	"bytes"
	"testing"
)

// buildMinimalDocx creates a minimal valid .docx (zip archive with word/document.xml).
func buildMinimalDocx(t *testing.T, bodyXML string) []byte {
	t.Helper()

	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body>` + bodyXML + `</w:body></w:document>`

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, err := w.Create("word/document.xml")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := f.Write([]byte(xmlContent)); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}

	return buf.Bytes()
}

func TestTextExtractionService_ExtractDocx(t *testing.T) {
	svc := NewTextExtractionService()

	bodyXML := `<w:p><w:r><w:t>Hello World</w:t></w:r></w:p>` +
		`<w:p><w:r><w:t>Second paragraph</w:t></w:r></w:p>`

	data := buildMinimalDocx(t, bodyXML)

	text, err := svc.Extract(data, MimeDocx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Hello World\nSecond paragraph" {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestTextExtractionService_ExtractDocx_MultipleRunsInParagraph(t *testing.T) {
	svc := NewTextExtractionService()

	bodyXML := `<w:p>` +
		`<w:r><w:t>Part one </w:t></w:r>` +
		`<w:r><w:t>part two</w:t></w:r>` +
		`</w:p>`

	data := buildMinimalDocx(t, bodyXML)

	text, err := svc.Extract(data, MimeDocx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Part one part two" {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestTextExtractionService_ExtractPlainText(t *testing.T) {
	svc := NewTextExtractionService()

	input := "This is a plain text document.\nWith multiple lines."
	text, err := svc.Extract([]byte(input), MimeText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != input {
		t.Errorf("expected %q, got %q", input, text)
	}
}

func TestTextExtractionService_ExtractCSV(t *testing.T) {
	svc := NewTextExtractionService()

	input := "name,age\nAlice,30\nBob,25"
	text, err := svc.Extract([]byte(input), MimeCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != input {
		t.Errorf("expected %q, got %q", input, text)
	}
}

func TestTextExtractionService_CanExtract(t *testing.T) {
	svc := NewTextExtractionService()

	tests := []struct {
		mimeType string
		expected bool
	}{
		{MimeDocx, true},
		{MimePDF, true},
		{MimeText, true},
		{MimeCSV, true},
		{"application/octet-stream", false},
		{"image/png", false},
		{"", false},
	}

	for _, tt := range tests {
		got := svc.CanExtract(tt.mimeType)
		if got != tt.expected {
			t.Errorf("CanExtract(%q): expected %v, got %v", tt.mimeType, tt.expected, got)
		}
	}
}

func TestTextExtractionService_UnsupportedMIME(t *testing.T) {
	svc := NewTextExtractionService()

	_, err := svc.Extract([]byte("data"), "application/octet-stream")
	if err == nil {
		t.Fatal("expected error for unsupported MIME type, got nil")
	}
}

func TestTextExtractionService_InvalidDocx(t *testing.T) {
	svc := NewTextExtractionService()

	_, err := svc.Extract([]byte("not a zip file"), MimeDocx)
	if err == nil {
		t.Fatal("expected error for invalid docx, got nil")
	}
}
