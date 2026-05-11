// Package docx provides minimal stdlib-only helpers for DOCX template
// substitution. A DOCX file is a ZIP archive whose primary content is
// word/document.xml; Substitute applies string replacements to that entry
// and returns a new DOCX byte stream preserving all original entries.
package docx

import "errors"

// ErrInvalidTemplate is returned when input bytes cannot be parsed as a
// DOCX (missing word/document.xml entry or corrupted ZIP archive).
var ErrInvalidTemplate = errors.New("docx: invalid template")

// Substitute applies string replacements to word/document.xml inside
// templateBytes and returns the resulting DOCX bytes. Replacement values are
// inserted verbatim; callers are responsible for XML-escaping when needed.
func Substitute(templateBytes []byte, replacements map[string]string) ([]byte, error) {
	return nil, errors.New("docx: substitute not implemented")
}
