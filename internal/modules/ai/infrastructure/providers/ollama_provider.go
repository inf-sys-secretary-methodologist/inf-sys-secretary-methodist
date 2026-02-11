// Package providers contains external service providers for the AI module.
package providers

import (
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

// OllamaConfig holds configuration for Ollama API
type OllamaConfig struct {
	BaseURL        string
	EmbeddingModel string
	ChatModel      string
	Timeout        time.Duration
}

// DefaultOllamaConfig returns the default Ollama configuration
func DefaultOllamaConfig() OllamaConfig {
	return OllamaConfig{
		BaseURL:        "http://localhost:11434",
		EmbeddingModel: "nomic-embed-text",
		ChatModel:      "llama3.2",
		Timeout:        120 * time.Second,
	}
}

// OllamaProvider implements embedding and LLM providers using Ollama API
type OllamaProvider struct {
	config OllamaConfig
	client *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config OllamaConfig) *OllamaProvider {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = "nomic-embed-text"
	}
	if config.ChatModel == "" {
		config.ChatModel = "llama3.2"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}

	return &OllamaProvider{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// ollamaEmbeddingRequest represents an Ollama embedding API request
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbeddingResponse represents an Ollama embedding API response
type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

// GenerateEmbedding generates an embedding vector for text
func (p *OllamaProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := ollamaEmbeddingRequest{
		Model:  p.config.EmbeddingModel,
		Prompt: text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var embeddingResp ollamaEmbeddingResponse
	if err := json.Unmarshal(respBody, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if embeddingResp.Error != "" {
		return nil, fmt.Errorf("Ollama API error: %s", embeddingResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Convert float64 to float32
	embedding := make([]float32, len(embeddingResp.Embedding))
	for i, v := range embeddingResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// GenerateEmbeddings generates embedding vectors for multiple texts
func (p *OllamaProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := p.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// ollamaChatRequest represents an Ollama chat API request
type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

// ollamaChatMessage represents a message in Ollama chat format
type ollamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatResponse represents an Ollama chat API response
type ollamaChatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
	Error              string `json:"error,omitempty"`
}

// GenerateResponse generates a response from the LLM
func (p *OllamaProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []entities.Message, contextText string) (string, int, error) {
	chatMessages := make([]ollamaChatMessage, 0, len(messages)+2)

	// Add system prompt with context
	systemContent := systemPrompt
	if contextText != "" {
		systemContent += "\n\n" + contextText
	}
	chatMessages = append(chatMessages, ollamaChatMessage{
		Role:    "system",
		Content: systemContent,
	})

	// Add conversation history
	for _, m := range messages {
		role := string(m.Role)
		if role == "system" {
			continue // Skip system messages from history
		}
		chatMessages = append(chatMessages, ollamaChatMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	reqBody := ollamaChatRequest{
		Model:    p.config.ChatModel,
		Messages: chatMessages,
		Stream:   false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != "" {
		return "", 0, fmt.Errorf("Ollama API error: %s", chatResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	content := strings.TrimSpace(chatResp.Message.Content)
	// Ollama provides eval_count which is roughly the token count
	tokensUsed := chatResp.PromptEvalCount + chatResp.EvalCount

	return content, tokensUsed, nil
}
