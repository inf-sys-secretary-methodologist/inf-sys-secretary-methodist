package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
)

// CreateDirectConversationInput represents input for creating a direct conversation.
type CreateDirectConversationInput struct {
	RecipientID int64 `json:"recipient_id" validate:"required,gt=0"`
}

// CreateGroupConversationInput represents input for creating a group conversation.
type CreateGroupConversationInput struct {
	Title          string  `json:"title" validate:"required,min=1,max=255"`
	Description    *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	ParticipantIDs []int64 `json:"participant_ids" validate:"required,min=1,dive,gt=0"`
}

// UpdateConversationInput represents input for updating a conversation.
type UpdateConversationInput struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	AvatarURL   *string `json:"avatar_url,omitempty" validate:"omitempty,url,max=500"`
}

// AddParticipantsInput represents input for adding participants to a conversation.
type AddParticipantsInput struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1,dive,gt=0"`
}

// UpdateParticipantInput represents input for updating a participant.
type UpdateParticipantInput struct {
	Role    *string `json:"role,omitempty" validate:"omitempty,oneof=member admin"`
	IsMuted *bool   `json:"is_muted,omitempty"`
}

// ConversationFilterInput represents query parameters for listing conversations.
type ConversationFilterInput struct {
	Type   *string `form:"type" validate:"omitempty,oneof=direct group"`
	Search *string `form:"search" validate:"omitempty,max=100"`
	Limit  int     `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int     `form:"offset" validate:"omitempty,min=0"`
}

// ParticipantOutput represents a participant in API responses.
type ParticipantOutput struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Name      string     `json:"name"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	Role      string     `json:"role"`
	IsMuted   bool       `json:"is_muted"`
	JoinedAt  time.Time  `json:"joined_at"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
}

// ConversationOutput represents a conversation in API responses.
type ConversationOutput struct {
	ID           int64               `json:"id"`
	Type         string              `json:"type"`
	Title        *string             `json:"title,omitempty"`
	Description  *string             `json:"description,omitempty"`
	AvatarURL    *string             `json:"avatar_url,omitempty"`
	CreatedBy    int64               `json:"created_by"`
	LastMessage  *MessageOutput      `json:"last_message,omitempty"`
	UnreadCount  int                 `json:"unread_count"`
	Participants []ParticipantOutput `json:"participants,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// ConversationListOutput represents a paginated list of conversations.
type ConversationListOutput struct {
	Conversations []ConversationOutput `json:"conversations"`
	Total         int64                `json:"total"`
	Limit         int                  `json:"limit"`
	Offset        int                  `json:"offset"`
}

// ToConversationOutput converts a conversation entity to output DTO.
func ToConversationOutput(c *entities.Conversation, currentUserID int64) ConversationOutput {
	output := ConversationOutput{
		ID:          c.ID,
		Type:        string(c.Type),
		Title:       c.Title,
		Description: c.Description,
		AvatarURL:   c.AvatarURL,
		CreatedBy:   c.CreatedBy,
		UnreadCount: c.UnreadCount,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// For direct conversations, use the other participant's name as title
	if c.IsDirectConversation() && output.Title == nil {
		if other := c.GetOtherParticipant(currentUserID); other != nil {
			output.Title = &other.UserName
			output.AvatarURL = other.UserAvatarURL
		}
	}

	if c.LastMessage != nil {
		msg := ToMessageOutput(c.LastMessage)
		output.LastMessage = &msg
	}

	output.Participants = make([]ParticipantOutput, 0, len(c.Participants))
	for _, p := range c.Participants {
		output.Participants = append(output.Participants, ToParticipantOutput(&p))
	}

	return output
}

// ToParticipantOutput converts a participant entity to output DTO.
func ToParticipantOutput(p *entities.Participant) ParticipantOutput {
	return ParticipantOutput{
		ID:        p.ID,
		UserID:    p.UserID,
		Name:      p.UserName,
		AvatarURL: p.UserAvatarURL,
		Role:      string(p.Role),
		IsMuted:   p.IsMuted,
		JoinedAt:  p.JoinedAt,
		LeftAt:    p.LeftAt,
	}
}
