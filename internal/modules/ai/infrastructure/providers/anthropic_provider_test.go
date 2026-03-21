package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

func TestDefaultAnthropicConfig(t *testing.T) {
	cfg := DefaultAnthropicConfig()
	if cfg.BaseURL != defaultAnthropicBaseURL {
		t.Errorf("unexpected BaseURL: %q", cfg.BaseURL)
	}
	if cfg.ChatModel != defaultAnthropicChatModel {
		t.Errorf("unexpected ChatModel: %q", cfg.ChatModel)
	}
	if cfg.MaxTokens != 2048 {
		t.Errorf("unexpected MaxTokens: %d", cfg.MaxTokens)
	}
	if cfg.RPMLimit != defaultRPMLimit {
		t.Errorf("unexpected RPMLimit: %d", cfg.RPMLimit)
	}
}

func TestNewAnthropicProvider_Defaults(t *testing.T) {
	p := NewAnthropicProvider(AnthropicConfig{APIKey: "key"})
	if p.config.BaseURL != defaultAnthropicBaseURL {
		t.Errorf("expected default BaseURL, got %q", p.config.BaseURL)
	}
	if p.config.ChatModel != defaultAnthropicChatModel {
		t.Errorf("expected default ChatModel, got %q", p.config.ChatModel)
	}
	if p.config.MaxTokens != 2048 {
		t.Errorf("expected default MaxTokens, got %d", p.config.MaxTokens)
	}
	if p.config.Timeout != 60*time.Second {
		t.Errorf("expected default Timeout, got %v", p.config.Timeout)
	}
	if p.config.RPMLimit != defaultRPMLimit {
		t.Errorf("expected default RPMLimit, got %d", p.config.RPMLimit)
	}
}

func TestNewAnthropicProvider_CustomConfig(t *testing.T) {
	cfg := AnthropicConfig{
		APIKey:    "key",
		BaseURL:   "https://custom.anthropic.com",
		ChatModel: "custom-model",
		MaxTokens: 4096,
		Timeout:   30 * time.Second,
		RPMLimit:  10,
	}
	p := NewAnthropicProvider(cfg)
	if p.config.BaseURL != "https://custom.anthropic.com" {
		t.Errorf("unexpected BaseURL: %q", p.config.BaseURL)
	}
	if p.config.RPMLimit != 10 {
		t.Errorf("unexpected RPMLimit: %d", p.config.RPMLimit)
	}
}

func TestErrRateLimited_Error(t *testing.T) {
	err := &ErrRateLimited{RetryAfter: 30 * time.Second}
	msg := err.Error()
	if !strings.Contains(msg, "30") {
		t.Errorf("expected 30 seconds in message, got: %s", msg)
	}
	if !strings.Contains(msg, "Методыч") {
		t.Errorf("expected personality in message, got: %s", msg)
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &rateLimitError{retryAfter: 5 * time.Second}
	if err.Error() != "rate limited by Anthropic API" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"", 0},
		{"5", 5 * time.Second},
		{"1.5", 1500 * time.Millisecond},
		{"invalid", 0},
	}
	for _, tt := range tests {
		got := parseRetryAfter(tt.input)
		if got != tt.expected {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestAnthropic_GenerateResponse_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("unexpected api key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("unexpected version header")
		}

		var req anthropicRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if !strings.Contains(req.System, "context data") {
			t.Error("expected context in system prompt")
		}
		// System messages should be filtered out
		for _, m := range req.Messages {
			if m.Role == "system" {
				t.Error("system messages should not appear in messages array")
			}
		}

		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: "  Response text  "},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 10, OutputTokens: 5},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "test-key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{
		{Role: entities.MessageRoleSystem, Content: "skip"},
		{Role: entities.MessageRoleUser, Content: "Hello"},
	}

	content, tokens, err := p.GenerateResponse(context.Background(), "prompt", msgs, "context data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Response text" {
		t.Errorf("expected trimmed 'Response text', got %q", content)
	}
	if tokens != 15 {
		t.Errorf("expected 15 tokens, got %d", tokens)
	}
}

func TestAnthropic_GenerateResponse_NoMessages(t *testing.T) {
	p := NewAnthropicProvider(AnthropicConfig{APIKey: "key", RPMLimit: 100})
	// Only system messages, which get filtered out
	msgs := []entities.Message{
		{Role: entities.MessageRoleSystem, Content: "system only"},
	}
	_, _, err := p.GenerateResponse(context.Background(), "prompt", msgs, "")
	if err == nil {
		t.Fatal("expected error for no user messages")
	}
	if !strings.Contains(err.Error(), "at least one user message") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnthropic_GenerateResponse_RateLimit429(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("retry-after", "0.01")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"type":"rate_limit","message":"too many requests"}}`))
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
		Timeout:  5 * time.Second,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	var rateLimited *ErrRateLimited
	if !errors.As(err, &rateLimited) {
		t.Errorf("expected ErrRateLimited, got: %v", err)
	}
	// Should have retried maxRetries + 1 times
	if callCount != maxRetries+1 {
		t.Errorf("expected %d calls, got %d", maxRetries+1, callCount)
	}
}

func TestAnthropic_GenerateResponse_529Overloaded(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(529)
			_, _ = w.Write([]byte(`{"error":{"type":"overloaded","message":"overloaded"}}`))
			return
		}
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "OK"}},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 1, OutputTokens: 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	content, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "OK" {
		t.Errorf("expected 'OK', got %q", content)
	}
}

func TestAnthropic_GenerateResponse_NonRetryableError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := anthropicErrorResponse{
			Type: "error",
			Error: struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			}{Type: "invalid_request_error", Message: "bad request"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnthropic_GenerateResponse_NonRetryableStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status 403") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnthropic_GenerateResponse_NoTextContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "image", Text: ""},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err == nil {
		t.Fatal("expected error for no text content")
	}
}

func TestAnthropic_GenerateResponse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{bad json"))
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAnthropic_GenerateResponseStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "Streamed response"}},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 5, OutputTokens: 3},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	var chunks []string
	content, tokens, err := p.GenerateResponseStream(context.Background(), "p", msgs, "", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Streamed response" {
		t.Errorf("expected 'Streamed response', got %q", content)
	}
	if tokens != 8 {
		t.Errorf("expected 8 tokens, got %d", tokens)
	}
	if len(chunks) != 1 || chunks[0] != "Streamed response" {
		t.Errorf("expected single chunk with full content, got %v", chunks)
	}
}

func TestAnthropic_GenerateResponseStream_OnChunkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "Response"}},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 1, OutputTokens: 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	chunkErr := errors.New("chunk callback failed")
	_, _, err := p.GenerateResponseStream(context.Background(), "p", msgs, "", func(chunk string) error {
		return chunkErr
	})
	if !errors.Is(err, chunkErr) {
		t.Errorf("expected chunk error, got: %v", err)
	}
}

func TestAnthropic_GenerateResponseStream_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := anthropicErrorResponse{
			Type: "error",
			Error: struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			}{Type: "invalid_request", Message: "bad"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponseStream(context.Background(), "p", msgs, "", func(chunk string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAnthropic_WaitForRateLimit_UnderLimit(t *testing.T) {
	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		RPMLimit: 100,
	})
	err := p.waitForRateLimit(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnthropic_WaitForRateLimit_ContextCanceled(t *testing.T) {
	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		RPMLimit: 1,
	})
	// Fill up the rate limit
	_ = p.waitForRateLimit(context.Background())

	// Next call should wait, but context is already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.waitForRateLimit(ctx)
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestAnthropic_GenerateResponse_NoContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if strings.Contains(req.System, "\n\n") {
			t.Error("system should not have double newline when no context")
		}

		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "OK"}},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 1, OutputTokens: 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	_, _, err := p.GenerateResponse(context.Background(), "prompt", msgs, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnthropic_GenerateResponse_MultipleTextBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: "Part 1 "},
				{Type: "text", Text: "Part 2"},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 5, OutputTokens: 3},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:   "key",
		BaseURL:  server.URL,
		RPMLimit: 100,
	})
	msgs := []entities.Message{{Role: entities.MessageRoleUser, Content: "Hi"}}

	content, _, err := p.GenerateResponse(context.Background(), "p", msgs, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Part 1 Part 2" {
		t.Errorf("expected joined text, got %q", content)
	}
}
