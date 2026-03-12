// Package providers contains external service providers for the AI module.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeminiEmbeddingConfig holds configuration for Gemini Embedding API
type GeminiEmbeddingConfig struct {
	APIKey               string
	BaseURL              string
	Model                string
	OutputDimensionality int
	Timeout              time.Duration
}

// GeminiEmbeddingProvider implements EmbeddingProvider using Gemini Embedding API
type GeminiEmbeddingProvider struct {
	config GeminiEmbeddingConfig
	client *http.Client
}

// NewGeminiEmbeddingProvider creates a new Gemini embedding provider
func NewGeminiEmbeddingProvider(config GeminiEmbeddingConfig) *GeminiEmbeddingProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	if config.Model == "" {
		config.Model = "gemini-embedding-001"
	}
	if config.OutputDimensionality == 0 {
		config.OutputDimensionality = 3072
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &GeminiEmbeddingProvider{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// --- Request/Response types for Gemini Embedding API ---

type geminiEmbedContentRequest struct {
	Content              geminiContent `json:"content"`
	TaskType             string        `json:"taskType,omitempty"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
}

type geminiBatchEmbedRequest struct {
	Requests []geminiBatchEmbedEntry `json:"requests"`
}

type geminiBatchEmbedEntry struct {
	Model                string        `json:"model"`
	Content              geminiContent `json:"content"`
	TaskType             string        `json:"taskType,omitempty"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiEmbedContentResponse struct {
	Embedding geminiEmbeddingValues `json:"embedding"`
	Error     *geminiAPIError       `json:"error,omitempty"`
}

type geminiBatchEmbedResponse struct {
	Embeddings []geminiEmbeddingValues `json:"embeddings"`
	Error      *geminiAPIError         `json:"error,omitempty"`
}

type geminiEmbeddingValues struct {
	Values []float32 `json:"values"`
}

type geminiAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// geminiBatchSizeLimit is the maximum number of texts per Gemini batch API call.
const geminiBatchSizeLimit = 100

// GenerateEmbedding generates an embedding vector for a single text (RETRIEVAL_DOCUMENT task type)
func (p *GeminiEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// GenerateQueryEmbedding generates an embedding vector optimized for search queries (RETRIEVAL_QUERY task type).
func (p *GeminiEmbeddingProvider) GenerateQueryEmbedding(ctx context.Context, text string) ([]float32, error) {
	return p.generateBatchEmbeddings(ctx, []string{text}, "RETRIEVAL_QUERY")
}

// GenerateEmbeddings generates embedding vectors for multiple texts using batch API (RETRIEVAL_DOCUMENT task type).
// Automatically splits into sub-batches of geminiBatchSizeLimit texts.
func (p *GeminiEmbeddingProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Split into sub-batches if needed
	if len(texts) <= geminiBatchSizeLimit {
		return p.generateBatchEmbeddingsMulti(ctx, texts, "RETRIEVAL_DOCUMENT")
	}

	allEmbeddings := make([][]float32, 0, len(texts))
	for i := 0; i < len(texts); i += geminiBatchSizeLimit {
		end := i + geminiBatchSizeLimit
		if end > len(texts) {
			end = len(texts)
		}
		batch, err := p.generateBatchEmbeddingsMulti(ctx, texts[i:end], "RETRIEVAL_DOCUMENT")
		if err != nil {
			return nil, fmt.Errorf("failed batch %d-%d: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, batch...)
	}
	return allEmbeddings, nil
}

// generateBatchEmbeddings calls the batch API for a single text and returns the first embedding.
func (p *GeminiEmbeddingProvider) generateBatchEmbeddings(ctx context.Context, texts []string, taskType string) ([]float32, error) {
	results, err := p.generateBatchEmbeddingsMulti(ctx, texts, taskType)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return results[0], nil
}

// generateBatchEmbeddingsMulti is the core batch embedding method.
func (p *GeminiEmbeddingProvider) generateBatchEmbeddingsMulti(ctx context.Context, texts []string, taskType string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	modelPath := "models/" + p.config.Model

	requests := make([]geminiBatchEmbedEntry, len(texts))
	for i, text := range texts {
		requests[i] = geminiBatchEmbedEntry{
			Model: modelPath,
			Content: geminiContent{
				Parts: []geminiPart{{Text: text}},
			},
			TaskType:             taskType,
			OutputDimensionality: p.config.OutputDimensionality,
		}
	}

	reqBody := geminiBatchEmbedRequest{Requests: requests}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:batchEmbedContents", p.config.BaseURL, modelPath)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error geminiAPIError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("Gemini Embedding API error (%s): %s", errResp.Error.Status, errResp.Error.Message)
		}
		return nil, fmt.Errorf("Gemini Embedding API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var batchResp geminiBatchEmbedResponse
	if err := json.Unmarshal(respBody, &batchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(batchResp.Embeddings) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(batchResp.Embeddings))
	}

	embeddings := make([][]float32, len(texts))
	for i, emb := range batchResp.Embeddings {
		embeddings[i] = emb.Values
	}

	return embeddings, nil
}
