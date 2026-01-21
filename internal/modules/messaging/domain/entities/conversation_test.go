package entities

import (
	"testing"
	"time"
)

func TestNewDirectConversation(t *testing.T) {
	creatorID := int64(1)
	recipientID := int64(2)

	conv := NewDirectConversation(creatorID, recipientID)

	if conv.Type != ConversationTypeDirect {
		t.Errorf("expected type %q, got %q", ConversationTypeDirect, conv.Type)
	}
	if conv.CreatedBy != creatorID {
		t.Errorf("expected created by %d, got %d", creatorID, conv.CreatedBy)
	}
	if len(conv.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(conv.Participants))
	}
	if conv.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewGroupConversation(t *testing.T) {
	creatorID := int64(1)
	title := "Test Group"
	participantIDs := []int64{2, 3, 4}

	conv := NewGroupConversation(creatorID, title, participantIDs)

	if conv.Type != ConversationTypeGroup {
		t.Errorf("expected type %q, got %q", ConversationTypeGroup, conv.Type)
	}
	if conv.Title == nil || *conv.Title != title {
		t.Errorf("expected title %q, got %v", title, conv.Title)
	}
	if conv.CreatedBy != creatorID {
		t.Errorf("expected created by %d, got %d", creatorID, conv.CreatedBy)
	}
	// Creator + 3 participants = 4
	if len(conv.Participants) != 4 {
		t.Errorf("expected 4 participants, got %d", len(conv.Participants))
	}
	// First participant should be the creator with admin role
	if conv.Participants[0].UserID != creatorID {
		t.Errorf("expected first participant to be creator %d, got %d", creatorID, conv.Participants[0].UserID)
	}
	if conv.Participants[0].Role != ParticipantRoleAdmin {
		t.Errorf("expected creator role %q, got %q", ParticipantRoleAdmin, conv.Participants[0].Role)
	}
}

func TestNewGroupConversation_CreatorInList(t *testing.T) {
	creatorID := int64(1)
	title := "Test Group"
	// Creator is also in the participant list
	participantIDs := []int64{1, 2, 3}

	conv := NewGroupConversation(creatorID, title, participantIDs)

	// Should only have 3 participants (creator not duplicated)
	if len(conv.Participants) != 3 {
		t.Errorf("expected 3 participants, got %d", len(conv.Participants))
	}
}

func TestConversation_IsDirectConversation(t *testing.T) {
	direct := NewDirectConversation(1, 2)
	group := NewGroupConversation(1, "Group", []int64{2})

	if !direct.IsDirectConversation() {
		t.Error("expected direct conversation to return true")
	}
	if group.IsDirectConversation() {
		t.Error("expected group conversation to return false")
	}
}

func TestConversation_IsGroupConversation(t *testing.T) {
	direct := NewDirectConversation(1, 2)
	group := NewGroupConversation(1, "Group", []int64{2})

	if direct.IsGroupConversation() {
		t.Error("expected direct conversation to return false")
	}
	if !group.IsGroupConversation() {
		t.Error("expected group conversation to return true")
	}
}

func TestConversation_HasParticipant(t *testing.T) {
	conv := NewDirectConversation(1, 2)

	if !conv.HasParticipant(1) {
		t.Error("expected creator to be a participant")
	}
	if !conv.HasParticipant(2) {
		t.Error("expected recipient to be a participant")
	}
	if conv.HasParticipant(99) {
		t.Error("expected non-participant to return false")
	}
}

func TestConversation_HasParticipant_LeftUser(t *testing.T) {
	conv := NewGroupConversation(1, "Group", []int64{2, 3})
	now := time.Now()
	conv.Participants[1].LeftAt = &now

	if conv.HasParticipant(2) {
		t.Error("expected left user to not be counted as participant")
	}
}

func TestConversation_GetParticipant(t *testing.T) {
	conv := NewDirectConversation(1, 2)

	p := conv.GetParticipant(1)
	if p == nil {
		t.Fatal("expected to find participant")
	}
	if p.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", p.UserID)
	}

	p = conv.GetParticipant(99)
	if p != nil {
		t.Error("expected nil for non-participant")
	}
}

func TestConversation_IsAdmin(t *testing.T) {
	conv := NewGroupConversation(1, "Group", []int64{2})

	if !conv.IsAdmin(1) {
		t.Error("expected creator to be admin")
	}
	if conv.IsAdmin(2) {
		t.Error("expected member to not be admin")
	}
	if conv.IsAdmin(99) {
		t.Error("expected non-participant to not be admin")
	}
}

func TestConversation_GetOtherParticipant(t *testing.T) {
	conv := NewDirectConversation(1, 2)

	other := conv.GetOtherParticipant(1)
	if other == nil {
		t.Fatal("expected to find other participant")
	}
	if other.UserID != 2 {
		t.Errorf("expected user ID 2, got %d", other.UserID)
	}

	other = conv.GetOtherParticipant(2)
	if other == nil {
		t.Fatal("expected to find other participant")
	}
	if other.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", other.UserID)
	}
}

func TestConversation_GetOtherParticipant_GroupConversation(t *testing.T) {
	conv := NewGroupConversation(1, "Group", []int64{2, 3})

	other := conv.GetOtherParticipant(1)
	if other != nil {
		t.Error("expected nil for group conversation")
	}
}

func TestConversation_ActiveParticipantCount(t *testing.T) {
	conv := NewGroupConversation(1, "Group", []int64{2, 3, 4})

	if conv.ActiveParticipantCount() != 4 {
		t.Errorf("expected 4 active participants, got %d", conv.ActiveParticipantCount())
	}

	// Mark one as left
	now := time.Now()
	conv.Participants[1].LeftAt = &now

	if conv.ActiveParticipantCount() != 3 {
		t.Errorf("expected 3 active participants, got %d", conv.ActiveParticipantCount())
	}
}

func TestConversationTypeConstants(t *testing.T) {
	if ConversationTypeDirect != "direct" {
		t.Errorf("expected %q, got %q", "direct", ConversationTypeDirect)
	}
	if ConversationTypeGroup != "group" {
		t.Errorf("expected %q, got %q", "group", ConversationTypeGroup)
	}
}

func TestParticipantRoleConstants(t *testing.T) {
	if ParticipantRoleMember != "member" {
		t.Errorf("expected %q, got %q", "member", ParticipantRoleMember)
	}
	if ParticipantRoleAdmin != "admin" {
		t.Errorf("expected %q, got %q", "admin", ParticipantRoleAdmin)
	}
}
