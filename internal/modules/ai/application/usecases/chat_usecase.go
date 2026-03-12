// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/ports"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Deprecated: Use ports.LLMProvider directly. This alias exists for backward compatibility.
type LLMProvider = ports.LLMProvider

// ChatUseCaseOptions holds optional configuration for ChatUseCase
type ChatUseCaseOptions struct {
	ModelName       string
	SearchTopK      int
	SearchThreshold float64
	MoodUseCase     *MoodUseCase
}

// ChatUseCase handles AI chat interactions
type ChatUseCase struct {
	conversationRepo    repositories.ConversationRepository
	messageRepo         repositories.MessageRepository
	embeddingRepo       repositories.EmbeddingRepository
	embeddingUseCase    *EmbeddingUseCase
	llmProvider         LLMProvider
	auditLogger         *logging.AuditLogger
	personalityProvider services.PersonalityProvider
	moodUseCase         *MoodUseCase
	modelName           string
	searchTopK          int
	searchThreshold     float64
}

// NewChatUseCase creates a new ChatUseCase
func NewChatUseCase(
	conversationRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
	embeddingRepo repositories.EmbeddingRepository,
	embeddingUseCase *EmbeddingUseCase,
	llmProvider LLMProvider,
	personalityProvider services.PersonalityProvider,
	auditLogger *logging.AuditLogger,
	opts ...ChatUseCaseOptions,
) *ChatUseCase {
	uc := &ChatUseCase{
		conversationRepo:    conversationRepo,
		messageRepo:         messageRepo,
		embeddingRepo:       embeddingRepo,
		embeddingUseCase:    embeddingUseCase,
		llmProvider:         llmProvider,
		personalityProvider: personalityProvider,
		auditLogger:         auditLogger,
		modelName:           "llm",
		searchTopK:          10,
		searchThreshold:     0.7,
	}
	if len(opts) > 0 {
		if opts[0].ModelName != "" {
			uc.modelName = opts[0].ModelName
		}
		if opts[0].SearchTopK > 0 {
			uc.searchTopK = opts[0].SearchTopK
		}
		if opts[0].SearchThreshold > 0 {
			uc.searchThreshold = opts[0].SearchThreshold
		}
		if opts[0].MoodUseCase != nil {
			uc.moodUseCase = opts[0].MoodUseCase
		}
	}
	return uc
}

// Chat sends a message and gets an AI response
func (uc *ChatUseCase) Chat(ctx context.Context, userID int64, req *dto.SendMessageRequest) (*dto.ChatResponse, error) {
	var conversation *entities.Conversation
	var err error

	// Get or create conversation
	if req.ConversationID != nil {
		conversation, err = uc.conversationRepo.GetByID(ctx, *req.ConversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}
		if conversation.UserID != userID {
			return nil, fmt.Errorf("unauthorized access to conversation")
		}
	} else {
		// Create new conversation with truncated message as title
		title := req.Content
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		conversation = entities.NewConversation(userID, title, uc.modelName)
		if err := uc.conversationRepo.Create(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
	}

	// Create user message
	userMessage := entities.NewUserMessage(conversation.ID, req.Content)
	if err := uc.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to create user message: %w", err)
	}

	// Get conversation history for context
	messages, _, err := uc.messageRepo.GetByConversationID(ctx, conversation.ID, 10, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// RAG is always enabled — build search query expanded with conversation context
	searchQuery := buildSearchQuery(messages, req.Content)

	maxSources := uc.searchTopK
	if req.MaxSources > 0 {
		maxSources = req.MaxSources
	}

	var contextText string
	var sources []entities.ChunkWithScore
	sources, err = uc.embeddingUseCase.SearchSimilar(ctx, searchQuery, maxSources, uc.searchThreshold)
	if err != nil {
		// Log error but continue without context
		if uc.auditLogger != nil {
			uc.auditLogger.LogAuditEvent(ctx, "warning", "ai_search", map[string]any{
				"error": err.Error(),
			})
		}
	} else {
		contextText = uc.personalityProvider.FormatRAGContext(sources)
	}

	// Build system prompt with mood
	var mood *entities.MoodContext
	if uc.moodUseCase != nil {
		moodResp, err := uc.moodUseCase.GetCurrentMood(ctx)
		if err == nil && moodResp != nil {
			mood = &entities.MoodContext{
				State: entities.MoodState(moodResp.State),
			}
		}
	}
	systemPrompt := uc.personalityProvider.BuildSystemPrompt(moodForPrompt(mood))

	// Inject RAG context into the last user message instead of system prompt.
	// Models (especially smaller ones like Gemini Flash) pay much more attention
	// to user message content than to long system prompts, reducing hallucinations.
	llmMessages := messages
	if contextText != "" {
		llmMessages = make([]entities.Message, len(messages))
		copy(llmMessages, messages)
		for i := len(llmMessages) - 1; i >= 0; i-- {
			if llmMessages[i].Role == entities.MessageRoleUser {
				llmMessages[i].Content = contextText + "\nВопрос пользователя: " + llmMessages[i].Content
				break
			}
		}
	}
	response, tokensUsed, err := uc.llmProvider.GenerateResponse(ctx, systemPrompt, llmMessages, "")
	if err != nil {
		// Create error message
		errMsg := err.Error()
		assistantMessage := &entities.Message{
			ConversationID: conversation.ID,
			Role:           entities.MessageRoleAssistant,
			Content:        "Извините, произошла ошибка при обработке вашего запроса.",
			ErrorMessage:   &errMsg,
		}
		_ = uc.messageRepo.Create(ctx, assistantMessage) // best-effort error message save
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Create assistant message
	model := uc.modelName
	assistantMessage := entities.NewAssistantMessage(conversation.ID, response, model, tokensUsed)
	if err := uc.messageRepo.Create(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("failed to create assistant message: %w", err)
	}

	// Create message sources
	for _, source := range sources {
		if err := uc.messageRepo.CreateMessageSource(ctx, assistantMessage.ID, source.Chunk.ID, source.SimilarityScore); err != nil {
			// Log error but don't fail
			if uc.auditLogger != nil {
				uc.auditLogger.LogAuditEvent(ctx, "warning", "ai_source", map[string]any{
					"error": err.Error(),
				})
			}
		}
	}

	// Load sources into message
	assistantMessage.Sources = make([]entities.MessageSource, 0, len(sources))
	for _, source := range sources {
		assistantMessage.Sources = append(assistantMessage.Sources, entities.MessageSource{
			ChunkID:         source.Chunk.ID,
			DocumentID:      source.Chunk.DocumentID,
			DocumentTitle:   source.DocumentTitle,
			ChunkText:       source.Chunk.ChunkText,
			SimilarityScore: source.SimilarityScore,
			PageNumber:      source.Chunk.PageNumber,
		})
	}

	// Log audit event
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "ai_chat", map[string]any{
			"conversation_id": conversation.ID,
			"user_id":         userID,
			"tokens_used":     tokensUsed,
		})
	}

	return &dto.ChatResponse{
		Message:        *dto.ToMessageResponse(assistantMessage),
		ConversationID: conversation.ID,
		Sources:        dto.ToMessageResponse(assistantMessage).Sources,
	}, nil
}

// ChatStream sends a message and streams the AI response via onChunk callback.
// The callback is invoked for each text fragment as it arrives from the LLM.
func (uc *ChatUseCase) ChatStream(ctx context.Context, userID int64, req *dto.SendMessageRequest, onChunk func(chunk string) error) (*dto.ChatResponse, error) {
	var conversation *entities.Conversation
	var err error

	// Get or create conversation
	if req.ConversationID != nil {
		conversation, err = uc.conversationRepo.GetByID(ctx, *req.ConversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}
		if conversation.UserID != userID {
			return nil, fmt.Errorf("unauthorized access to conversation")
		}
	} else {
		title := req.Content
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		conversation = entities.NewConversation(userID, title, uc.modelName)
		if err := uc.conversationRepo.Create(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
	}

	// Create user message
	userMessage := entities.NewUserMessage(conversation.ID, req.Content)
	if err := uc.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to create user message: %w", err)
	}

	// Get conversation history for context
	messages, _, err := uc.messageRepo.GetByConversationID(ctx, conversation.ID, 10, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// RAG search
	searchQuery := buildSearchQuery(messages, req.Content)
	maxSources := uc.searchTopK
	if req.MaxSources > 0 {
		maxSources = req.MaxSources
	}

	var contextText string
	var sources []entities.ChunkWithScore
	sources, err = uc.embeddingUseCase.SearchSimilar(ctx, searchQuery, maxSources, uc.searchThreshold)
	if err != nil {
		if uc.auditLogger != nil {
			uc.auditLogger.LogAuditEvent(ctx, "warning", "ai_search", map[string]any{
				"error": err.Error(),
			})
		}
	} else {
		contextText = uc.personalityProvider.FormatRAGContext(sources)
	}

	// Build system prompt with mood
	var mood *entities.MoodContext
	if uc.moodUseCase != nil {
		moodResp, err := uc.moodUseCase.GetCurrentMood(ctx)
		if err == nil && moodResp != nil {
			mood = &entities.MoodContext{
				State: entities.MoodState(moodResp.State),
			}
		}
	}
	systemPrompt := uc.personalityProvider.BuildSystemPrompt(moodForPrompt(mood))

	// Inject RAG context into the last user message
	llmMessages := messages
	if contextText != "" {
		llmMessages = make([]entities.Message, len(messages))
		copy(llmMessages, messages)
		for i := len(llmMessages) - 1; i >= 0; i-- {
			if llmMessages[i].Role == entities.MessageRoleUser {
				llmMessages[i].Content = contextText + "\nВопрос пользователя: " + llmMessages[i].Content
				break
			}
		}
	}

	// Stream response from LLM
	response, tokensUsed, err := uc.llmProvider.GenerateResponseStream(ctx, systemPrompt, llmMessages, "", onChunk)
	if err != nil {
		errMsg := err.Error()
		assistantMessage := &entities.Message{
			ConversationID: conversation.ID,
			Role:           entities.MessageRoleAssistant,
			Content:        "Извините, произошла ошибка при обработке вашего запроса.",
			ErrorMessage:   &errMsg,
		}
		_ = uc.messageRepo.Create(ctx, assistantMessage) // best-effort error message save
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Save assistant message
	model := uc.modelName
	assistantMessage := entities.NewAssistantMessage(conversation.ID, response, model, tokensUsed)
	if err := uc.messageRepo.Create(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("failed to create assistant message: %w", err)
	}

	// Create message sources
	for _, source := range sources {
		if err := uc.messageRepo.CreateMessageSource(ctx, assistantMessage.ID, source.Chunk.ID, source.SimilarityScore); err != nil {
			if uc.auditLogger != nil {
				uc.auditLogger.LogAuditEvent(ctx, "warning", "ai_source", map[string]any{
					"error": err.Error(),
				})
			}
		}
	}

	// Load sources into message
	assistantMessage.Sources = make([]entities.MessageSource, 0, len(sources))
	for _, source := range sources {
		assistantMessage.Sources = append(assistantMessage.Sources, entities.MessageSource{
			ChunkID:         source.Chunk.ID,
			DocumentID:      source.Chunk.DocumentID,
			DocumentTitle:   source.DocumentTitle,
			ChunkText:       source.Chunk.ChunkText,
			SimilarityScore: source.SimilarityScore,
			PageNumber:      source.Chunk.PageNumber,
		})
	}

	// Log audit event
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "ai_chat", map[string]any{
			"conversation_id": conversation.ID,
			"user_id":         userID,
			"tokens_used":     tokensUsed,
		})
	}

	return &dto.ChatResponse{
		Message:        *dto.ToMessageResponse(assistantMessage),
		ConversationID: conversation.ID,
		Sources:        dto.ToMessageResponse(assistantMessage).Sources,
	}, nil
}

// buildSearchQuery expands the current query with recent conversation context
// to handle follow-up questions with pronouns (e.g., "а что там с этим приказом?").
func buildSearchQuery(messages []entities.Message, currentQuery string) string {
	var context strings.Builder
	count := 0
	for i := len(messages) - 1; i >= 0 && count < 2; i-- {
		if messages[i].Role == entities.MessageRoleUser || messages[i].Role == entities.MessageRoleAssistant {
			content := messages[i].Content
			if len([]rune(content)) > 200 {
				content = string([]rune(content)[:200])
			}
			context.WriteString(content)
			context.WriteString(" ")
			count++
		}
	}
	context.WriteString(currentQuery)
	return context.String()
}

// GetConversations retrieves conversations for a user
func (uc *ChatUseCase) GetConversations(ctx context.Context, userID int64, search string, limit, offset int) (*dto.ConversationListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var conversations []entities.Conversation
	var total int
	var err error

	if search != "" {
		conversations, total, err = uc.conversationRepo.Search(ctx, userID, search, limit, offset)
	} else {
		conversations, total, err = uc.conversationRepo.GetByUserID(ctx, userID, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	response := &dto.ConversationListResponse{
		Conversations: make([]dto.ConversationResponse, 0, len(conversations)),
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}

	for _, c := range conversations {
		response.Conversations = append(response.Conversations, *dto.ToConversationResponse(&c))
	}

	return response, nil
}

// GetConversation retrieves a single conversation
func (uc *ChatUseCase) GetConversation(ctx context.Context, userID, conversationID int64) (*dto.ConversationResponse, error) {
	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conversation.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to conversation")
	}
	return dto.ToConversationResponse(conversation), nil
}

// UpdateConversation updates a conversation
func (uc *ChatUseCase) UpdateConversation(ctx context.Context, userID, conversationID int64, req *dto.UpdateConversationRequest) (*dto.ConversationResponse, error) {
	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conversation.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to conversation")
	}

	conversation.Title = req.Title
	if err := uc.conversationRepo.Update(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return dto.ToConversationResponse(conversation), nil
}

// DeleteConversation deletes a conversation
func (uc *ChatUseCase) DeleteConversation(ctx context.Context, userID, conversationID int64) error {
	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}
	if conversation.UserID != userID {
		return fmt.Errorf("unauthorized access to conversation")
	}

	if err := uc.messageRepo.DeleteByConversationID(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	if err := uc.conversationRepo.Delete(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "delete", "ai_conversation", map[string]any{
			"conversation_id": conversationID,
			"user_id":         userID,
		})
	}

	return nil
}

// GetMessages retrieves messages for a conversation
func (uc *ChatUseCase) GetMessages(ctx context.Context, userID, conversationID int64, limit int, beforeID *int64) (*dto.MessageListResponse, error) {
	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conversation.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to conversation")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, hasMore, err := uc.messageRepo.GetByConversationID(ctx, conversationID, limit, beforeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Load sources for each assistant message
	for i := range messages {
		if messages[i].Role == entities.MessageRoleAssistant {
			sources, err := uc.messageRepo.GetMessageSources(ctx, messages[i].ID)
			if err == nil {
				messages[i].Sources = sources
			}
		}
	}

	response := &dto.MessageListResponse{
		Messages: make([]dto.MessageResponse, 0, len(messages)),
		HasMore:  hasMore,
	}

	for _, m := range messages {
		response.Messages = append(response.Messages, *dto.ToMessageResponse(&m))
	}

	return response, nil
}

// moodForPrompt converts a nullable MoodContext pointer to a value,
// defaulting to MoodContent if nil.
func moodForPrompt(mood *entities.MoodContext) entities.MoodContext {
	if mood != nil {
		return *mood
	}
	return entities.MoodContext{State: entities.MoodContent}
}
