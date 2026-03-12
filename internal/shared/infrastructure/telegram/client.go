// Package telegram provides a client for the Telegram Bot API.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// BaseURL is the Telegram Bot API base URL.
	BaseURL = "https://api.telegram.org"
)

// Client represents a Telegram Bot API client.
type Client struct {
	botToken   string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Telegram client.
func NewClient(botToken string) *Client {
	return &Client{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
	}
}

// SendMessageRequest represents a request to send a message.
type SendMessageRequest struct {
	ChatID                int64  `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool   `json:"disable_notification,omitempty"`
}

// APIResponse represents a generic Telegram API response.
type APIResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	Description string          `json:"description,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
}

// Message represents a Telegram message.
type Message struct {
	MessageID int64  `json:"message_id"`
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text,omitempty"`
}

// Chat represents a Telegram chat.
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// User represents a Telegram user.
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Update represents an incoming update from Telegram.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

// BotInfo represents bot information.
type BotInfo struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// SendMessage sends a text message to a chat.
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error) {
	if req.ChatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}
	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.botToken)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", apiResp.Description, apiResp.ErrorCode)
	}

	var message Message
	if err := json.Unmarshal(apiResp.Result, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetMe tests the bot token and returns basic information about the bot.
func (c *Client) GetMe(ctx context.Context) (*BotInfo, error) {
	url := fmt.Sprintf("%s/bot%s/getMe", c.baseURL, c.botToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("invalid bot token: %s", apiResp.Description)
	}

	var botInfo BotInfo
	if err := json.Unmarshal(apiResp.Result, &botInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bot info: %w", err)
	}

	return &botInfo, nil
}

// SetWebhook sets the webhook URL for the bot.
func (c *Client) SetWebhook(ctx context.Context, webhookURL string, secretToken string) error {
	url := fmt.Sprintf("%s/bot%s/setWebhook", c.baseURL, c.botToken)

	reqData := map[string]interface{}{
		"url": webhookURL,
	}
	if secretToken != "" {
		reqData["secret_token"] = secretToken
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("failed to set webhook: %s", apiResp.Description)
	}

	return nil
}

// DeleteWebhook removes the webhook.
func (c *Client) DeleteWebhook(ctx context.Context) error {
	url := fmt.Sprintf("%s/bot%s/deleteWebhook", c.baseURL, c.botToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("failed to delete webhook: %s", apiResp.Description)
	}

	return nil
}

// IsValidToken checks if the bot token is valid.
func (c *Client) IsValidToken(ctx context.Context) bool {
	_, err := c.GetMe(ctx)
	return err == nil
}

// SendNotification sends a formatted notification message.
func (c *Client) SendNotification(ctx context.Context, chatID int64, title, message string, priority string) error {
	// Format notification with emoji based on priority
	var emoji string
	switch priority {
	case "urgent":
		emoji = "🚨"
	case "high":
		emoji = "⚠️"
	case "normal":
		emoji = "📬"
	case "low":
		emoji = "ℹ️"
	default:
		emoji = "📬"
	}

	// Build HTML formatted message
	text := fmt.Sprintf(
		"%s <b>%s</b>\n\n%s",
		emoji,
		escapeHTML(title),
		escapeHTML(message),
	)

	req := &SendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}

	_, err := c.SendMessage(ctx, req)
	return err
}

// escapeHTML escapes special HTML characters for Telegram HTML parse mode.
func escapeHTML(text string) string {
	replacements := map[string]string{
		"&": "&amp;",
		"<": "&lt;",
		">": "&gt;",
	}

	result := text
	for old, newStr := range replacements {
		newResult := ""
		for _, char := range result {
			if string(char) == old {
				newResult += newStr
			} else {
				newResult += string(char)
			}
		}
		result = newResult
	}
	return result
}
