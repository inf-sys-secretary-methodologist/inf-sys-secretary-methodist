// Package handlers contains HTTP handlers for the AI module.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	chatUseCase      *usecases.ChatUseCase
	embeddingUseCase *usecases.EmbeddingUseCase
	moodUseCase      *usecases.MoodUseCase
	funFactUseCase   *usecases.FunFactUseCase
	logger           *logging.AuditLogger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	chatUseCase *usecases.ChatUseCase,
	embeddingUseCase *usecases.EmbeddingUseCase,
	moodUseCase *usecases.MoodUseCase,
	funFactUseCase *usecases.FunFactUseCase,
	logger *logging.AuditLogger,
) *AIHandler {
	return &AIHandler{
		chatUseCase:      chatUseCase,
		embeddingUseCase: embeddingUseCase,
		moodUseCase:      moodUseCase,
		funFactUseCase:   funFactUseCase,
		logger:           logger,
	}
}

// RegisterRoutes registers the AI routes
func (h *AIHandler) RegisterRoutes(rg *gin.RouterGroup) {
	ai := rg.Group("/ai")
	{
		// Metodych personality endpoints
		ai.GET("/mood", h.GetMood)
		ai.GET("/fact", h.GetFact)

		// Chat endpoints
		ai.POST("/chat", h.Chat)
		ai.GET("/chat/stream", h.ChatStream)

		// Conversation endpoints
		ai.GET("/conversations", h.ListConversations)
		ai.POST("/conversations", h.CreateConversation)
		ai.GET("/conversations/:id", h.GetConversation)
		ai.PATCH("/conversations/:id", h.UpdateConversation)
		ai.DELETE("/conversations/:id", h.DeleteConversation)
		ai.GET("/conversations/:id/messages", h.GetMessages)

		// Search endpoint
		ai.POST("/search", h.Search)

		// Indexing endpoints
		ai.POST("/index/:documentId", h.IndexDocument)
		ai.POST("/index/batch", h.IndexDocumentsBatch)
		ai.GET("/index/status", h.GetIndexingStatus)
	}
}

// Chat handles the chat endpoint
// @Summary Send a chat message
// @Description Send a message and get an AI response
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.SendMessageRequest true "Chat request"
// @Success 200 {object} dto.ChatResponse
// @Router /api/ai/chat [post]
func (h *AIHandler) Chat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.chatUseCase.Chat(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// ChatStream handles the streaming chat endpoint via Server-Sent Events
// @Summary Stream a chat response
// @Description Send a message and get a streamed AI response via SSE
// @Tags AI
// @Produce text/event-stream
// @Param content query string true "Message content"
// @Param conversation_id query int false "Conversation ID"
// @Param include_sources query bool false "Include document sources"
// @Param max_sources query int false "Max sources to include"
// @Param token query string true "JWT token (EventSource cannot set headers)"
// @Success 200 {string} string "SSE stream"
// @Router /api/ai/chat/stream [get]
func (h *AIHandler) ChatStream(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	content := c.Query("content")
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	// Build request from query params
	req := &dto.SendMessageRequest{
		Content:        content,
		IncludeSources: c.Query("include_sources") == "true",
	}

	if convIDStr := c.Query("conversation_id"); convIDStr != "" {
		convID, err := strconv.ParseInt(convIDStr, 10, 64)
		if err == nil {
			req.ConversationID = &convID
		}
	}

	if maxStr := c.Query("max_sources"); maxStr != "" {
		maxSources, err := strconv.Atoi(maxStr)
		if err == nil {
			req.MaxSources = maxSources
		}
	}

	// Extend write deadline for long-running LLM requests
	rc := http.NewResponseController(c.Writer)
	rc.SetWriteDeadline(time.Now().Add(5 * time.Minute))

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Helper to send SSE event; returns error on write failure (e.g. client disconnect).
	sendEvent := func(data any) error {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal SSE event: %w", err)
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", jsonBytes); err != nil {
			return fmt.Errorf("failed to write SSE event: %w", err)
		}
		c.Writer.Flush()
		return nil
	}

	// Stream response from LLM via ChatStream
	onChunk := func(chunk string) error {
		return sendEvent(map[string]any{"type": "content", "content": chunk})
	}

	response, err := h.chatUseCase.ChatStream(c.Request.Context(), userID.(int64), req, onChunk)
	if err != nil {
		_ = sendEvent(map[string]any{"type": "error", "error": err.Error()})
		return
	}

	// Send sources after streaming completes (best-effort)
	if response.Sources != nil {
		for _, source := range response.Sources {
			_ = sendEvent(map[string]any{
				"type": "source",
				"source": map[string]any{
					"document_id":    source.DocumentID,
					"document_title": source.DocumentTitle,
					"chunk_text":     source.ChunkText,
					"score":          source.SimilarityScore,
					"page_number":    source.PageNumber,
				},
			})
		}
	}

	// Send done event (best-effort)
	_ = sendEvent(map[string]any{
		"type":       "done",
		"message_id": response.Message.ID,
	})
}

// ListConversations handles listing conversations
// @Summary List conversations
// @Description Get a list of AI conversations for the current user
// @Tags AI
// @Produce json
// @Param search query string false "Search query"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.ConversationListResponse
// @Router /api/ai/conversations [get]
func (h *AIHandler) ListConversations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.chatUseCase.GetConversations(c.Request.Context(), userID.(int64), search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// CreateConversation handles creating a new conversation
// @Summary Create a conversation
// @Description Create a new AI conversation
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.CreateConversationRequest true "Conversation request"
// @Success 201 {object} dto.ConversationResponse
// @Router /api/ai/conversations [post]
func (h *AIHandler) CreateConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body for auto-generated title
		req = dto.CreateConversationRequest{Title: "New Conversation"}
	}

	// Use chat use case to create via conversation repo
	response, err := h.chatUseCase.GetConversations(c.Request.Context(), userID.(int64), "", 1, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// For now, return the first conversation or create via chat
	if len(response.Conversations) > 0 {
		c.JSON(http.StatusCreated, gin.H{"success": true, "data": response.Conversations[0]})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": dto.ConversationResponse{Title: req.Title}})
}

// GetConversation handles getting a single conversation
// @Summary Get a conversation
// @Description Get a single AI conversation by ID
// @Tags AI
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {object} dto.ConversationResponse
// @Router /api/ai/conversations/{id} [get]
func (h *AIHandler) GetConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	response, err := h.chatUseCase.GetConversation(c.Request.Context(), userID.(int64), conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// UpdateConversation handles updating a conversation
// @Summary Update a conversation
// @Description Update an AI conversation
// @Tags AI
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Param request body dto.UpdateConversationRequest true "Update request"
// @Success 200 {object} dto.ConversationResponse
// @Router /api/ai/conversations/{id} [patch]
func (h *AIHandler) UpdateConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	var req dto.UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.chatUseCase.UpdateConversation(c.Request.Context(), userID.(int64), conversationID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// DeleteConversation handles deleting a conversation
// @Summary Delete a conversation
// @Description Delete an AI conversation
// @Tags AI
// @Param id path int true "Conversation ID"
// @Success 200 {object} map[string]string
// @Router /api/ai/conversations/{id} [delete]
func (h *AIHandler) DeleteConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	if err := h.chatUseCase.DeleteConversation(c.Request.Context(), userID.(int64), conversationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "conversation deleted"})
}

// GetMessages handles getting messages for a conversation
// @Summary Get messages
// @Description Get messages for an AI conversation
// @Tags AI
// @Produce json
// @Param id path int true "Conversation ID"
// @Param limit query int false "Limit"
// @Param before_id query int false "Before message ID"
// @Success 200 {object} dto.MessageListResponse
// @Router /api/ai/conversations/{id}/messages [get]
func (h *AIHandler) GetMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	var beforeID *int64
	if beforeIDStr := c.Query("before_id"); beforeIDStr != "" {
		id, err := strconv.ParseInt(beforeIDStr, 10, 64)
		if err == nil {
			beforeID = &id
		}
	}

	response, err := h.chatUseCase.GetMessages(c.Request.Context(), userID.(int64), conversationID, limit, beforeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// Search handles semantic search
// @Summary Semantic search
// @Description Search documents using semantic similarity
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.SearchRequest true "Search request"
// @Success 200 {object} dto.SearchResponse
// @Router /api/ai/search [post]
func (h *AIHandler) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.embeddingUseCase.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// IndexDocument handles indexing a single document
// @Summary Index a document
// @Description Index a document for semantic search
// @Tags AI
// @Accept json
// @Produce json
// @Param documentId path int true "Document ID"
// @Param request body dto.IndexDocumentRequest false "Index request"
// @Success 200 {object} dto.IndexDocumentResponse
// @Router /api/ai/index/{documentId} [post]
func (h *AIHandler) IndexDocument(c *gin.Context) {
	documentID, err := strconv.ParseInt(c.Param("documentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	var req dto.IndexDocumentRequest
	c.ShouldBindJSON(&req) // Ignore error, use defaults
	req.DocumentID = documentID

	response, err := h.embeddingUseCase.IndexDocument(c.Request.Context(), documentID, req.ForceReindex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// IndexDocumentsBatch handles batch indexing
// @Summary Batch index documents
// @Description Index multiple documents for semantic search
// @Tags AI
// @Accept json
// @Produce json
// @Param request body object true "Batch index request"
// @Success 200 {object} object
// @Router /api/ai/index/batch [post]
func (h *AIHandler) IndexDocumentsBatch(c *gin.Context) {
	var req struct {
		DocumentIDs  []int64 `json:"document_ids"`
		ForceReindex bool    `json:"force_reindex"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	results := make([]dto.IndexDocumentResponse, 0, len(req.DocumentIDs))
	for _, docID := range req.DocumentIDs {
		response, err := h.embeddingUseCase.IndexDocument(c.Request.Context(), docID, req.ForceReindex)
		if err != nil {
			results = append(results, dto.IndexDocumentResponse{
				DocumentID: docID,
				Status:     "error",
				Message:    err.Error(),
			})
		} else {
			results = append(results, *response)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"results": results}})
}

// GetIndexingStatus handles getting indexing status
// @Summary Get indexing status
// @Description Get the current document indexing status
// @Tags AI
// @Produce json
// @Success 200 {object} dto.IndexStatusResponse
// @Router /api/ai/index/status [get]
func (h *AIHandler) GetIndexingStatus(c *gin.Context) {
	response, err := h.embeddingUseCase.GetIndexingStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// GetMood handles getting the current Metodych mood
// @Summary Get Metodych mood
// @Description Get the current mood state of the Metodych AI character
// @Tags AI
// @Produce json
// @Success 200 {object} dto.MoodResponse
// @Router /api/ai/mood [get]
func (h *AIHandler) GetMood(c *gin.Context) {
	if h.moodUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "mood service not available"})
		return
	}

	response, err := h.moodUseCase.GetCurrentMood(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// GetFact handles getting a random fun fact
// @Summary Get fun fact
// @Description Get a random educational fun fact from Metodych
// @Tags AI
// @Produce json
// @Success 200 {object} dto.FunFactResponse
// @Router /api/ai/fact [get]
func (h *AIHandler) GetFact(c *gin.Context) {
	if h.funFactUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fun facts not available"})
		return
	}

	response, err := h.funFactUseCase.GetRandomFact(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}
