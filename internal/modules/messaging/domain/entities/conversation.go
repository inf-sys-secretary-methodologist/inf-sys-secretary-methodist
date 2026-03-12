package entities

import (
	"errors"
	"time"
)

// ConversationType represents the type of conversation.
type ConversationType string

// ConversationType values.
const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

// Validation errors.
var (
	ErrConversationNotFound     = errors.New("conversation not found")
	ErrAlreadyParticipant       = errors.New("user is already a participant")
	ErrNotParticipant           = errors.New("user is not a participant")
	ErrCannotRemoveLastAdmin    = errors.New("cannot remove the last admin")
	ErrDirectConversationExists = errors.New("direct conversation already exists")
	ErrInvalidConversationType  = errors.New("invalid conversation type")
	ErrCannotAddToDirectChat    = errors.New("cannot add participants to direct chat")
	ErrCannotLeaveDirectChat    = errors.New("cannot leave direct chat")
)

// Conversation represents a chat conversation (direct or group).
type Conversation struct {
	ID           int64            `json:"id"`
	Type         ConversationType `json:"type"`
	Title        *string          `json:"title,omitempty"`
	Description  *string          `json:"description,omitempty"`
	AvatarURL    *string          `json:"avatar_url,omitempty"`
	CreatedBy    int64            `json:"created_by"`
	LastMessage  *Message         `json:"last_message,omitempty"`
	UnreadCount  int              `json:"unread_count"`
	Participants []Participant    `json:"participants,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// ParticipantRole represents the role of a participant in a conversation.
type ParticipantRole string

// ParticipantRole values.
const (
	ParticipantRoleMember ParticipantRole = "member"
	ParticipantRoleAdmin  ParticipantRole = "admin"
)

// Participant represents a user's participation in a conversation.
type Participant struct {
	ID             int64           `json:"id"`
	ConversationID int64           `json:"conversation_id"`
	UserID         int64           `json:"user_id"`
	Role           ParticipantRole `json:"role"`
	LastReadAt     *time.Time      `json:"last_read_at,omitempty"`
	IsMuted        bool            `json:"is_muted"`
	JoinedAt       time.Time       `json:"joined_at"`
	LeftAt         *time.Time      `json:"left_at,omitempty"`
	// Joined user info
	UserName      string  `json:"user_name,omitempty"`
	UserAvatarURL *string `json:"user_avatar_url,omitempty"`
}

// NewDirectConversation creates a new direct conversation between two users.
func NewDirectConversation(creatorID, recipientID int64) *Conversation {
	now := time.Now()
	return &Conversation{
		Type:      ConversationTypeDirect,
		CreatedBy: creatorID,
		Participants: []Participant{
			{UserID: creatorID, Role: ParticipantRoleMember, JoinedAt: now},
			{UserID: recipientID, Role: ParticipantRoleMember, JoinedAt: now},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewGroupConversation creates a new group conversation.
func NewGroupConversation(creatorID int64, title string, participantIDs []int64) *Conversation {
	now := time.Now()
	participants := make([]Participant, 0, len(participantIDs)+1)

	// Creator is always an admin
	participants = append(participants, Participant{
		UserID:   creatorID,
		Role:     ParticipantRoleAdmin,
		JoinedAt: now,
	})

	// Add other participants as members
	for _, userID := range participantIDs {
		if userID != creatorID {
			participants = append(participants, Participant{
				UserID:   userID,
				Role:     ParticipantRoleMember,
				JoinedAt: now,
			})
		}
	}

	return &Conversation{
		Type:         ConversationTypeGroup,
		Title:        &title,
		CreatedBy:    creatorID,
		Participants: participants,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// IsDirectConversation checks if this is a direct (1:1) conversation.
func (c *Conversation) IsDirectConversation() bool {
	return c.Type == ConversationTypeDirect
}

// IsGroupConversation checks if this is a group conversation.
func (c *Conversation) IsGroupConversation() bool {
	return c.Type == ConversationTypeGroup
}

// HasParticipant checks if a user is a participant in the conversation.
func (c *Conversation) HasParticipant(userID int64) bool {
	for _, p := range c.Participants {
		if p.UserID == userID && p.LeftAt == nil {
			return true
		}
	}
	return false
}

// GetParticipant returns the participant with the given user ID.
func (c *Conversation) GetParticipant(userID int64) *Participant {
	for i := range c.Participants {
		if c.Participants[i].UserID == userID && c.Participants[i].LeftAt == nil {
			return &c.Participants[i]
		}
	}
	return nil
}

// IsAdmin checks if a user is an admin of the conversation.
func (c *Conversation) IsAdmin(userID int64) bool {
	p := c.GetParticipant(userID)
	return p != nil && p.Role == ParticipantRoleAdmin
}

// GetOtherParticipant returns the other participant in a direct conversation.
func (c *Conversation) GetOtherParticipant(userID int64) *Participant {
	if !c.IsDirectConversation() {
		return nil
	}
	for i := range c.Participants {
		if c.Participants[i].UserID != userID && c.Participants[i].LeftAt == nil {
			return &c.Participants[i]
		}
	}
	return nil
}

// ActiveParticipantCount returns the number of active participants.
func (c *Conversation) ActiveParticipantCount() int {
	count := 0
	for _, p := range c.Participants {
		if p.LeftAt == nil {
			count++
		}
	}
	return count
}

// ConversationFilter represents filters for listing conversations.
type ConversationFilter struct {
	UserID int64
	Type   *ConversationType
	Search *string
	Limit  int
	Offset int
}
