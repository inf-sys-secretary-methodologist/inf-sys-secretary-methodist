// Package http contains HTTP handlers for the notifications module.
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	aiServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	aiUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
	domainServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/telegram"
)

// TelegramWebhookHandler handles Telegram webhook requests
type TelegramWebhookHandler struct {
	verificationService *services.TelegramVerificationService
	telegramService     domainServices.TelegramService // Composio for sending messages
	webhookSecret       string
	logger              *slog.Logger
	personalityService  *aiServices.TelegramPersonalityService
	moodUseCase         *aiUsecases.MoodUseCase
}

// NewTelegramWebhookHandler creates a new Telegram webhook handler
func NewTelegramWebhookHandler(
	verificationService *services.TelegramVerificationService,
	telegramService domainServices.TelegramService,
	webhookSecret string,
	logger *slog.Logger,
	personalityService *aiServices.TelegramPersonalityService,
	moodUseCase *aiUsecases.MoodUseCase,
) *TelegramWebhookHandler {
	return &TelegramWebhookHandler{
		verificationService: verificationService,
		telegramService:     telegramService,
		webhookSecret:       webhookSecret,
		logger:              logger,
		personalityService:  personalityService,
		moodUseCase:         moodUseCase,
	}
}

// HandleWebhook handles incoming Telegram updates
// @Summary Handle Telegram webhook
// @Description Receives and processes updates from Telegram Bot API
// @Tags telegram
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/telegram/webhook [post]
func (h *TelegramWebhookHandler) HandleWebhook(c *gin.Context) {
	// Verify webhook secret if configured
	if h.webhookSecret != "" {
		secretHeader := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
		if secretHeader != h.webhookSecret {
			h.logger.Warn("invalid webhook secret",
				"remote_addr", c.ClientIP(),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid secret"})
			return
		}
	}

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Parse the update
	var update telegram.Update
	if err := json.Unmarshal(body, &update); err != nil {
		h.logger.Error("failed to parse update", "error", err, "body", string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update"})
		return
	}

	// Process the update
	go h.ProcessUpdate(&update)

	// Always return 200 OK to Telegram
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ProcessUpdate processes a Telegram update (exported for polling mode)
func (h *TelegramWebhookHandler) ProcessUpdate(update *telegram.Update) {
	if update.Message == nil {
		return
	}

	message := update.Message
	if message.Chat == nil {
		return
	}

	chatID := message.Chat.ID
	text := strings.TrimSpace(message.Text)

	// Handle /start command with verification code
	if strings.HasPrefix(text, "/start") {
		h.handleStartCommand(chatID, text, message.Chat)
		return
	}

	// Handle plain verification code (8 hex characters)
	if len(text) == 8 && isHexString(text) {
		h.handleVerificationCode(chatID, text, message.Chat)
		return
	}

	// Handle /help command
	if text == "/help" {
		h.sendHelpMessage(chatID)
		return
	}

	// Handle /status command
	if text == "/status" {
		h.handleStatusCommand(chatID)
		return
	}

	// Handle /fact command
	if text == "/fact" {
		h.handleFactCommand(chatID)
		return
	}

	// Handle /mood command
	if text == "/mood" {
		h.handleMoodCommand(chatID)
		return
	}

	// Unknown command or message
	h.sendUnknownCommandMessage(chatID)
}

// handleStartCommand handles the /start command
func (h *TelegramWebhookHandler) handleStartCommand(chatID int64, text string, chat *telegram.Chat) {
	parts := strings.Fields(text)

	// If just /start without code, send welcome message
	if len(parts) == 1 {
		h.sendStartMessage(chatID)
		return
	}

	// Extract the verification code
	code := parts[1]
	h.handleVerificationCode(chatID, code, chat)
}

// handleVerificationCode handles a verification code
func (h *TelegramWebhookHandler) handleVerificationCode(chatID int64, code string, chat *telegram.Chat) {
	ctx := context.Background()

	req := &services.VerifyCodeRequest{
		Code:              code,
		TelegramChatID:    chatID,
		TelegramUsername:  chat.Username,
		TelegramFirstName: chat.FirstName,
	}

	result, err := h.verificationService.VerifyCode(ctx, req)
	if err != nil {
		h.logger.Error("failed to verify code",
			"error", err,
			"chat_id", chatID,
			"code", code,
		)
		h.sendMessage(chatID, "Произошла ошибка при проверке кода. Пожалуйста, попробуйте позже.")
		return
	}

	if !result.Success {
		h.sendMessage(chatID, result.Message)
		return
	}

	// Send welcome message
	firstName := chat.FirstName
	if firstName == "" {
		firstName = "друг"
	}

	if err := h.verificationService.SendWelcomeMessage(ctx, chatID, firstName); err != nil {
		h.logger.Error("failed to send welcome message", "error", err, "chat_id", chatID)
	}
}

// handleStatusCommand handles the /status command
func (h *TelegramWebhookHandler) handleStatusCommand(chatID int64) {
	ctx := context.Background()

	// Check if this chat is linked to a user
	conn, err := h.verificationService.GetConnection(ctx, 0) // We need to look up by chat ID
	if err != nil {
		h.logger.Error("failed to get connection", "error", err, "chat_id", chatID)
		h.sendMessage(chatID, "Произошла ошибка. Пожалуйста, попробуйте позже.")
		return
	}

	if conn == nil {
		h.sendMessage(chatID, "❌ Этот чат не привязан к аккаунту.\n\nДля привязки получите код в настройках уведомлений.")
		return
	}

	status := "✅ Активно"
	if !conn.IsActive {
		status = "⏸ Приостановлено"
	}

	message := "📊 <b>Статус подключения</b>\n\n" +
		"Состояние: " + status + "\n" +
		"Подключено: " + conn.ConnectedAt.Format("02.01.2006 15:04")

	h.sendHTMLMessage(chatID, message)
}

// sendStartMessage sends the start message
func (h *TelegramWebhookHandler) sendStartMessage(chatID int64) {
	message := "👋 <b>Здравствуйте! Я — Методыч!</b>\n\n" +
		"Ветеран-методист с 40-летним стажем, теперь ещё и в цифровом формате! 😄\n\n" +
		"Я отправляю уведомления из системы управления документами и делюсь интересными фактами об образовании.\n\n" +
		"<b>Как привязать аккаунт:</b>\n" +
		"1. Зайдите в «Настройки» → «Уведомления»\n" +
		"2. Нажмите «Привязать Telegram»\n" +
		"3. Отправьте полученный код сюда\n\n" +
		"Или перейдите по ссылке с кодом из настроек.\n\n" +
		"<i>— Ваш Методыч</i>"
	h.sendHTMLMessage(chatID, message)
}

// sendHelpMessage sends the help message
func (h *TelegramWebhookHandler) sendHelpMessage(chatID int64) {
	message := "📚 <b>Справка от Методыча</b>\n\n" +
		"За 40 лет я выучил все команды наизусть:\n\n" +
		"/start — Познакомиться с Методычем\n" +
		"/status — Проверить статус подключения\n" +
		"/mood — Узнать моё настроение 🎭\n" +
		"/fact — Получить интересный факт 💡\n" +
		"/help — Эта справка\n\n" +
		"<b>Привязка аккаунта:</b>\n" +
		"Отправьте код верификации из настроек уведомлений.\n\n" +
		"<i>— Методыч всегда на связи!</i>"
	h.sendHTMLMessage(chatID, message)
}

// sendUnknownCommandMessage sends a message for unknown commands
func (h *TelegramWebhookHandler) sendUnknownCommandMessage(chatID int64) {
	message := "🤔 Хм, за 40 лет такой команды не встречал!\n\nПопробуйте /help — там всё расписано."
	h.sendMessage(chatID, message)
}

// sendMessage sends a plain text message via Composio
func (h *TelegramWebhookHandler) sendMessage(chatID int64, text string) {
	ctx := context.Background()
	chatIDStr := strconv.FormatInt(chatID, 10)

	req := &domainServices.SendTelegramMessageRequest{
		ChatID: chatIDStr,
		Text:   text,
	}
	if err := h.telegramService.SendMessage(ctx, req); err != nil {
		h.logger.Error("failed to send message", "error", err, "chat_id", chatID)
	}
}

// sendHTMLMessage sends an HTML formatted message via Composio
func (h *TelegramWebhookHandler) sendHTMLMessage(chatID int64, text string) {
	ctx := context.Background()
	chatIDStr := strconv.FormatInt(chatID, 10)

	req := &domainServices.SendTelegramMessageRequest{
		ChatID:    chatIDStr,
		Text:      text,
		ParseMode: "HTML",
	}
	if err := h.telegramService.SendMessage(ctx, req); err != nil {
		h.logger.Error("failed to send HTML message", "error", err, "chat_id", chatID)
	}
}

// handleFactCommand handles the /fact command
func (h *TelegramWebhookHandler) handleFactCommand(chatID int64) {
	if h.personalityService == nil || h.moodUseCase == nil {
		h.sendMessage(chatID, "💡 Функция фактов пока настраивается. Заходите позже!\n\n— Методыч")
		return
	}

	ctx := context.Background()
	mood, err := h.moodUseCase.GetCurrentMood(ctx)
	if err != nil {
		h.logger.Error("failed to get mood for fact command", "error", err)
		h.sendMessage(chatID, "😅 Что-то пошло не так... Попробуйте позже!\n\n— Методыч")
		return
	}

	// For now, send a mood-based response since fun facts are in Part 6
	h.sendHTMLMessage(chatID, fmt.Sprintf("💡 <b>Факт от Методыча</b>\n\n"+
		"Знаете ли вы, что слово «методист» происходит от греческого «methodos» — путь исследования? "+
		"За 40 лет работы я прошёл этот путь не один раз!\n\n"+
		"<i>Настроение: %s</i>\n\n"+
		"<i>— Ваш Методыч</i>", mood.State))
}

// handleMoodCommand handles the /mood command
func (h *TelegramWebhookHandler) handleMoodCommand(chatID int64) {
	if h.moodUseCase == nil {
		h.sendMessage(chatID, "🎭 Mood Engine пока настраивается. Заходите позже!\n\n— Методыч")
		return
	}

	ctx := context.Background()
	mood, err := h.moodUseCase.GetCurrentMood(ctx)
	if err != nil {
		h.logger.Error("failed to get mood", "error", err)
		h.sendMessage(chatID, "😅 Не могу определить настроение... Попробуйте позже!\n\n— Методыч")
		return
	}

	text := fmt.Sprintf("🎭 <b>Настроение Методыча</b>\n\n%s\n\n%s\n\n", mood.Greeting, mood.Message)

	if mood.OverdueDocuments > 0 {
		text += fmt.Sprintf("📋 Просроченных документов: %d\n", mood.OverdueDocuments)
	}
	if mood.AtRiskStudents > 0 {
		text += fmt.Sprintf("⚠️ Студентов в зоне риска: %d\n", mood.AtRiskStudents)
	}

	text += fmt.Sprintf("\n<i>Состояние: %s</i>", mood.State)

	h.sendHTMLMessage(chatID, text)
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
