// Package providers contains external service providers for the AI module.
package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// OpenAIConfig holds configuration for OpenAI API
type OpenAIConfig struct {
	APIKey         string
	BaseURL        string
	EmbeddingModel string
	ChatModel      string
	MaxTokens      int
	Temperature    float64
	Timeout        time.Duration
}

// DefaultOpenAIConfig returns the default OpenAI configuration
func DefaultOpenAIConfig() OpenAIConfig {
	return OpenAIConfig{
		BaseURL:        "https://api.openai.com/v1",
		EmbeddingModel: "text-embedding-3-small",
		ChatModel:      "gemini-2.5-flash",
		MaxTokens:      2048,
		Temperature:    0.3,
		Timeout:        60 * time.Second,
	}
}

// OpenAIProvider implements embedding and LLM providers using OpenAI API
type OpenAIProvider struct {
	config       OpenAIConfig
	client       *http.Client
	streamClient *http.Client // no timeout — context handles cancellation
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config OpenAIConfig) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = "text-embedding-3-small"
	}
	if config.ChatModel == "" {
		config.ChatModel = "gemini-2.5-flash"
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &OpenAIProvider{
		config:       config,
		client:       &http.Client{Timeout: config.Timeout},
		streamClient: &http.Client{},
	}
}

// embeddingRequest represents an OpenAI embedding API request
type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// embeddingResponse represents an OpenAI embedding API response
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// GenerateQueryEmbedding generates an embedding for search queries.
// OpenAI doesn't distinguish task types, so this delegates to GenerateEmbedding.
func (p *OpenAIProvider) GenerateQueryEmbedding(ctx context.Context, text string) ([]float32, error) {
	return p.GenerateEmbedding(ctx, text)
}

// GenerateEmbedding generates an embedding vector for text
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates embedding vectors for multiple texts
func (p *OpenAIProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := embeddingRequest{
		Input: texts,
		Model: p.config.EmbeddingModel,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var embeddingResp embeddingResponse
	if err := json.Unmarshal(respBody, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if embeddingResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", embeddingResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Sort by index and extract embeddings
	embeddings := make([][]float32, len(texts))
	for _, data := range embeddingResp.Data {
		if data.Index < len(embeddings) {
			embeddings[data.Index] = data.Embedding
		}
	}

	return embeddings, nil
}

// chatMessage represents a message in chat format
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest represents an OpenAI chat API request
type chatRequest struct {
	Model         string            `json:"model"`
	Messages      []chatMessage     `json:"messages"`
	MaxTokens     int               `json:"max_tokens,omitempty"`
	Temperature   float64           `json:"temperature,omitempty"`
	Stream        bool              `json:"stream,omitempty"`
	StreamOptions *chatStreamOption `json:"stream_options,omitempty"`
}

// chatStreamOption enables usage reporting in stream mode.
type chatStreamOption struct {
	IncludeUsage bool `json:"include_usage"`
}

// chatStreamChunk represents a single SSE chunk in OpenAI streaming response.
type chatStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// chatResponse represents an OpenAI chat API response
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// GenerateResponse generates a response from the LLM
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []entities.Message, contextText string) (string, int, error) {
	chatMessages := make([]chatMessage, 0, len(messages)+2)

	// Add system prompt with context
	systemContent := systemPrompt
	if contextText != "" {
		systemContent += "\n\n" + contextText
	}
	chatMessages = append(chatMessages, chatMessage{
		Role:    "system",
		Content: systemContent,
	})

	// Add conversation history
	for _, m := range messages {
		role := string(m.Role)
		if role == "system" {
			continue // Skip system messages from history
		}
		chatMessages = append(chatMessages, chatMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	reqBody := chatRequest{
		Model:       p.config.ChatModel,
		Messages:    chatMessages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", 0, fmt.Errorf("OpenAI API error: %s", chatResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if len(chatResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no response choices returned")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	tokensUsed := chatResp.Usage.TotalTokens

	return content, tokensUsed, nil
}

// GenerateResponseStream generates a streaming response from the LLM.
// Each text fragment is forwarded to onChunk as it arrives.
func (p *OpenAIProvider) GenerateResponseStream(ctx context.Context, systemPrompt string, messages []entities.Message, contextText string, onChunk func(string) error) (string, int, error) {
	chatMessages := make([]chatMessage, 0, len(messages)+2)

	systemContent := systemPrompt
	if contextText != "" {
		systemContent += "\n\n" + contextText
	}
	chatMessages = append(chatMessages, chatMessage{
		Role:    "system",
		Content: systemContent,
	})

	for _, m := range messages {
		role := string(m.Role)
		if role == "system" {
			continue
		}
		chatMessages = append(chatMessages, chatMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	reqBody := chatRequest{
		Model:       p.config.ChatModel,
		Messages:    chatMessages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Stream:      true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.streamClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var fullContent strings.Builder
	var totalTokens int

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024) // 256KB max line
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk chatStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Extract usage from the final chunk (stream_options.include_usage).
		if chunk.Usage != nil {
			totalTokens = chunk.Usage.TotalTokens
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			text := chunk.Choices[0].Delta.Content
			fullContent.WriteString(text)
			if err := onChunk(text); err != nil {
				return fullContent.String(), totalTokens, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fullContent.String(), totalTokens, fmt.Errorf("error reading stream: %w", err)
	}

	content := strings.TrimSpace(fullContent.String())
	if content == "" {
		return "", 0, fmt.Errorf("no text content in stream response")
	}

	return content, totalTokens, nil
}
