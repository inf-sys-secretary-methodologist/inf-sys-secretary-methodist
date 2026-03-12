// Package services contains application services for the AI module.
package services

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dslipak/pdf"
)

// Supported MIME types for text extraction.
const (
	MimeDocx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MimePDF  = "application/pdf"
	MimeText = "text/plain"
	MimeCSV  = "text/csv"
)

// TextExtractionService extracts plain text from binary document formats.
type TextExtractionService struct {
	supportedMIME map[string]bool
}

// NewTextExtractionService creates a new TextExtractionService.
func NewTextExtractionService() *TextExtractionService {
	return &TextExtractionService{
		supportedMIME: map[string]bool{
			MimeDocx: true,
			MimePDF:  true,
			MimeText: true,
			MimeCSV:  true,
		},
	}
}

// CanExtract returns true if the given MIME type is supported for text extraction.
func (s *TextExtractionService) CanExtract(mimeType string) bool {
	return s.supportedMIME[mimeType]
}

// Extract extracts plain text from the given data based on MIME type.
func (s *TextExtractionService) Extract(data []byte, mimeType string) (string, error) {
	switch mimeType {
	case MimeDocx:
		return s.extractDocx(data)
	case MimePDF:
		return s.extractPDF(data)
	case MimeText, MimeCSV:
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported MIME type: %s", mimeType)
	}
}

// extractDocx extracts text from a .docx file by parsing word/document.xml.
func (s *TextExtractionService) extractDocx(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open docx as zip: %w", err)
	}

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open word/document.xml: %w", err)
			}
			defer func() { _ = rc.Close() }()

			content, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read word/document.xml: %w", err)
			}

			return parseDocxXML(content)
		}
	}

	return "", fmt.Errorf("word/document.xml not found in docx archive")
}

// parseDocxXML extracts text from Office Open XML word/document.xml content.
// It walks the XML tree, collecting text from <w:t> elements and inserting
// paragraph breaks (<w:p>) as newlines.
func parseDocxXML(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var builder strings.Builder
	inText := false
	paragraphHasText := false

	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to parse docx XML: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p": // <w:p> — paragraph
				if paragraphHasText {
					builder.WriteString("\n")
				}
				paragraphHasText = false
			case "t": // <w:t> — text run
				inText = true
			case "tab": // <w:tab> — tab character
				builder.WriteString("\t")
			case "br": // <w:br> — line break
				builder.WriteString("\n")
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				text := string(t)
				if len(text) > 0 {
					builder.WriteString(text)
					paragraphHasText = true
				}
			}
		}
	}

	return strings.TrimSpace(builder.String()), nil
}

// extractPDF extracts text from a PDF file using a pure-Go PDF reader.
func (s *TextExtractionService) extractPDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}

	var builder strings.Builder
	numPages := reader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Skip pages that fail to parse, continue with others.
			continue
		}
		if builder.Len() > 0 && len(text) > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(text)
	}

	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "", fmt.Errorf("no text content extracted from PDF")
	}
	return result, nil
}
