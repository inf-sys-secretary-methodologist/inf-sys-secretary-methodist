package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMessageOutput_Basic(t *testing.T) {
	now := time.Now()
	avatar := "https://example.com/avatar.jpg"
	msg := &entities.Message{
		ID:              1,
		ConversationID:  10,
		SenderID:        42,
		SenderName:      "Alice",
		SenderAvatarURL: &avatar,
		Type:            entities.MessageTypeText,
		Content:         "Hello world",
		IsEdited:        false,
		IsDeleted:       false,
		CreatedAt:       now,
	}

	output := ToMessageOutput(msg)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.ConversationID)
	assert.Equal(t, int64(42), output.SenderID)
	assert.Equal(t, "Alice", output.SenderName)
	assert.Equal(t, &avatar, output.SenderAvatar)
	assert.Equal(t, "text", output.Type)
	assert.Equal(t, "Hello world", output.Content)
	assert.False(t, output.IsEdited)
	assert.False(t, output.IsDeleted)
	assert.Empty(t, output.Attachments)
	assert.Nil(t, output.ReplyTo)
}

func TestToMessageOutput_Deleted(t *testing.T) {
	msg := &entities.Message{
		ID:        1,
		SenderID:  42,
		Type:      entities.MessageTypeText,
		Content:   "original content",
		IsDeleted: true,
		CreatedAt: time.Now(),
	}

	output := ToMessageOutput(msg)
	assert.Equal(t, "[Message deleted]", output.Content)
	assert.True(t, output.IsDeleted)
}

func TestToMessageOutput_WithReply(t *testing.T) {
	now := time.Now()
	reply := &entities.Message{
		ID:        2,
		SenderID:  43,
		Type:      entities.MessageTypeText,
		Content:   "Reply",
		CreatedAt: now,
	}

	msg := &entities.Message{
		ID:        1,
		SenderID:  42,
		Type:      entities.MessageTypeText,
		Content:   "Original",
		ReplyTo:   reply,
		CreatedAt: now,
	}

	output := ToMessageOutput(msg)

	require.NotNil(t, output.ReplyTo)
	assert.Equal(t, int64(2), output.ReplyTo.ID)
	assert.Equal(t, "Reply", output.ReplyTo.Content)
}

func TestToMessageOutput_WithAttachments(t *testing.T) {
	msg := &entities.Message{
		ID:       1,
		SenderID: 42,
		Type:     entities.MessageTypeFile,
		Content:  "See attachment",
		Attachments: []entities.Attachment{
			{
				ID:       1,
				FileName: "doc.pdf",
				FileSize: 1024,
				MimeType: "application/pdf",
				URL:      "https://example.com/doc.pdf",
			},
		},
		CreatedAt: time.Now(),
	}

	output := ToMessageOutput(msg)

	require.Len(t, output.Attachments, 1)
	assert.Equal(t, "doc.pdf", output.Attachments[0].FileName)
	assert.Equal(t, int64(1024), output.Attachments[0].FileSize)
	assert.Equal(t, "application/pdf", output.Attachments[0].MimeType)
	assert.Equal(t, "https://example.com/doc.pdf", output.Attachments[0].URL)
}

func TestToMessageOutput_Edited(t *testing.T) {
	editedAt := time.Now()
	msg := &entities.Message{
		ID:        1,
		SenderID:  42,
		Type:      entities.MessageTypeText,
		Content:   "Edited content",
		IsEdited:  true,
		EditedAt:  &editedAt,
		CreatedAt: time.Now(),
	}

	output := ToMessageOutput(msg)
	assert.True(t, output.IsEdited)
	assert.Equal(t, &editedAt, output.EditedAt)
}
