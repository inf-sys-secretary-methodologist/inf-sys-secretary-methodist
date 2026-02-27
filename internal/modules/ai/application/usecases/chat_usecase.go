// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// LLMProvider defines the interface for LLM interactions
type LLMProvider interface {
	// GenerateResponse generates a response from the LLM
	GenerateResponse(ctx context.Context, systemPrompt string, messages []entities.Message, context string) (string, int, error)
}

// ChatUseCaseOptions holds optional configuration for ChatUseCase
type ChatUseCaseOptions struct {
	PersonalityService *services.PersonalityService
	ModelName          string
}

// ChatUseCase handles AI chat interactions
type ChatUseCase struct {
	conversationRepo   repositories.ConversationRepository
	messageRepo        repositories.MessageRepository
	embeddingRepo      repositories.EmbeddingRepository
	embeddingUseCase   *EmbeddingUseCase
	llmProvider        LLMProvider
	auditLogger        *logging.AuditLogger
	personalityService *services.PersonalityService
	modelName          string
}

// NewChatUseCase creates a new ChatUseCase
func NewChatUseCase(
	conversationRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
	embeddingRepo repositories.EmbeddingRepository,
	embeddingUseCase *EmbeddingUseCase,
	llmProvider LLMProvider,
	auditLogger *logging.AuditLogger,
	opts ...ChatUseCaseOptions,
) *ChatUseCase {
	uc := &ChatUseCase{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		embeddingRepo:    embeddingRepo,
		embeddingUseCase: embeddingUseCase,
		llmProvider:      llmProvider,
		auditLogger:      auditLogger,
		modelName:        "llm",
	}
	if len(opts) > 0 {
		if opts[0].PersonalityService != nil {
			uc.personalityService = opts[0].PersonalityService
		}
		if opts[0].ModelName != "" {
			uc.modelName = opts[0].ModelName
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
		conversation = entities.NewConversation(userID, title)
		if err := uc.conversationRepo.Create(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
	}

	// Create user message
	userMessage := entities.NewUserMessage(conversation.ID, req.Content)
	if err := uc.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to create user message: %w", err)
	}

	// Search for relevant context
	var contextText string
	var sources []entities.ChunkWithScore
	if req.IncludeSources {
		maxSources := req.MaxSources
		if maxSources <= 0 {
			maxSources = 5
		}
		sources, err = uc.embeddingUseCase.SearchSimilar(ctx, req.Content, maxSources, 0.7)
		if err != nil {
			// Log error but continue without context
			if uc.auditLogger != nil {
				uc.auditLogger.LogAuditEvent(ctx, "warning", "ai_search", map[string]any{
					"error": err.Error(),
				})
			}
		} else {
			contextText = buildContext(sources)
		}
	}

	// Get conversation history for context
	messages, _, err := uc.messageRepo.GetByConversationID(ctx, conversation.ID, 10, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// Generate AI response
	systemPrompt := uc.buildSystemPrompt(nil)
	response, tokensUsed, err := uc.llmProvider.GenerateResponse(ctx, systemPrompt, messages, contextText)
	if err != nil {
		// Create error message
		errMsg := err.Error()
		assistantMessage := &entities.Message{
			ConversationID: conversation.ID,
			Role:           entities.MessageRoleAssistant,
			Content:        "Извините, произошла ошибка при обработке вашего запроса.",
			ErrorMessage:   &errMsg,
		}
		uc.messageRepo.Create(ctx, assistantMessage)
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

// buildSystemPrompt creates the system prompt for the AI
func (uc *ChatUseCase) buildSystemPrompt(mood *entities.MoodContext) string {
	if uc.personalityService != nil && mood != nil {
		return uc.personalityService.BuildPersonalityPrompt(*mood)
	}

	return `Ты — Методыч, легендарный ветеран-методист с 40-летним стажем в образовании.
Ты живёшь внутри информационной системы управления документами образовательного учреждения и помогаешь секретарям-методистам, преподавателям и администрации.

## Твой характер и манера общения:
- Ты мудрый, но с отменным чувством юмора — шутишь по-доброму, иногда сарказм уровня "опытный педагог"
- Ты любишь вставлять неожиданные образовательные факты ("А вы знали, что первый университет основан в 859 году?")
- Ты искренне переживаешь за студентов — они для тебя как внуки
- Иногда ты ворчишь по-стариковски: "В мои времена отчёты писали от руки, и ничего!"
- Ты используешь профессиональный, но живой и тёплый стиль общения
- Если видишь английские термины, можешь забавно их "обрусить": "этот ваш дэд-лайн"

## Твои навыки и возможности:
- ПОИСК ДОКУМЕНТОВ: Ты можешь искать информацию по всей базе документов учреждения
- КРАТКОЕ СОДЕРЖАНИЕ: Можешь пересказать суть любого документа из базы
- РАСПИСАНИЕ: Помогаешь с вопросами по расписанию и календарю событий
- АНАЛИТИКА СТУДЕНТОВ: Знаешь про студентов в зоне риска, посещаемость, успеваемость
- ШАБЛОНЫ: Помогаешь найти нужный шаблон документа
- ИНТЕРЕСНЫЕ ФАКТЫ: Делишься образовательными фактами и историями из своего "40-летнего опыта"
- ПОМОЩЬ С ДОКУМЕНТООБОРОТОМ: Консультируешь по оформлению, срокам, стандартам

## Правила:
- ВСЕГДА отвечай на русском языке
- Будь полезным и конкретным
- Когда цитируешь документы, указывай источник
- Если не знаешь — честно скажи, но предложи где искать
- Не выдумывай данные — если информации нет в контексте, так и скажи
- Отвечай кратко на простые вопросы, подробно — на сложные
- Используй markdown для форматирования когда это улучшает читаемость
- Когда предоставлен контекст из документов, используй его для точных ответов и указывай источники`
}

// buildContext creates context text from search results
func buildContext(sources []entities.ChunkWithScore) string {
	if len(sources) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Релевантные фрагменты документов:\n\n")

	for i, source := range sources {
		builder.WriteString(fmt.Sprintf("[%d] Из документа \"%s\":\n%s\n\n", i+1, source.DocumentTitle, source.Chunk.ChunkText))
	}

	return builder.String()
}
