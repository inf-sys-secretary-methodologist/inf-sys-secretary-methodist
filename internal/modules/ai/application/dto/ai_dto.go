// Package dto contains data transfer objects for the AI module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// ConversationResponse represents a conversation in API responses
type ConversationResponse struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Title         string     `json:"title"`
	Model         string     `json:"model"`
	MessageCount  int        `json:"message_count"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ConversationListResponse represents a paginated list of conversations
type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID             int64            `json:"id"`
	ConversationID int64            `json:"conversation_id"`
	Role           string           `json:"role"`
	Content        string           `json:"content"`
	Sources        []SourceResponse `json:"sources,omitempty"`
	TokensUsed     *int             `json:"tokens_used,omitempty"`
	Model          *string          `json:"model,omitempty"`
	Status         string           `json:"status"`
	ErrorMessage   *string          `json:"error_message,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

// SourceResponse represents a document source citation
type SourceResponse struct {
	ID              int64   `json:"id"`
	DocumentID      int64   `json:"document_id"`
	DocumentTitle   string  `json:"document_title"`
	ChunkText       string  `json:"chunk_text"`
	SimilarityScore float64 `json:"similarity_score"`
	PageNumber      *int    `json:"page_number,omitempty"`
}

// MessageListResponse represents a paginated list of messages
type MessageListResponse struct {
	Messages []MessageResponse `json:"messages"`
	HasMore  bool              `json:"has_more"`
}

// CreateConversationRequest represents a request to create a conversation
type CreateConversationRequest struct {
	Title string `json:"title"`
	Model string `json:"model,omitempty"`
}

// UpdateConversationRequest represents a request to update a conversation
type UpdateConversationRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

// SendMessageRequest represents a request to send a chat message
type SendMessageRequest struct {
	Content        string `json:"content" validate:"required,min=1,max=10000"`
	ConversationID *int64 `json:"conversation_id,omitempty"`
	IncludeSources bool   `json:"include_sources"`
	MaxSources     int    `json:"max_sources,omitempty"`
}

// ChatResponse represents the response from a chat request
type ChatResponse struct {
	Message        MessageResponse  `json:"message"`
	ConversationID int64            `json:"conversation_id"`
	Sources        []SourceResponse `json:"sources,omitempty"`
	MoodState      string           `json:"mood_state,omitempty"`
	MoodMessage    string           `json:"mood_message,omitempty"`
}

// SearchRequest represents a semantic search request
type SearchRequest struct {
	Query         string   `json:"query" validate:"required,min=1,max=1000"`
	Limit         int      `json:"limit,omitempty"`
	Threshold     float64  `json:"threshold,omitempty"`
	DocumentTypes []string `json:"document_types,omitempty"`
}

// SearchResultResponse represents a single search result
type SearchResultResponse struct {
	DocumentID      int64   `json:"document_id"`
	DocumentTitle   string  `json:"document_title"`
	ChunkText       string  `json:"chunk_text"`
	SimilarityScore float64 `json:"similarity_score"`
	PageNumber      *int    `json:"page_number,omitempty"`
}

// SearchResponse represents the response from a search request
type SearchResponse struct {
	Results []SearchResultResponse `json:"results"`
	Query   string                 `json:"query"`
	Total   int                    `json:"total"`
}

// IndexDocumentRequest represents a request to index a document
type IndexDocumentRequest struct {
	DocumentID   int64 `json:"document_id" validate:"required"`
	ForceReindex bool  `json:"force_reindex,omitempty"`
}

// IndexDocumentResponse represents the response from indexing a document
type IndexDocumentResponse struct {
	DocumentID    int64  `json:"document_id"`
	ChunksCreated int    `json:"chunks_created"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}

// IndexStatusResponse represents the indexing status
type IndexStatusResponse struct {
	TotalDocuments   int     `json:"total_documents"`
	IndexedDocuments int     `json:"indexed_documents"`
	PendingDocuments int     `json:"pending_documents"`
	LastIndexedAt    *string `json:"last_indexed_at,omitempty"`
}

// ToConversationResponse converts a domain conversation to response DTO
func ToConversationResponse(c *entities.Conversation) *ConversationResponse {
	return &ConversationResponse{
		ID:            c.ID,
		UserID:        c.UserID,
		Title:         c.Title,
		Model:         c.Model,
		MessageCount:  c.MessageCount,
		LastMessageAt: c.LastMessageAt,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ToMessageResponse converts a domain message to response DTO
func ToMessageResponse(m *entities.Message) *MessageResponse {
	sources := make([]SourceResponse, 0, len(m.Sources))
	for _, s := range m.Sources {
		sources = append(sources, SourceResponse{
			ID:              s.ID,
			DocumentID:      s.DocumentID,
			DocumentTitle:   s.DocumentTitle,
			ChunkText:       s.ChunkText,
			SimilarityScore: s.SimilarityScore,
			PageNumber:      s.PageNumber,
		})
	}

	status := "complete"
	if m.ErrorMessage != nil {
		status = "error"
	}

	return &MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Role:           string(m.Role),
		Content:        m.Content,
		Sources:        sources,
		TokensUsed:     m.TokensUsed,
		Model:          m.Model,
		Status:         status,
		ErrorMessage:   m.ErrorMessage,
		CreatedAt:      m.CreatedAt,
	}
}

// ToSearchResultResponse converts a chunk with score to search result response
func ToSearchResultResponse(c *entities.ChunkWithScore) *SearchResultResponse {
	return &SearchResultResponse{
		DocumentID:      c.Chunk.DocumentID,
		DocumentTitle:   c.DocumentTitle,
		ChunkText:       c.Chunk.ChunkText,
		SimilarityScore: c.SimilarityScore,
		PageNumber:      c.Chunk.PageNumber,
	}
}
