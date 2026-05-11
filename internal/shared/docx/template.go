// Package docx provides minimal stdlib-only helpers for DOCX template
// substitution. A DOCX file is a ZIP archive whose primary content is
// word/document.xml; Substitute applies string replacements to that entry
// and returns a new DOCX byte stream preserving all original entries.
package docx

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const documentEntryName = "word/document.xml"

// ErrInvalidTemplate is returned when input bytes cannot be parsed as a
// DOCX (missing word/document.xml entry or corrupted ZIP archive).
var ErrInvalidTemplate = errors.New("docx: invalid template")

// Substitute applies string replacements to word/document.xml inside
// templateBytes and returns the resulting DOCX bytes. Replacement values are
// inserted verbatim; callers are responsible for XML-escaping when needed.
func Substitute(templateBytes []byte, replacements map[string]string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(templateBytes), int64(len(templateBytes)))
	if err != nil {
		return nil, errors.Join(ErrInvalidTemplate, err)
	}

	if !containsEntry(r, documentEntryName) {
		return nil, fmt.Errorf("%w: missing %s", ErrInvalidTemplate, documentEntryName)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for _, f := range r.File {
		header := f.FileHeader
		dst, err := w.CreateHeader(&header)
		if err != nil {
			return nil, fmt.Errorf("docx: create header %s: %w", f.Name, err)
		}

		content, err := readEntry(f)
		if err != nil {
			return nil, err
		}

		if f.Name == documentEntryName {
			content = applyReplacements(content, replacements)
		}

		if _, err := dst.Write(content); err != nil {
			return nil, fmt.Errorf("docx: write %s: %w", f.Name, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("docx: finalize zip: %w", err)
	}

	return buf.Bytes(), nil
}

func containsEntry(r *zip.Reader, name string) bool {
	for _, f := range r.File {
		if f.Name == name {
			return true
		}
	}
	return false
}

func readEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("docx: open %s: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()

	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("docx: read %s: %w", f.Name, err)
	}
	return b, nil
}

func applyReplacements(content []byte, replacements map[string]string) []byte {
	if len(replacements) == 0 {
		return content
	}
	s := string(content)
	for k, v := range replacements {
		s = strings.ReplaceAll(s, k, v)
	}
	return []byte(s)
}
