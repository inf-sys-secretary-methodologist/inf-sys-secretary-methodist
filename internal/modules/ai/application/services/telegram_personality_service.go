// Package services contains application services for the AI module.
package services

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	domainServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

// TelegramPersonalityService decorates TelegramService with Metodych personality
type TelegramPersonalityService struct {
	telegramService    domainServices.TelegramService
	personalityProvider PersonalityProvider
}

// NewTelegramPersonalityService creates a new TelegramPersonalityService
func NewTelegramPersonalityService(
	telegramService domainServices.TelegramService,
	personalityProvider PersonalityProvider,
) *TelegramPersonalityService {
	return &TelegramPersonalityService{
		telegramService:    telegramService,
		personalityProvider: personalityProvider,
	}
}

// SendPersonalizedNotification sends a notification with personality formatting
func (s *TelegramPersonalityService) SendPersonalizedNotification(
	ctx context.Context,
	chatID string,
	notifType string,
	title string,
	message string,
	mood entities.MoodContext,
) error {
	formattedMessage := s.personalityProvider.FormatNotification(notifType, title, message, mood)

	req := &domainServices.SendTelegramMessageRequest{
		ChatID:    chatID,
		Text:      formattedMessage,
		ParseMode: "HTML",
	}
	return s.telegramService.SendMessage(ctx, req)
}

// SendFactMessage sends a fun fact through Telegram
func (s *TelegramPersonalityService) SendFactMessage(ctx context.Context, chatID string, fact string, mood entities.MoodContext) error {
	emoji := "💡"
	switch mood.State {
	case entities.MoodHappy:
		emoji = "🎓"
	case entities.MoodRelaxed:
		emoji = "☕"
	case entities.MoodInspired:
		emoji = "✨"
	}

	text := fmt.Sprintf("%s <b>Факт от Методыча</b>\n\n%s\n\n<i>— Ваш Методыч, 40 лет в образовании</i>", emoji, fact)

	req := &domainServices.SendTelegramMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}
	return s.telegramService.SendMessage(ctx, req)
}

// SendMoodMessage sends the current mood status through Telegram
func (s *TelegramPersonalityService) SendMoodMessage(ctx context.Context, chatID string, mood entities.MoodContext) error {
	comment := s.personalityProvider.GetMoodComment(mood)
	greeting := s.personalityProvider.GetGreeting(mood.TimeOfDay)

	text := fmt.Sprintf("🎭 <b>Настроение Методыча</b>\n\n%s\n\n%s\n\n", greeting, comment)

	if mood.OverdueDocuments > 0 {
		text += fmt.Sprintf("📋 Новых документов: %d\n", mood.OverdueDocuments)
	}
	if mood.AtRiskStudents > 0 {
		text += fmt.Sprintf("⚠️ Студентов в зоне риска: %d\n", mood.AtRiskStudents)
	}

	text += fmt.Sprintf("\n<i>Состояние: %s</i>", string(mood.State))

	req := &domainServices.SendTelegramMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}
	return s.telegramService.SendMessage(ctx, req)
}
