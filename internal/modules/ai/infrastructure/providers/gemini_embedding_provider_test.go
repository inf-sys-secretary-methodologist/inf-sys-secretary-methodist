package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewGeminiEmbeddingProvider_Defaults(t *testing.T) {
	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{APIKey: "key"})
	if p.config.BaseURL != "https://generativelanguage.googleapis.com/v1beta" {
		t.Errorf("unexpected BaseURL: %q", p.config.BaseURL)
	}
	if p.config.Model != "gemini-embedding-001" {
		t.Errorf("unexpected Model: %q", p.config.Model)
	}
	if p.config.OutputDimensionality != 3072 {
		t.Errorf("unexpected OutputDimensionality: %d", p.config.OutputDimensionality)
	}
	if p.config.Timeout != 60*time.Second {
		t.Errorf("unexpected Timeout: %v", p.config.Timeout)
	}
}

func TestNewGeminiEmbeddingProvider_CustomConfig(t *testing.T) {
	cfg := GeminiEmbeddingConfig{
		APIKey:               "key",
		BaseURL:              "https://custom.api.com",
		Model:                "custom-model",
		OutputDimensionality: 768,
		Timeout:              30 * time.Second,
	}
	p := NewGeminiEmbeddingProvider(cfg)
	if p.config.BaseURL != "https://custom.api.com" {
		t.Errorf("unexpected BaseURL: %q", p.config.BaseURL)
	}
	if p.config.Model != "custom-model" {
		t.Errorf("unexpected Model: %q", p.config.Model)
	}
	if p.config.OutputDimensionality != 768 {
		t.Errorf("unexpected OutputDimensionality: %d", p.config.OutputDimensionality)
	}
}

func TestGemini_GenerateEmbeddings_EmptyInput(t *testing.T) {
	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{APIKey: "key"})
	result, err := p.GenerateEmbeddings(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestGemini_GenerateEmbeddings_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("unexpected api key header")
		}
		if !strings.Contains(r.URL.Path, "batchEmbedContents") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req geminiBatchEmbedRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Requests) != 2 {
			t.Errorf("expected 2 requests, got %d", len(req.Requests))
		}
		if req.Requests[0].TaskType != "RETRIEVAL_DOCUMENT" {
			t.Errorf("expected RETRIEVAL_DOCUMENT task type, got %q", req.Requests[0].TaskType)
		}

		resp := geminiBatchEmbedResponse{
			Embeddings: []geminiEmbeddingValues{
				{Values: []float32{0.1, 0.2}},
				{Values: []float32{0.3, 0.4}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	result, err := p.GenerateEmbeddings(context.Background(), []string{"text1", "text2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(result))
	}
	if result[0][0] != 0.1 || result[1][0] != 0.3 {
		t.Errorf("unexpected embeddings: %v", result)
	}
}

func TestGemini_GenerateEmbedding_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiBatchEmbedResponse{
			Embeddings: []geminiEmbeddingValues{
				{Values: []float32{1.0, 2.0, 3.0}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	result, err := p.GenerateEmbedding(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 || result[0] != 1.0 {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestGemini_GenerateEmbedding_NoEmbeddings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mismatched count
		resp := geminiBatchEmbedResponse{
			Embeddings: []geminiEmbeddingValues{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	_, err := p.GenerateEmbedding(context.Background(), "text")
	if err == nil {
		t.Fatal("expected error for mismatched embedding count")
	}
}

func TestGemini_GenerateQueryEmbedding_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req geminiBatchEmbedRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Requests) != 1 {
			t.Errorf("expected 1 request, got %d", len(req.Requests))
		}
		if req.Requests[0].TaskType != "RETRIEVAL_QUERY" {
			t.Errorf("expected RETRIEVAL_QUERY, got %q", req.Requests[0].TaskType)
		}

		resp := geminiBatchEmbedResponse{
			Embeddings: []geminiEmbeddingValues{
				{Values: []float32{0.5, 0.6}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	result, err := p.GenerateQueryEmbedding(context.Background(), "query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != 0.5 {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestGemini_GenerateEmbeddings_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := struct {
			Error geminiAPIError `json:"error"`
		}{
			Error: geminiAPIError{
				Code:    400,
				Message: "invalid input",
				Status:  "INVALID_ARGUMENT",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	_, err := p.GenerateEmbeddings(context.Background(), []string{"text"})
	if err == nil {
		t.Fatal("expected error for API error")
	}
	if !strings.Contains(err.Error(), "invalid input") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGemini_GenerateEmbeddings_NonOKStatusNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	_, err := p.GenerateEmbeddings(context.Background(), []string{"text"})
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGemini_GenerateEmbeddings_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	_, err := p.GenerateEmbeddings(context.Background(), []string{"text"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGemini_GenerateEmbeddings_MismatchedCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiBatchEmbedResponse{
			Embeddings: []geminiEmbeddingValues{
				{Values: []float32{0.1}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	_, err := p.GenerateEmbeddings(context.Background(), []string{"text1", "text2"})
	if err == nil {
		t.Fatal("expected error for mismatched count")
	}
	if !strings.Contains(err.Error(), "expected 2 embeddings, got 1") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGemini_GenerateEmbeddings_LargeBatch(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req geminiBatchEmbedRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		embeddings := make([]geminiEmbeddingValues, len(req.Requests))
		for i := range embeddings {
			embeddings[i] = geminiEmbeddingValues{Values: []float32{float32(i)}}
		}
		resp := geminiBatchEmbedResponse{Embeddings: embeddings}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	// Create texts exceeding batch size limit
	texts := make([]string, geminiBatchSizeLimit+10)
	for i := range texts {
		texts[i] = "text"
	}

	result, err := p.GenerateEmbeddings(context.Background(), texts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != geminiBatchSizeLimit+10 {
		t.Errorf("expected %d embeddings, got %d", geminiBatchSizeLimit+10, len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 batch calls, got %d", callCount)
	}
}

func TestGemini_GenerateEmbeddings_LargeBatchError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error"))
			return
		}
		var req geminiBatchEmbedRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		embeddings := make([]geminiEmbeddingValues, len(req.Requests))
		for i := range embeddings {
			embeddings[i] = geminiEmbeddingValues{Values: []float32{0.1}}
		}
		resp := geminiBatchEmbedResponse{Embeddings: embeddings}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	texts := make([]string, geminiBatchSizeLimit+10)
	for i := range texts {
		texts[i] = "text"
	}

	_, err := p.GenerateEmbeddings(context.Background(), texts)
	if err == nil {
		t.Fatal("expected error from failed batch")
	}
	if !strings.Contains(err.Error(), "failed batch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGemini_GenerateQueryEmbedding_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiBatchEmbedResponse{
			Embeddings: []geminiEmbeddingValues{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiEmbeddingProvider(GeminiEmbeddingConfig{
		APIKey:  "key",
		BaseURL: server.URL,
	})

	_, err := p.GenerateQueryEmbedding(context.Background(), "query")
	if err == nil {
		t.Fatal("expected error for no embeddings")
	}
}
