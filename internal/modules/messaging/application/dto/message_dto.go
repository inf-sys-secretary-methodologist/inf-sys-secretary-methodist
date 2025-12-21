package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
)

// AttachmentInput represents an attachment to include with a message.
type AttachmentInput struct {
	FileID   int64  `json:"file_id" validate:"required,gt=0"`
	FileName string `json:"file_name" validate:"required,min=1,max=500"`
	FileSize int64  `json:"file_size" validate:"required,gt=0"`
	MimeType string `json:"mime_type" validate:"required,min=1,max=100"`
	URL      string `json:"url" validate:"required,url"`
}

// SendMessageInput represents input for sending a message.
type SendMessageInput struct {
	Content     string            `json:"content" validate:"required,min=1,max=10000"`
	ReplyToID   *int64            `json:"reply_to_id,omitempty" validate:"omitempty,gt=0"`
	Type        *string           `json:"type,omitempty" validate:"omitempty,oneof=text image file"`
	Attachments []AttachmentInput `json:"attachments,omitempty" validate:"omitempty,max=10,dive"`
}

// EditMessageInput represents input for editing a message.
type EditMessageInput struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

// MessageFilterInput represents query parameters for listing messages.
type MessageFilterInput struct {
	BeforeID *int64  `form:"before_id" validate:"omitempty,gt=0"`
	AfterID  *int64  `form:"after_id" validate:"omitempty,gt=0"`
	Search   *string `form:"search" validate:"omitempty,max=100"`
	Limit    int     `form:"limit" validate:"omitempty,min=1,max=100"`
}

// AttachmentOutput represents a message attachment in API responses.
type AttachmentOutput struct {
	ID       int64  `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}

// MessageOutput represents a message in API responses.
type MessageOutput struct {
	ID             int64              `json:"id"`
	ConversationID int64              `json:"conversation_id"`
	SenderID       int64              `json:"sender_id"`
	SenderName     string             `json:"sender_name"`
	SenderAvatar   *string            `json:"sender_avatar,omitempty"`
	Type           string             `json:"type"`
	Content        string             `json:"content"`
	ReplyTo        *MessageOutput     `json:"reply_to,omitempty"`
	Attachments    []AttachmentOutput `json:"attachments,omitempty"`
	IsEdited       bool               `json:"is_edited"`
	EditedAt       *time.Time         `json:"edited_at,omitempty"`
	IsDeleted      bool               `json:"is_deleted"`
	CreatedAt      time.Time          `json:"created_at"`
}

// MessageListOutput represents a list of messages.
type MessageListOutput struct {
	Messages []MessageOutput `json:"messages"`
	HasMore  bool            `json:"has_more"`
}

// ToMessageOutput converts a message entity to output DTO.
func ToMessageOutput(m *entities.Message) MessageOutput {
	output := MessageOutput{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		SenderName:     m.SenderName,
		SenderAvatar:   m.SenderAvatarURL,
		Type:           string(m.Type),
		Content:        m.Content,
		IsEdited:       m.IsEdited,
		EditedAt:       m.EditedAt,
		IsDeleted:      m.IsDeleted,
		CreatedAt:      m.CreatedAt,
	}

	// Handle deleted messages
	if m.IsDeleted {
		output.Content = "[Message deleted]"
	}

	// Convert reply
	if m.ReplyTo != nil {
		reply := ToMessageOutput(m.ReplyTo)
		output.ReplyTo = &reply
	}

	// Convert attachments
	output.Attachments = make([]AttachmentOutput, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		output.Attachments = append(output.Attachments, AttachmentOutput{
			ID:       a.ID,
			FileName: a.FileName,
			FileSize: a.FileSize,
			MimeType: a.MimeType,
			URL:      a.URL,
		})
	}

	return output
}

// TypingInput represents input for typing indicator.
type TypingInput struct {
	IsTyping bool `json:"is_typing"`
}

// MarkReadInput represents input for marking messages as read.
type MarkReadInput struct {
	MessageID int64 `json:"message_id" validate:"required,gt=0"`
}

// SearchMessagesInput represents query parameters for searching messages.
type SearchMessagesInput struct {
	Query  string `form:"q" validate:"required,min=1,max=100"`
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
}

// SearchMessagesOutput represents a search result list.
type SearchMessagesOutput struct {
	Messages []MessageOutput `json:"messages"`
	Total    int64           `json:"total"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}
