package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToDocumentOutput(t *testing.T) {
	now := time.Now()
	authorName := "Author"
	recipientName := "Recipient"
	fileName := "doc.pdf"
	filePath := "/files/doc.pdf"
	fileSize := int64(2048)
	mimeType := "application/pdf"
	subject := "Subject"
	content := "Content"
	catID := int64(5)

	doc := &entities.Document{
		ID:             1,
		DocumentTypeID: 2,
		CategoryID:     &catID,
		Title:          "Test Doc",
		Subject:        &subject,
		Content:        &content,
		AuthorID:       10,
		AuthorName:     &authorName,
		RecipientID:    ptrInt64(20),
		RecipientName:  &recipientName,
		Status:         entities.DocumentStatusDraft,
		FileName:       &fileName,
		FilePath:       &filePath,
		FileSize:       &fileSize,
		MimeType:       &mimeType,
		Version:        3,
		Importance:     entities.ImportanceNormal,
		IsPublic:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	output := ToDocumentOutput(doc)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(2), output.DocumentTypeID)
	assert.Equal(t, &catID, output.CategoryID)
	assert.Equal(t, "Test Doc", output.Title)
	assert.Equal(t, &subject, output.Subject)
	assert.Equal(t, &content, output.Content)
	assert.Equal(t, int64(10), output.AuthorID)
	assert.Equal(t, "Author", output.AuthorName)
	assert.Equal(t, &recipientName, output.RecipientName)
	assert.Equal(t, "draft", output.Status)
	assert.True(t, output.HasFile)
	assert.Equal(t, 3, output.Version)
	assert.Equal(t, "normal", output.Importance)
	assert.True(t, output.IsPublic)
}

func TestToDocumentOutput_NoFile(t *testing.T) {
	doc := &entities.Document{
		ID:         1,
		AuthorID:   1,
		Title:      "No File",
		Status:     entities.DocumentStatusDraft,
		Importance: entities.ImportanceLow,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	output := ToDocumentOutput(doc)
	assert.False(t, output.HasFile)
	assert.Nil(t, output.FileName)
}

func ptrInt64(v int64) *int64 {
	return &v
}
