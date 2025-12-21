package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// MessagingHandler handles HTTP requests for messaging.
type MessagingHandler struct {
	useCase   *usecases.MessagingUseCase
	hub       *websocket.Hub
	logger    *logging.Logger
	validator *validation.Validator
}

// NewMessagingHandler creates a new messaging handler.
func NewMessagingHandler(
	useCase *usecases.MessagingUseCase,
	hub *websocket.Hub,
	logger *logging.Logger,
	validator *validation.Validator,
) *MessagingHandler {
	return &MessagingHandler{
		useCase:   useCase,
		hub:       hub,
		logger:    logger,
		validator: validator,
	}
}

// getUserID extracts user ID from context.
func (h *MessagingHandler) getUserID(c *gin.Context) (int64, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return 0, false
	}
	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return 0, false
	}
	return userID, true
}

// getConversationID extracts conversation ID from URL params.
func (h *MessagingHandler) getConversationID(c *gin.Context) (int64, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid conversation ID"))
		return 0, false
	}
	return id, true
}

// getMessageID extracts message ID from URL params.
func (h *MessagingHandler) getMessageID(c *gin.Context) (int64, bool) {
	idStr := c.Param("messageId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid message ID"))
		return 0, false
	}
	return id, true
}

// CreateDirectConversation creates a new direct conversation.
// @Summary Create direct conversation
// @Tags Messaging
// @Accept json
// @Produce json
// @Param input body dto.CreateDirectConversationInput true "Recipient info"
// @Success 201 {object} dto.ConversationOutput
// @Router /api/conversations/direct [post]
func (h *MessagingHandler) CreateDirectConversation(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.CreateDirectConversationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	conv, err := h.useCase.CreateDirectConversation(c.Request.Context(), userID, input)
	if err != nil {
		h.logger.Error("failed to create direct conversation", map[string]any{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, response.InternalError(""))
		return
	}

	c.JSON(http.StatusCreated, response.Success(dto.ToConversationOutput(conv, userID)))
}

// CreateGroupConversation creates a new group conversation.
// @Summary Create group conversation
// @Tags Messaging
// @Accept json
// @Produce json
// @Param input body dto.CreateGroupConversationInput true "Group info"
// @Success 201 {object} dto.ConversationOutput
// @Router /api/conversations/group [post]
func (h *MessagingHandler) CreateGroupConversation(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.CreateGroupConversationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	conv, err := h.useCase.CreateGroupConversation(c.Request.Context(), userID, input)
	if err != nil {
		h.logger.Error("failed to create group conversation", map[string]any{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, response.InternalError(""))
		return
	}

	c.JSON(http.StatusCreated, response.Success(dto.ToConversationOutput(conv, userID)))
}

// ListConversations lists user's conversations.
// @Summary List conversations
// @Tags Messaging
// @Produce json
// @Param type query string false "Conversation type (direct/group)"
// @Param search query string false "Search query"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.ConversationListOutput
// @Router /api/conversations [get]
func (h *MessagingHandler) ListConversations(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.ConversationFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid query parameters"))
		return
	}

	conversations, total, err := h.useCase.ListConversations(c.Request.Context(), userID, input)
	if err != nil {
		h.logger.Error("failed to list conversations", map[string]any{
			"error":   err.Error(),
			"user_id": userID,
		})
		c.JSON(http.StatusInternalServerError, response.InternalError(""))
		return
	}

	output := dto.ConversationListOutput{
		Conversations: make([]dto.ConversationOutput, 0, len(conversations)),
		Total:         total,
		Limit:         input.Limit,
		Offset:        input.Offset,
	}
	for _, conv := range conversations {
		output.Conversations = append(output.Conversations, dto.ToConversationOutput(conv, userID))
	}

	c.JSON(http.StatusOK, response.Success(output))
}

// GetConversation retrieves a conversation by ID.
// @Summary Get conversation
// @Tags Messaging
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {object} dto.ConversationOutput
// @Router /api/conversations/{id} [get]
func (h *MessagingHandler) GetConversation(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	conv, err := h.useCase.GetConversation(c.Request.Context(), userID, conversationID)
	if err != nil {
		h.logger.Error("failed to get conversation", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusNotFound, response.ErrorResponse("NOT_FOUND", "Conversation not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(dto.ToConversationOutput(conv, userID)))
}

// UpdateConversation updates a conversation.
// @Summary Update conversation
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Param input body dto.UpdateConversationInput true "Update info"
// @Success 200 {object} dto.ConversationOutput
// @Router /api/conversations/{id} [patch]
func (h *MessagingHandler) UpdateConversation(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	var input dto.UpdateConversationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	conv, err := h.useCase.UpdateConversation(c.Request.Context(), userID, conversationID, input)
	if err != nil {
		h.logger.Error("failed to update conversation", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(dto.ToConversationOutput(conv, userID)))
}

// AddParticipants adds participants to a group conversation.
// @Summary Add participants
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Param input body dto.AddParticipantsInput true "Participants"
// @Success 200 {object} response.Response
// @Router /api/conversations/{id}/participants [post]
func (h *MessagingHandler) AddParticipants(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	var input dto.AddParticipantsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	if err := h.useCase.AddParticipants(c.Request.Context(), userID, conversationID, input); err != nil {
		h.logger.Error("failed to add participants", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Participants added"}))
}

// LeaveConversation removes the current user from a conversation.
// @Summary Leave conversation
// @Tags Messaging
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {object} response.Response
// @Router /api/conversations/{id}/leave [post]
func (h *MessagingHandler) LeaveConversation(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	if err := h.useCase.LeaveConversation(c.Request.Context(), userID, conversationID); err != nil {
		h.logger.Error("failed to leave conversation", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Left conversation"}))
}

// SendMessage sends a message to a conversation.
// @Summary Send message
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Param input body dto.SendMessageInput true "Message"
// @Success 201 {object} dto.MessageOutput
// @Router /api/conversations/{id}/messages [post]
func (h *MessagingHandler) SendMessage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	var input dto.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	msg, err := h.useCase.SendMessage(c.Request.Context(), userID, conversationID, input)
	if err != nil {
		h.logger.Error("failed to send message", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(dto.ToMessageOutput(msg)))
}

// GetMessages retrieves messages from a conversation.
// @Summary Get messages
// @Tags Messaging
// @Produce json
// @Param id path int true "Conversation ID"
// @Param before_id query int false "Before message ID"
// @Param after_id query int false "After message ID"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.MessageListOutput
// @Router /api/conversations/{id}/messages [get]
func (h *MessagingHandler) GetMessages(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	var input dto.MessageFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid query parameters"))
		return
	}

	messages, hasMore, err := h.useCase.GetMessages(c.Request.Context(), userID, conversationID, input)
	if err != nil {
		h.logger.Error("failed to get messages", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	output := dto.MessageListOutput{
		Messages: make([]dto.MessageOutput, 0, len(messages)),
		HasMore:  hasMore,
	}
	for _, msg := range messages {
		output.Messages = append(output.Messages, dto.ToMessageOutput(msg))
	}

	c.JSON(http.StatusOK, response.Success(output))
}

// EditMessage edits a message.
// @Summary Edit message
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Param messageId path int true "Message ID"
// @Param input body dto.EditMessageInput true "New content"
// @Success 200 {object} dto.MessageOutput
// @Router /api/conversations/{id}/messages/{messageId} [patch]
func (h *MessagingHandler) EditMessage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	messageID, ok := h.getMessageID(c)
	if !ok {
		return
	}

	var input dto.EditMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	msg, err := h.useCase.EditMessage(c.Request.Context(), userID, messageID, input)
	if err != nil {
		h.logger.Error("failed to edit message", map[string]any{
			"error":      err.Error(),
			"message_id": messageID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(dto.ToMessageOutput(msg)))
}

// DeleteMessage deletes a message.
// @Summary Delete message
// @Tags Messaging
// @Produce json
// @Param id path int true "Conversation ID"
// @Param messageId path int true "Message ID"
// @Success 200 {object} response.Response
// @Router /api/conversations/{id}/messages/{messageId} [delete]
func (h *MessagingHandler) DeleteMessage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	messageID, ok := h.getMessageID(c)
	if !ok {
		return
	}

	if err := h.useCase.DeleteMessage(c.Request.Context(), userID, messageID); err != nil {
		h.logger.Error("failed to delete message", map[string]any{
			"error":      err.Error(),
			"message_id": messageID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Message deleted"}))
}

// MarkAsRead marks messages as read.
// @Summary Mark as read
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Param input body dto.MarkReadInput true "Last read message"
// @Success 200 {object} response.Response
// @Router /api/conversations/{id}/read [post]
func (h *MessagingHandler) MarkAsRead(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	var input dto.MarkReadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	if err := h.useCase.MarkAsRead(c.Request.Context(), userID, conversationID, input.MessageID); err != nil {
		h.logger.Error("failed to mark as read", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Marked as read"}))
}

// HandleWebSocket handles WebSocket connections.
// @Summary WebSocket connection
// @Tags Messaging
// @Success 101 "Switching Protocols"
// @Router /api/ws [get]
func (h *MessagingHandler) HandleWebSocket(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	websocket.ServeWs(h.hub, c.Writer, c.Request, userID, h.logger)
}

// SearchMessages searches messages within a conversation.
// @Summary Search messages
// @Tags Messaging
// @Produce json
// @Param id path int true "Conversation ID"
// @Param q query string true "Search query"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.SearchMessagesOutput
// @Router /api/conversations/{id}/messages/search [get]
func (h *MessagingHandler) SearchMessages(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversationID, ok := h.getConversationID(c)
	if !ok {
		return
	}

	var input dto.SearchMessagesInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid query parameters"))
		return
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}

	messages, total, err := h.useCase.SearchMessages(c.Request.Context(), userID, conversationID, input.Query, input.Limit, input.Offset)
	if err != nil {
		h.logger.Error("failed to search messages", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
			"query":           input.Query,
		})
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	output := dto.SearchMessagesOutput{
		Messages: make([]dto.MessageOutput, 0, len(messages)),
		Total:    total,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}
	for _, msg := range messages {
		output.Messages = append(output.Messages, dto.ToMessageOutput(msg))
	}

	c.JSON(http.StatusOK, response.Success(output))
}

// RegisterRoutes registers messaging routes.
func (h *MessagingHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	messaging := router.Group("/conversations")
	messaging.Use(authMiddleware)
	{
		messaging.POST("/direct", h.CreateDirectConversation)
		messaging.POST("/group", h.CreateGroupConversation)
		messaging.GET("", h.ListConversations)
		messaging.GET("/:id", h.GetConversation)
		messaging.PATCH("/:id", h.UpdateConversation)
		messaging.POST("/:id/participants", h.AddParticipants)
		messaging.POST("/:id/leave", h.LeaveConversation)
		messaging.POST("/:id/messages", h.SendMessage)
		messaging.GET("/:id/messages", h.GetMessages)
		messaging.GET("/:id/messages/search", h.SearchMessages)
		messaging.PATCH("/:id/messages/:messageId", h.EditMessage)
		messaging.DELETE("/:id/messages/:messageId", h.DeleteMessage)
		messaging.POST("/:id/read", h.MarkAsRead)
	}

	// WebSocket endpoint
	router.GET("/ws", authMiddleware, h.HandleWebSocket)
}
