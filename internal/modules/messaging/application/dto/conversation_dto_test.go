package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToParticipantOutput(t *testing.T) {
	now := time.Now()
	avatar := "https://example.com/avatar.jpg"
	p := &entities.Participant{
		ID:            1,
		UserID:        42,
		UserName:      "John",
		UserAvatarURL: &avatar,
		Role:          entities.ParticipantRoleAdmin,
		IsMuted:       true,
		JoinedAt:      now,
	}

	output := ToParticipantOutput(p)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(42), output.UserID)
	assert.Equal(t, "John", output.Name)
	assert.Equal(t, &avatar, output.AvatarURL)
	assert.Equal(t, "admin", output.Role)
	assert.True(t, output.IsMuted)
	assert.Equal(t, now, output.JoinedAt)
	assert.Nil(t, output.LeftAt)
}

func TestToConversationOutput_GroupConversation(t *testing.T) {
	now := time.Now()
	title := "Team Chat"
	desc := "Team discussion"

	conv := &entities.Conversation{
		ID:          1,
		Type:        entities.ConversationTypeGroup,
		Title:       &title,
		Description: &desc,
		CreatedBy:   42,
		UnreadCount: 5,
		Participants: []entities.Participant{
			{ID: 1, UserID: 42, UserName: "Alice", Role: entities.ParticipantRoleAdmin, JoinedAt: now},
			{ID: 2, UserID: 43, UserName: "Bob", Role: entities.ParticipantRoleMember, JoinedAt: now},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	output := ToConversationOutput(conv, 42)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "group", output.Type)
	assert.Equal(t, &title, output.Title)
	assert.Equal(t, &desc, output.Description)
	assert.Equal(t, int64(42), output.CreatedBy)
	assert.Equal(t, 5, output.UnreadCount)
	require.Len(t, output.Participants, 2)
	assert.Nil(t, output.LastMessage)
}

func TestToConversationOutput_WithLastMessage(t *testing.T) {
	now := time.Now()
	title := "Chat"
	lastMsg := &entities.Message{
		ID:             10,
		ConversationID: 1,
		SenderID:       42,
		SenderName:     "Alice",
		Type:           entities.MessageTypeText,
		Content:        "Hello!",
		CreatedAt:      now,
	}

	conv := &entities.Conversation{
		ID:          1,
		Type:        entities.ConversationTypeGroup,
		Title:       &title,
		CreatedBy:   42,
		LastMessage: lastMsg,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	output := ToConversationOutput(conv, 42)

	require.NotNil(t, output.LastMessage)
	assert.Equal(t, int64(10), output.LastMessage.ID)
	assert.Equal(t, "Hello!", output.LastMessage.Content)
}
