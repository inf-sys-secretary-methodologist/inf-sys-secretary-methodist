// Package providers contains external service providers for the AI module.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

const (
	// roleSystem is the system role identifier used in message filtering.
	roleSystem = "system"
	// maxRetries is the maximum number of retry attempts for rate-limited requests.
	maxRetries = 3
	// baseRetryDelay is the initial delay before first retry.
	baseRetryDelay = 2 * time.Second
	// rateLimitWindow is the sliding window for rate tracking.
	rateLimitWindow = time.Minute
	// defaultRPMLimit is the default requests-per-minute limit (Free Tier).
	defaultRPMLimit = 5
)

// AnthropicConfig holds configuration for Anthropic API
type AnthropicConfig struct {
	APIKey      string
	BaseURL     string
	ChatModel   string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
	RPMLimit    int // Requests per minute limit (0 = use default 5)
}

// DefaultAnthropicConfig returns the default Anthropic configuration
func DefaultAnthropicConfig() AnthropicConfig {
	return AnthropicConfig{
		BaseURL:     "https://api.anthropic.com",
		ChatModel:   "claude-haiku-4-5-20251001",
		MaxTokens:   2048,
		Temperature: 0.3,
		Timeout:     60 * time.Second,
		RPMLimit:    defaultRPMLimit,
	}
}

// AnthropicProvider implements LLM provider using Anthropic Messages API
type AnthropicProvider struct {
	config AnthropicConfig
	client *http.Client

	mu         sync.Mutex
	timestamps []time.Time // sliding window of request timestamps
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config AnthropicConfig) *AnthropicProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}
	if config.ChatModel == "" {
		config.ChatModel = "claude-haiku-4-5-20251001"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2048
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.RPMLimit == 0 {
		config.RPMLimit = defaultRPMLimit
	}

	return &AnthropicProvider{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// anthropicMessage represents a message in the Anthropic Messages API format
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicRequest represents an Anthropic Messages API request
type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature float64            `json:"temperature,omitempty"`
}

// anthropicResponse represents an Anthropic Messages API response
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// anthropicErrorResponse represents an Anthropic API error envelope
type anthropicErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ErrRateLimited is returned when the rate limit is exceeded after all retries.
type ErrRateLimited struct {
	RetryAfter time.Duration
}

func (e *ErrRateLimited) Error() string {
	return fmt.Sprintf(
		"Методыч сейчас перегружен запросами (лимит: %d запросов в минуту). Пожалуйста, подождите %d секунд и попробуйте снова.",
		defaultRPMLimit, int(e.RetryAfter.Seconds()),
	)
}

// waitForRateLimit checks the sliding window and waits if necessary.
// Returns an error if the context is canceled while waiting.
func (p *AnthropicProvider) waitForRateLimit(ctx context.Context) error {
	p.mu.Lock()

	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)

	// Prune old timestamps
	valid := p.timestamps[:0]
	for _, ts := range p.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	p.timestamps = valid

	if len(p.timestamps) < p.config.RPMLimit {
		// Under limit — record and proceed
		p.timestamps = append(p.timestamps, now)
		p.mu.Unlock()
		return nil
	}

	// At limit — calculate wait time until the oldest request expires
	waitUntil := p.timestamps[0].Add(rateLimitWindow)
	waitDuration := waitUntil.Sub(now)
	p.mu.Unlock()

	select {
	case <-time.After(waitDuration):
		// Record after waiting
		p.mu.Lock()
		p.timestamps = append(p.timestamps, time.Now())
		p.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GenerateResponse generates a response from the Anthropic LLM
func (p *AnthropicProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []entities.Message, contextText string) (string, int, error) {
	// Build system prompt with context
	systemContent := systemPrompt
	if contextText != "" {
		systemContent += "\n\n" + contextText
	}

	// Build messages (Anthropic does not use system role in messages array)
	chatMessages := make([]anthropicMessage, 0, len(messages))
	for _, m := range messages {
		role := string(m.Role)
		if role == roleSystem {
			continue
		}
		chatMessages = append(chatMessages, anthropicMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	if len(chatMessages) == 0 {
		return "", 0, fmt.Errorf("at least one user message is required")
	}

	reqBody := anthropicRequest{
		Model:       p.config.ChatModel,
		MaxTokens:   p.config.MaxTokens,
		System:      systemContent,
		Messages:    chatMessages,
		Temperature: p.config.Temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Client-side rate limiting
	if err := p.waitForRateLimit(ctx); err != nil {
		return "", 0, fmt.Errorf("rate limit wait canceled: %w", err)
	}

	// Execute with retry on 429
	return p.doRequestWithRetry(ctx, body)
}

// doRequestWithRetry sends the HTTP request and retries on 429 with exponential backoff.
func (p *AnthropicProvider) doRequestWithRetry(ctx context.Context, body []byte) (string, int, error) {
	var lastErr error

	for attempt := range maxRetries + 1 {
		content, tokens, err := p.doRequest(ctx, body)
		if err == nil {
			return content, tokens, nil
		}

		// Check if it's a rate limit error (429)
		var rateLimitErr *rateLimitError
		if errors.As(err, &rateLimitErr) {
			lastErr = err
			delay := rateLimitErr.retryAfter
			if delay == 0 {
				delay = baseRetryDelay * time.Duration(1<<attempt)
			}

			if attempt == maxRetries {
				break
			}

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return "", 0, ctx.Err()
			}
		}

		// Non-retryable error
		return "", 0, err
	}

	// All retries exhausted — return user-friendly error
	_ = lastErr
	return "", 0, &ErrRateLimited{RetryAfter: 30 * time.Second}
}

// rateLimitError is an internal error for 429 responses.
type rateLimitError struct {
	retryAfter time.Duration
}

func (e *rateLimitError) Error() string {
	return "rate limited by Anthropic API"
}

// doRequest sends a single HTTP request to the Anthropic Messages API.
func (p *AnthropicProvider) doRequest(ctx context.Context, body []byte) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle 429 — return internal error for retry logic
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("retry-after"))
		return "", 0, &rateLimitError{retryAfter: retryAfter}
	}

	// Handle 529 — Anthropic overloaded
	if resp.StatusCode == 529 {
		return "", 0, &rateLimitError{retryAfter: 10 * time.Second}
	}

	if resp.StatusCode != http.StatusOK {
		var errResp anthropicErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return "", 0, fmt.Errorf("anthropic API error (%s): %s", errResp.Error.Type, errResp.Error.Message)
		}
		return "", 0, fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract text from content blocks
	var textParts []string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			textParts = append(textParts, block.Text)
		}
	}

	if len(textParts) == 0 {
		return "", 0, fmt.Errorf("no text content in response")
	}

	content := strings.TrimSpace(strings.Join(textParts, ""))
	tokensUsed := anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens

	return content, tokensUsed, nil
}

// GenerateResponseStream generates a streaming response from the Anthropic LLM.
// Falls back to non-streaming GenerateResponse and delivers the full text as a single chunk.
func (p *AnthropicProvider) GenerateResponseStream(ctx context.Context, systemPrompt string, messages []entities.Message, contextText string, onChunk func(string) error) (string, int, error) {
	content, tokens, err := p.GenerateResponse(ctx, systemPrompt, messages, contextText)
	if err != nil {
		return "", 0, err
	}
	if err := onChunk(content); err != nil {
		return content, tokens, err
	}
	return content, tokens, nil
}

// parseRetryAfter parses the retry-after header value (seconds).
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
