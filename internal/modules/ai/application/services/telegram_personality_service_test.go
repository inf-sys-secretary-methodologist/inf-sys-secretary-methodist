package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	domainServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

// --- Mock TelegramService ---

type mockTelegramService struct {
	lastReq *domainServices.SendTelegramMessageRequest
	err     error
}

func (m *mockTelegramService) SendMessage(_ context.Context, req *domainServices.SendTelegramMessageRequest) error {
	m.lastReq = req
	return m.err
}

func (m *mockTelegramService) SendNotification(_ context.Context, _ string, _, _ string, _ string) error {
	return m.err
}

// --- Mock PersonalityProvider ---

type mockPersonalityProvider struct {
	systemPrompt string
	ragContext   string
	greeting     string
	moodComment  string
	notification string
}

func (m *mockPersonalityProvider) BuildSystemPrompt(_ entities.MoodContext) string {
	return m.systemPrompt
}

func (m *mockPersonalityProvider) FormatRAGContext(_ []entities.ChunkWithScore) string {
	return m.ragContext
}

func (m *mockPersonalityProvider) GetGreeting(_ string) string {
	return m.greeting
}

func (m *mockPersonalityProvider) GetMoodComment(_ entities.MoodContext) string {
	return m.moodComment
}

func (m *mockPersonalityProvider) FormatNotification(_, _, _ string, _ entities.MoodContext) string {
	return m.notification
}

// --- Tests ---

func TestNewTelegramPersonalityService(t *testing.T) {
	tg := &mockTelegramService{}
	pp := &mockPersonalityProvider{}

	svc := NewTelegramPersonalityService(tg, pp)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.telegramService != tg {
		t.Error("telegramService not set correctly")
	}
	if svc.personalityProvider != pp {
		t.Error("personalityProvider not set correctly")
	}
}

func TestSendPersonalizedNotification_Success(t *testing.T) {
	tg := &mockTelegramService{}
	pp := &mockPersonalityProvider{notification: "formatted notification"}
	svc := NewTelegramPersonalityService(tg, pp)

	mood := entities.MoodContext{State: entities.MoodHappy}
	err := svc.SendPersonalizedNotification(context.Background(), "chat123", "info", "Title", "Body", mood)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tg.lastReq == nil {
		t.Fatal("expected request to be sent")
	}
	if tg.lastReq.ChatID != "chat123" {
		t.Errorf("expected chat_id=chat123, got %q", tg.lastReq.ChatID)
	}
	if tg.lastReq.Text != "formatted notification" {
		t.Errorf("expected formatted notification, got %q", tg.lastReq.Text)
	}
	if tg.lastReq.ParseMode != "HTML" {
		t.Errorf("expected parse_mode=HTML, got %q", tg.lastReq.ParseMode)
	}
}

func TestSendPersonalizedNotification_Error(t *testing.T) {
	tg := &mockTelegramService{err: errors.New("send failed")}
	pp := &mockPersonalityProvider{notification: "text"}
	svc := NewTelegramPersonalityService(tg, pp)

	err := svc.SendPersonalizedNotification(context.Background(), "chat1", "info", "T", "M", entities.MoodContext{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "send failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendFactMessage_MoodEmojis(t *testing.T) {
	tests := []struct {
		name     string
		mood     entities.MoodState
		expected string
	}{
		{"happy", entities.MoodHappy, "🎓"},
		{"relaxed", entities.MoodRelaxed, "☕"},
		{"inspired", entities.MoodInspired, "✨"},
		{"default", entities.MoodContent, "💡"},
		{"stressed", entities.MoodStressed, "💡"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &mockTelegramService{}
			pp := &mockPersonalityProvider{}
			svc := NewTelegramPersonalityService(tg, pp)

			mood := entities.MoodContext{State: tt.mood}
			err := svc.SendFactMessage(context.Background(), "chat1", "Fun fact!", mood)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tg.lastReq == nil {
				t.Fatal("expected request to be sent")
			}
			if !strings.HasPrefix(tg.lastReq.Text, tt.expected) {
				t.Errorf("expected text to start with %s, got %q", tt.expected, tg.lastReq.Text)
			}
			if !strings.Contains(tg.lastReq.Text, "Fun fact!") {
				t.Error("expected fact text in message")
			}
			if !strings.Contains(tg.lastReq.Text, "Факт от Методыча") {
				t.Error("expected title in message")
			}
			if tg.lastReq.ParseMode != "HTML" {
				t.Errorf("expected parse_mode=HTML, got %q", tg.lastReq.ParseMode)
			}
		})
	}
}

func TestSendFactMessage_Error(t *testing.T) {
	tg := &mockTelegramService{err: errors.New("fail")}
	pp := &mockPersonalityProvider{}
	svc := NewTelegramPersonalityService(tg, pp)

	err := svc.SendFactMessage(context.Background(), "c", "fact", entities.MoodContext{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendMoodMessage_WithOverdueAndAtRisk(t *testing.T) {
	tg := &mockTelegramService{}
	pp := &mockPersonalityProvider{greeting: "Hello!", moodComment: "Feeling great"}
	svc := NewTelegramPersonalityService(tg, pp)

	mood := entities.MoodContext{
		State:            entities.MoodWorried,
		TimeOfDay:        "morning",
		OverdueDocuments: 5,
		AtRiskStudents:   3,
	}
	err := svc.SendMoodMessage(context.Background(), "chat1", mood)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tg.lastReq == nil {
		t.Fatal("expected request to be sent")
	}

	text := tg.lastReq.Text
	if !strings.Contains(text, "Настроение Методыча") {
		t.Error("expected mood title in message")
	}
	if !strings.Contains(text, "Hello!") {
		t.Error("expected greeting in message")
	}
	if !strings.Contains(text, "Feeling great") {
		t.Error("expected mood comment in message")
	}
	if !strings.Contains(text, "Новых документов: 5") {
		t.Error("expected overdue documents count")
	}
	if !strings.Contains(text, "Студентов в зоне риска: 3") {
		t.Error("expected at-risk students count")
	}
	if !strings.Contains(text, string(entities.MoodWorried)) {
		t.Error("expected mood state in message")
	}
}

func TestSendMoodMessage_NoOverdueOrAtRisk(t *testing.T) {
	tg := &mockTelegramService{}
	pp := &mockPersonalityProvider{greeting: "Hi", moodComment: "OK"}
	svc := NewTelegramPersonalityService(tg, pp)

	mood := entities.MoodContext{
		State:            entities.MoodHappy,
		OverdueDocuments: 0,
		AtRiskStudents:   0,
	}
	err := svc.SendMoodMessage(context.Background(), "chat1", mood)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := tg.lastReq.Text
	if strings.Contains(text, "Новых документов") {
		t.Error("should not contain overdue documents when count is 0")
	}
	if strings.Contains(text, "Студентов в зоне риска") {
		t.Error("should not contain at-risk students when count is 0")
	}
}

func TestSendMoodMessage_Error(t *testing.T) {
	tg := &mockTelegramService{err: errors.New("fail")}
	pp := &mockPersonalityProvider{greeting: "Hi", moodComment: "OK"}
	svc := NewTelegramPersonalityService(tg, pp)

	err := svc.SendMoodMessage(context.Background(), "c", entities.MoodContext{})
	if err == nil {
		t.Fatal("expected error")
	}
}
