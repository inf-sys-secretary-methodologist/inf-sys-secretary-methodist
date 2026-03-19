package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	dashboardRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// ============================================================
// Mock implementations
// ============================================================

// --- MockConversationRepository ---

type MockConversationRepo struct {
	mock.Mock
}

func (m *MockConversationRepo) Create(_ context.Context, c *entities.Conversation) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockConversationRepo) GetByID(_ context.Context, id int64) (*entities.Conversation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Conversation), args.Error(1)
}

func (m *MockConversationRepo) GetByUserID(_ context.Context, userID int64, limit, offset int) ([]entities.Conversation, int, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entities.Conversation), args.Int(1), args.Error(2)
}

func (m *MockConversationRepo) Update(_ context.Context, c *entities.Conversation) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockConversationRepo) Delete(_ context.Context, id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockConversationRepo) Search(_ context.Context, userID int64, query string, limit, offset int) ([]entities.Conversation, int, error) {
	args := m.Called(userID, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entities.Conversation), args.Int(1), args.Error(2)
}

// --- MockMessageRepository ---

type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) Create(_ context.Context, msg *entities.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockMessageRepo) GetByConversationID(_ context.Context, convID int64, limit int, beforeID *int64) ([]entities.Message, bool, error) {
	args := m.Called(convID, limit, beforeID)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).([]entities.Message), args.Bool(1), args.Error(2)
}

func (m *MockMessageRepo) GetByID(_ context.Context, id int64) (*entities.Message, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepo) CreateMessageSource(_ context.Context, messageID, chunkID int64, score float64) error {
	args := m.Called(messageID, chunkID, score)
	return args.Error(0)
}

func (m *MockMessageRepo) GetMessageSources(_ context.Context, messageID int64) ([]entities.MessageSource, error) {
	args := m.Called(messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.MessageSource), args.Error(1)
}

func (m *MockMessageRepo) DeleteByConversationID(_ context.Context, convID int64) error {
	args := m.Called(convID)
	return args.Error(0)
}

// --- MockEmbeddingRepository ---

type MockEmbeddingRepo struct {
	mock.Mock
}

func (m *MockEmbeddingRepo) CreateChunk(_ context.Context, chunk *entities.DocumentChunk) error {
	args := m.Called(chunk)
	return args.Error(0)
}

func (m *MockEmbeddingRepo) CreateChunks(_ context.Context, chunks []entities.DocumentChunk) error {
	args := m.Called(chunks)
	return args.Error(0)
}

func (m *MockEmbeddingRepo) GetChunksByDocumentID(_ context.Context, docID int64) ([]entities.DocumentChunk, error) {
	args := m.Called(docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DocumentChunk), args.Error(1)
}

func (m *MockEmbeddingRepo) DeleteChunksByDocumentID(_ context.Context, docID int64) error {
	args := m.Called(docID)
	return args.Error(0)
}

func (m *MockEmbeddingRepo) CreateEmbedding(_ context.Context, emb *entities.Embedding) error {
	args := m.Called(emb)
	return args.Error(0)
}

func (m *MockEmbeddingRepo) CreateEmbeddings(_ context.Context, embeddings []entities.Embedding) error {
	args := m.Called(embeddings)
	return args.Error(0)
}

func (m *MockEmbeddingRepo) SearchSimilar(_ context.Context, embedding []float32, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	args := m.Called(embedding, limit, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.ChunkWithScore), args.Error(1)
}

func (m *MockEmbeddingRepo) SearchSimilarByDocumentTypes(_ context.Context, embedding []float32, docTypes []string, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	args := m.Called(embedding, docTypes, limit, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.ChunkWithScore), args.Error(1)
}

func (m *MockEmbeddingRepo) GetIndexStatus(_ context.Context, docID int64) (*entities.DocumentIndexStatus, error) {
	args := m.Called(docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentIndexStatus), args.Error(1)
}

func (m *MockEmbeddingRepo) SetIndexStatus(_ context.Context, status *entities.DocumentIndexStatus) error {
	args := m.Called(status)
	return args.Error(0)
}

func (m *MockEmbeddingRepo) GetPendingDocuments(_ context.Context, limit int) ([]int64, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockEmbeddingRepo) GetIndexingStats(_ context.Context) (int, int, int, *string, error) {
	args := m.Called()
	var lastIdx *string
	if args.Get(3) != nil {
		v := args.Get(3).(*string)
		lastIdx = v
	}
	return args.Int(0), args.Int(1), args.Int(2), lastIdx, args.Error(4)
}

func (m *MockEmbeddingRepo) SearchHybrid(_ context.Context, embedding []float32, queryText string, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	args := m.Called(embedding, queryText, limit, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.ChunkWithScore), args.Error(1)
}

func (m *MockEmbeddingRepo) GetAdjacentChunks(_ context.Context, chunkIDs []int64, windowSize int) ([]entities.DocumentChunk, error) {
	args := m.Called(chunkIDs, windowSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DocumentChunk), args.Error(1)
}

// --- MockLLMProvider ---

type MockLLMProvider struct {
	mock.Mock
}

func (m *MockLLMProvider) GenerateResponse(_ context.Context, systemPrompt string, messages []entities.Message, ctx string) (string, int, error) {
	args := m.Called(systemPrompt, messages, ctx)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *MockLLMProvider) GenerateResponseStream(_ context.Context, systemPrompt string, messages []entities.Message, ctx string, onChunk func(string) error) (string, int, error) {
	args := m.Called(systemPrompt, messages, ctx, onChunk)
	return args.String(0), args.Int(1), args.Error(2)
}

// --- MockEmbeddingProvider ---

type MockEmbeddingProvider struct {
	mock.Mock
}

func (m *MockEmbeddingProvider) GenerateEmbedding(_ context.Context, text string) ([]float32, error) {
	args := m.Called(text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockEmbeddingProvider) GenerateEmbeddings(_ context.Context, texts []string) ([][]float32, error) {
	args := m.Called(texts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

func (m *MockEmbeddingProvider) GenerateQueryEmbedding(_ context.Context, text string) ([]float32, error) {
	args := m.Called(text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}

// --- MockDocumentProvider ---

type MockDocumentProvider struct {
	mock.Mock
}

func (m *MockDocumentProvider) GetDocumentContent(_ context.Context, docID int64) (string, string, error) {
	args := m.Called(docID)
	return args.String(0), args.String(1), args.Error(2)
}

// ============================================================
// Helper to build a ChatUseCase with mocks
// ============================================================

type chatTestFixture struct {
	convRepo    *MockConversationRepo
	msgRepo     *MockMessageRepo
	embRepo     *MockEmbeddingRepo
	embProvider *MockEmbeddingProvider
	docProvider *MockDocumentProvider
	llmProvider *MockLLMProvider
	embUseCase  *EmbeddingUseCase
	chatUseCase *ChatUseCase
}

func newChatTestFixture(opts ...ChatUseCaseOptions) *chatTestFixture {
	f := &chatTestFixture{
		convRepo:    new(MockConversationRepo),
		msgRepo:     new(MockMessageRepo),
		embRepo:     new(MockEmbeddingRepo),
		embProvider: new(MockEmbeddingProvider),
		docProvider: new(MockDocumentProvider),
		llmProvider: new(MockLLMProvider),
	}
	f.embUseCase = NewEmbeddingUseCase(f.embRepo, f.embProvider, f.docProvider, nil, "test-model")
	f.chatUseCase = NewChatUseCase(
		f.convRepo, f.msgRepo, f.embRepo, f.embUseCase,
		f.llmProvider, &mockPersonalityProvider{}, nil,
		opts...,
	)
	return f
}

// ============================================================
// Tests: NewChatUseCase
// ============================================================

func TestNewChatUseCase_Defaults(t *testing.T) {
	f := newChatTestFixture()
	assert.Equal(t, "llm", f.chatUseCase.modelName)
	assert.Equal(t, 10, f.chatUseCase.searchTopK)
	assert.Equal(t, 0.7, f.chatUseCase.searchThreshold)
	assert.Nil(t, f.chatUseCase.moodUseCase)
}

func TestNewChatUseCase_WithOptions(t *testing.T) {
	f := newChatTestFixture(ChatUseCaseOptions{
		ModelName:       "gpt-4",
		SearchTopK:      5,
		SearchThreshold: 0.9,
	})
	assert.Equal(t, "gpt-4", f.chatUseCase.modelName)
	assert.Equal(t, 5, f.chatUseCase.searchTopK)
	assert.Equal(t, 0.9, f.chatUseCase.searchThreshold)
}

func TestNewChatUseCase_WithMoodUseCase(t *testing.T) {
	moodUC := &MoodUseCase{}
	f := newChatTestFixture(ChatUseCaseOptions{MoodUseCase: moodUC})
	assert.NotNil(t, f.chatUseCase.moodUseCase)
}

// ============================================================
// Tests: Chat
// ============================================================

func TestChat_NewConversation_Success(t *testing.T) {
	f := newChatTestFixture()
	ctx := context.Background()
	userID := int64(1)
	req := &dto.SendMessageRequest{Content: "Hello"}

	// Create conversation
	f.convRepo.On("Create", mock.AnythingOfType("*entities.Conversation")).Return(nil).Run(func(args mock.Arguments) {
		c := args.Get(0).(*entities.Conversation)
		c.ID = 10
	})
	// Create user message
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	// Get message history
	f.msgRepo.On("GetByConversationID", int64(10), 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "Hello"}}, false, nil,
	)
	// Search similar (embedding provider for cached query)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1, 0.2}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	// LLM response
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Hi there!", 50, nil)

	resp, err := f.chatUseCase.Chat(ctx, userID, req)
	require.NoError(t, err)
	assert.Equal(t, int64(10), resp.ConversationID)
	assert.Equal(t, "Hi there!", resp.Message.Content)
	f.convRepo.AssertExpectations(t)
	f.msgRepo.AssertExpectations(t)
}

func TestChat_ExistingConversation_Success(t *testing.T) {
	f := newChatTestFixture()
	ctx := context.Background()
	userID := int64(1)
	convID := int64(20)
	req := &dto.SendMessageRequest{Content: "Follow up", ConversationID: &convID}

	f.convRepo.On("GetByID", convID).Return(&entities.Conversation{ID: convID, UserID: userID}, nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 101
		}
	})
	f.msgRepo.On("GetByConversationID", convID, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "Follow up"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Response", 30, nil)

	resp, err := f.chatUseCase.Chat(ctx, userID, req)
	require.NoError(t, err)
	assert.Equal(t, convID, resp.ConversationID)
}

func TestChat_UnauthorizedConversation(t *testing.T) {
	f := newChatTestFixture()
	ctx := context.Background()
	convID := int64(20)
	req := &dto.SendMessageRequest{Content: "test", ConversationID: &convID}

	f.convRepo.On("GetByID", convID).Return(&entities.Conversation{ID: convID, UserID: 999}, nil)

	resp, err := f.chatUseCase.Chat(ctx, int64(1), req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "unauthorized access to conversation")
}

func TestChat_GetConversationError(t *testing.T) {
	f := newChatTestFixture()
	convID := int64(20)
	req := &dto.SendMessageRequest{Content: "test", ConversationID: &convID}

	f.convRepo.On("GetByID", convID).Return(nil, errors.New("db error"))

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversation")
}

func TestChat_CreateConversationError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(errors.New("db error"))

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create conversation")
}

func TestChat_CreateUserMessageError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(errors.New("db error")).Once()

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create user message")
}

func TestChat_GetMessageHistoryError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil)
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(nil, false, errors.New("db error"))

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get message history")
}

func TestChat_LLMError_SavesErrorMessage(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil)
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("", 0, errors.New("LLM timeout"))

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to generate response")
}

func TestChat_WithSources(t *testing.T) {
	f := newChatTestFixture()
	ctx := context.Background()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)

	pageNum := 1
	sources := []entities.ChunkWithScore{
		{
			Chunk:           &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "chunk text", PageNumber: &pageNum},
			DocumentTitle:   "Doc Title",
			SimilarityScore: 0.95,
		},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Answer", 40, nil)
	f.msgRepo.On("CreateMessageSource", int64(100), int64(1), mock.AnythingOfType("float64")).Return(nil)

	resp, err := f.chatUseCase.Chat(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "Answer", resp.Message.Content)
	assert.Len(t, resp.Message.Sources, 1)
}

func TestChat_SearchError_ContinuesWithoutContext(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return(nil, errors.New("embedding error"))
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Answer", 40, nil)

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.Equal(t, "Answer", resp.Message.Content)
}

func TestChat_TruncatesLongTitle(t *testing.T) {
	f := newChatTestFixture()
	longContent := "This is a very long message that exceeds fifty characters and should be truncated"
	req := &dto.SendMessageRequest{Content: longContent}

	var capturedConv *entities.Conversation
	f.convRepo.On("Create", mock.AnythingOfType("*entities.Conversation")).Return(nil).Run(func(args mock.Arguments) {
		capturedConv = args.Get(0).(*entities.Conversation)
	})
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: longContent}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("R", 10, nil)

	_, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.True(t, len(capturedConv.Title) <= 53) // 50 + "..."
	assert.Contains(t, capturedConv.Title, "...")
}

func TestChat_MaxSourcesFromRequest(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test", MaxSources: 3}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	// Verify that limit=3 is passed (from req.MaxSources)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, 3, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("R", 10, nil)

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	f.embRepo.AssertExpectations(t)
}

func TestChat_CreateAssistantMessageError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	// First Create (user message) succeeds, second (assistant) fails
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Once()
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("answer", 10, nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(errors.New("db error")).Once()

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create assistant message")
}

func TestChat_CreateMessageSourceError_ContinuesGracefully(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	sources := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "text"}, SimilarityScore: 0.9},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)
	f.msgRepo.On("CreateMessageSource", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("source error"))

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// ============================================================
// Tests: ChatStream
// ============================================================

func TestChatStream_NewConversation_Success(t *testing.T) {
	f := newChatTestFixture()
	ctx := context.Background()
	req := &dto.SendMessageRequest{Content: "Hello stream"}
	var chunks []string
	onChunk := func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	}

	f.convRepo.On("Create", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		c := args.Get(0).(*entities.Conversation)
		c.ID = 10
	})
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", int64(10), 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "Hello stream"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Streamed!", 30, nil)

	resp, err := f.chatUseCase.ChatStream(ctx, 1, req, onChunk)
	require.NoError(t, err)
	assert.Equal(t, int64(10), resp.ConversationID)
	assert.Equal(t, "Streamed!", resp.Message.Content)
}

func TestChatStream_ExistingConversation_Unauthorized(t *testing.T) {
	f := newChatTestFixture()
	convID := int64(20)
	req := &dto.SendMessageRequest{Content: "test", ConversationID: &convID}

	f.convRepo.On("GetByID", convID).Return(&entities.Conversation{ID: convID, UserID: 999}, nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "unauthorized access to conversation")
}

func TestChatStream_GetConversationError(t *testing.T) {
	f := newChatTestFixture()
	convID := int64(20)
	req := &dto.SendMessageRequest{Content: "test", ConversationID: &convID}

	f.convRepo.On("GetByID", convID).Return(nil, errors.New("db error"))

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversation")
}

func TestChatStream_CreateConversationError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(errors.New("db error"))

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create conversation")
}

func TestChatStream_CreateUserMessageError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(errors.New("db error")).Once()

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create user message")
}

func TestChatStream_GetMessageHistoryError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil)
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(nil, false, errors.New("db error"))

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get message history")
}

func TestChatStream_LLMStreamError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil)
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", 0, errors.New("stream error"))

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to generate response")
}

func TestChatStream_CreateAssistantMessageError(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	// First Create succeeds (user msg), second fails (assistant msg)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Once()
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[]entities.ChunkWithScore{}, nil,
	)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("answer", 10, nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(errors.New("db error")).Once()

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create assistant message")
}

func TestChatStream_WithSources(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	sources := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "chunk"}, SimilarityScore: 0.9, DocumentTitle: "Doc"},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)
	f.msgRepo.On("CreateMessageSource", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.Len(t, resp.Message.Sources, 1)
}

func TestChatStream_SearchError_ContinuesWithoutContext(t *testing.T) {
	f := newChatTestFixture()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return(nil, errors.New("embed err"))
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// ============================================================
// Tests: GetConversations
// ============================================================

func TestGetConversations_Success(t *testing.T) {
	f := newChatTestFixture()
	convs := []entities.Conversation{
		{ID: 1, UserID: 1, Title: "Conv1"},
		{ID: 2, UserID: 1, Title: "Conv2"},
	}
	f.convRepo.On("GetByUserID", int64(1), 20, 0).Return(convs, 2, nil)

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "", 20, 0)
	require.NoError(t, err)
	assert.Len(t, resp.Conversations, 2)
	assert.Equal(t, 2, resp.Total)
}

func TestGetConversations_WithSearch(t *testing.T) {
	f := newChatTestFixture()
	convs := []entities.Conversation{
		{ID: 1, UserID: 1, Title: "Found"},
	}
	f.convRepo.On("Search", int64(1), "test", 20, 0).Return(convs, 1, nil)

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "test", 20, 0)
	require.NoError(t, err)
	assert.Len(t, resp.Conversations, 1)
}

func TestGetConversations_DefaultLimit(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByUserID", int64(1), 20, 0).Return([]entities.Conversation{}, 0, nil)

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.Limit)
}

func TestGetConversations_OverLimit(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByUserID", int64(1), 20, 0).Return([]entities.Conversation{}, 0, nil)

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "", 200, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.Limit)
}

func TestGetConversations_NegativeOffset(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByUserID", int64(1), 20, 0).Return([]entities.Conversation{}, 0, nil)

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "", 20, -5)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Offset)
}

func TestGetConversations_Error(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByUserID", int64(1), 20, 0).Return(nil, 0, errors.New("db error"))

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "", 20, 0)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversations")
}

func TestGetConversations_SearchError(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("Search", int64(1), "q", 20, 0).Return(nil, 0, errors.New("db error"))

	resp, err := f.chatUseCase.GetConversations(context.Background(), 1, "q", 20, 0)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversations")
}

// ============================================================
// Tests: GetConversation
// ============================================================

func TestGetConversation_Success(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1, Title: "Test"}, nil)

	resp, err := f.chatUseCase.GetConversation(context.Background(), 1, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), resp.ID)
}

func TestGetConversation_NotFound(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(nil, errors.New("not found"))

	resp, err := f.chatUseCase.GetConversation(context.Background(), 1, 5)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversation")
}

func TestGetConversation_Unauthorized(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 999}, nil)

	resp, err := f.chatUseCase.GetConversation(context.Background(), 1, 5)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "unauthorized access to conversation")
}

// ============================================================
// Tests: UpdateConversation
// ============================================================

func TestUpdateConversation_Success(t *testing.T) {
	f := newChatTestFixture()
	conv := &entities.Conversation{ID: 5, UserID: 1, Title: "Old"}
	f.convRepo.On("GetByID", int64(5)).Return(conv, nil)
	f.convRepo.On("Update", mock.AnythingOfType("*entities.Conversation")).Return(nil)

	req := &dto.UpdateConversationRequest{Title: "New Title"}
	resp, err := f.chatUseCase.UpdateConversation(context.Background(), 1, 5, req)
	require.NoError(t, err)
	assert.Equal(t, "New Title", resp.Title)
}

func TestUpdateConversation_NotFound(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(nil, errors.New("not found"))

	resp, err := f.chatUseCase.UpdateConversation(context.Background(), 1, 5, &dto.UpdateConversationRequest{Title: "X"})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversation")
}

func TestUpdateConversation_Unauthorized(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 999}, nil)

	resp, err := f.chatUseCase.UpdateConversation(context.Background(), 1, 5, &dto.UpdateConversationRequest{Title: "X"})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "unauthorized access to conversation")
}

func TestUpdateConversation_UpdateError(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.convRepo.On("Update", mock.Anything).Return(errors.New("db error"))

	resp, err := f.chatUseCase.UpdateConversation(context.Background(), 1, 5, &dto.UpdateConversationRequest{Title: "X"})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to update conversation")
}

// ============================================================
// Tests: DeleteConversation
// ============================================================

func TestDeleteConversation_Success(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("DeleteByConversationID", int64(5)).Return(nil)
	f.convRepo.On("Delete", int64(5)).Return(nil)

	err := f.chatUseCase.DeleteConversation(context.Background(), 1, 5)
	assert.NoError(t, err)
}

func TestDeleteConversation_NotFound(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(nil, errors.New("not found"))

	err := f.chatUseCase.DeleteConversation(context.Background(), 1, 5)
	assert.ErrorContains(t, err, "failed to get conversation")
}

func TestDeleteConversation_Unauthorized(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 999}, nil)

	err := f.chatUseCase.DeleteConversation(context.Background(), 1, 5)
	assert.ErrorContains(t, err, "unauthorized access to conversation")
}

func TestDeleteConversation_DeleteMessagesError(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("DeleteByConversationID", int64(5)).Return(errors.New("db error"))

	err := f.chatUseCase.DeleteConversation(context.Background(), 1, 5)
	assert.ErrorContains(t, err, "failed to delete messages")
}

func TestDeleteConversation_DeleteConversationError(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("DeleteByConversationID", int64(5)).Return(nil)
	f.convRepo.On("Delete", int64(5)).Return(errors.New("db error"))

	err := f.chatUseCase.DeleteConversation(context.Background(), 1, 5)
	assert.ErrorContains(t, err, "failed to delete conversation")
}

// ============================================================
// Tests: GetMessages
// ============================================================

func TestGetMessages_Success(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	messages := []entities.Message{
		{ID: 1, Role: entities.MessageRoleUser, Content: "Hi"},
		{ID: 2, Role: entities.MessageRoleAssistant, Content: "Hello"},
	}
	f.msgRepo.On("GetByConversationID", int64(5), 50, (*int64)(nil)).Return(messages, false, nil)
	f.msgRepo.On("GetMessageSources", int64(2)).Return([]entities.MessageSource{}, nil)

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 50, nil)
	require.NoError(t, err)
	assert.Len(t, resp.Messages, 2)
	assert.False(t, resp.HasMore)
}

func TestGetMessages_DefaultLimit(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("GetByConversationID", int64(5), 50, (*int64)(nil)).Return([]entities.Message{}, false, nil)

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 0, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetMessages_OverLimit(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("GetByConversationID", int64(5), 50, (*int64)(nil)).Return([]entities.Message{}, false, nil)

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 200, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetMessages_Unauthorized(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 999}, nil)

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 50, nil)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "unauthorized access to conversation")
}

func TestGetMessages_ConversationError(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(nil, errors.New("not found"))

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 50, nil)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get conversation")
}

func TestGetMessages_GetMessagesError(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("GetByConversationID", int64(5), 50, (*int64)(nil)).Return(nil, false, errors.New("db error"))

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 50, nil)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get messages")
}

func TestGetMessages_SourcesError_ContinuesGracefully(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	messages := []entities.Message{
		{ID: 2, Role: entities.MessageRoleAssistant, Content: "Hello"},
	}
	f.msgRepo.On("GetByConversationID", int64(5), 50, (*int64)(nil)).Return(messages, false, nil)
	f.msgRepo.On("GetMessageSources", int64(2)).Return(nil, errors.New("sources error"))

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 50, nil)
	require.NoError(t, err)
	assert.Len(t, resp.Messages, 1)
}

func TestGetMessages_WithBeforeID(t *testing.T) {
	f := newChatTestFixture()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	beforeID := int64(10)
	f.msgRepo.On("GetByConversationID", int64(5), 50, &beforeID).Return([]entities.Message{}, true, nil)

	resp, err := f.chatUseCase.GetMessages(context.Background(), 1, 5, 50, &beforeID)
	require.NoError(t, err)
	assert.True(t, resp.HasMore)
}

// ============================================================
// Tests: buildSearchQuery
// ============================================================

func TestBuildSearchQuery_Empty(t *testing.T) {
	result := buildSearchQuery(nil, "hello")
	assert.Equal(t, "hello", result)
}

func TestBuildSearchQuery_WithHistory(t *testing.T) {
	messages := []entities.Message{
		{Role: entities.MessageRoleUser, Content: "First question"},
		{Role: entities.MessageRoleAssistant, Content: "First answer"},
		{Role: entities.MessageRoleUser, Content: "Second question"},
	}
	result := buildSearchQuery(messages, "follow up")
	assert.Contains(t, result, "Second question")
	assert.Contains(t, result, "follow up")
}

func TestBuildSearchQuery_TruncatesLongMessages(t *testing.T) {
	longContent := ""
	for i := 0; i < 300; i++ {
		longContent += "x"
	}
	messages := []entities.Message{
		{Role: entities.MessageRoleUser, Content: longContent},
	}
	result := buildSearchQuery(messages, "query")
	// The content should be truncated to 200 runes
	assert.Contains(t, result, "query")
	// Should not contain the full 300-char content
	assert.Less(t, len(result), 300+len("query")+10)
}

func TestBuildSearchQuery_SkipsSystemMessages(t *testing.T) {
	messages := []entities.Message{
		{Role: entities.MessageRoleSystem, Content: "system prompt"},
		{Role: entities.MessageRoleUser, Content: "user msg"},
	}
	result := buildSearchQuery(messages, "query")
	assert.Contains(t, result, "user msg")
	assert.Contains(t, result, "query")
}

// ============================================================
// Tests: moodForPrompt
// ============================================================

func TestMoodForPrompt_Nil(t *testing.T) {
	result := moodForPrompt(nil)
	assert.Equal(t, entities.MoodContent, result.State)
}

func TestMoodForPrompt_WithValue(t *testing.T) {
	mood := &entities.MoodContext{State: entities.MoodHappy, Intensity: 0.9}
	result := moodForPrompt(mood)
	assert.Equal(t, entities.MoodHappy, result.State)
	assert.Equal(t, 0.9, result.Intensity)
}

// ============================================================
// Tests: Chat and ChatStream with AuditLogger (covering logger branches)
// ============================================================

func newAuditLogger() *logging.AuditLogger {
	logger := logging.NewLogger("error") // suppress output
	return logging.NewAuditLogger(logger)
}

func newChatTestFixtureWithAudit(opts ...ChatUseCaseOptions) *chatTestFixture {
	f := &chatTestFixture{
		convRepo:    new(MockConversationRepo),
		msgRepo:     new(MockMessageRepo),
		embRepo:     new(MockEmbeddingRepo),
		embProvider: new(MockEmbeddingProvider),
		docProvider: new(MockDocumentProvider),
		llmProvider: new(MockLLMProvider),
	}
	auditLogger := newAuditLogger()
	f.embUseCase = NewEmbeddingUseCase(f.embRepo, f.embProvider, f.docProvider, auditLogger, "test-model")
	f.chatUseCase = NewChatUseCase(
		f.convRepo, f.msgRepo, f.embRepo, f.embUseCase,
		f.llmProvider, &mockPersonalityProvider{}, auditLogger,
		opts...,
	)
	return f
}

func TestChat_WithAuditLogger_SearchError(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	// Search fails => audit logger logs warning
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return(nil, errors.New("embed error"))
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChat_WithAuditLogger_SourceError(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	sources := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "text"}, SimilarityScore: 0.9},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)
	f.msgRepo.On("CreateMessageSource", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("source error"))

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChat_WithAuditLogger_Success(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ChunkWithScore{}, nil)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteConversation_WithAuditLogger(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	f.convRepo.On("GetByID", int64(5)).Return(&entities.Conversation{ID: 5, UserID: 1}, nil)
	f.msgRepo.On("DeleteByConversationID", int64(5)).Return(nil)
	f.convRepo.On("Delete", int64(5)).Return(nil)

	err := f.chatUseCase.DeleteConversation(context.Background(), 1, 5)
	assert.NoError(t, err)
}

// ============================================================
// Tests: Chat with MoodUseCase integration
// ============================================================

func TestChat_WithMoodUseCase(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 100, PreviousTotal: 100},
	}
	analyticsRepo := &mockAnalyticsRepo{}
	pp := &mockPersonalityProvider{}
	moodUC := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	f := newChatTestFixtureWithAudit(ChatUseCaseOptions{MoodUseCase: moodUC})
	req := &dto.SendMessageRequest{Content: "Hello with mood"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "Hello with mood"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ChunkWithScore{}, nil)
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).Return("Hi!", 10, nil)

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChatStream_WithMoodUseCase(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 100, PreviousTotal: 100},
	}
	analyticsRepo := &mockAnalyticsRepo{}
	pp := &mockPersonalityProvider{}
	moodUC := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	f := newChatTestFixtureWithAudit(ChatUseCaseOptions{MoodUseCase: moodUC})
	req := &dto.SendMessageRequest{Content: "Stream with mood"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "Stream with mood"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ChunkWithScore{}, nil)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Streamed!", 10, nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChatStream_WithAuditLogger_SearchError(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return(nil, errors.New("embed error"))
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChatStream_WithAuditLogger_SourceError(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	sources := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "text"}, SimilarityScore: 0.9},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)
	f.msgRepo.On("CreateMessageSource", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("source error"))

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestChatStream_WithAuditLogger_Success(t *testing.T) {
	f := newChatTestFixtureWithAudit()
	req := &dto.SendMessageRequest{Content: "test"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test"}}, false, nil,
	)
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ChunkWithScore{}, nil)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Answer", 10, nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// ============================================================
// Tests: Chat with RAG context injection (covering contextText != "" branch)
// ============================================================

func TestChat_WithRAGContextInjection(t *testing.T) {
	// Use a personality provider that returns non-empty RAG context
	f := &chatTestFixture{
		convRepo:    new(MockConversationRepo),
		msgRepo:     new(MockMessageRepo),
		embRepo:     new(MockEmbeddingRepo),
		embProvider: new(MockEmbeddingProvider),
		docProvider: new(MockDocumentProvider),
		llmProvider: new(MockLLMProvider),
	}
	f.embUseCase = NewEmbeddingUseCase(f.embRepo, f.embProvider, f.docProvider, nil, "test-model")
	pp := &ragPersonalityProvider{}
	f.chatUseCase = NewChatUseCase(
		f.convRepo, f.msgRepo, f.embRepo, f.embUseCase,
		f.llmProvider, pp, nil,
	)
	req := &dto.SendMessageRequest{Content: "test question"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test question"}}, false, nil,
	)
	sources := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "relevant info"}, SimilarityScore: 0.9},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	// Verify the LLM receives messages with injected context
	f.llmProvider.On("GenerateResponse", mock.Anything, mock.MatchedBy(func(msgs []entities.Message) bool {
		for _, m := range msgs {
			if m.Role == entities.MessageRoleUser && len(m.Content) > len("test question") {
				return true
			}
		}
		return false
	}), mock.Anything).Return("Answer with context", 20, nil)
	f.msgRepo.On("CreateMessageSource", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := f.chatUseCase.Chat(context.Background(), 1, req)
	require.NoError(t, err)
	assert.Equal(t, "Answer with context", resp.Message.Content)
}

// ragPersonalityProvider returns non-empty RAG context
type ragPersonalityProvider struct{}

func (r *ragPersonalityProvider) BuildSystemPrompt(_ entities.MoodContext) string { return "system" }
func (r *ragPersonalityProvider) FormatRAGContext(sources []entities.ChunkWithScore) string {
	if len(sources) == 0 {
		return ""
	}
	return "RAG CONTEXT: relevant document data"
}
func (r *ragPersonalityProvider) GetGreeting(_ string) string { return "Hi!" }
func (r *ragPersonalityProvider) GetMoodComment(_ entities.MoodContext) string {
	return "Mood comment"
}
func (r *ragPersonalityProvider) FormatNotification(_, title, msg string, _ entities.MoodContext) string {
	return title + "\n" + msg
}

func TestChatStream_WithRAGContextInjection(t *testing.T) {
	f := &chatTestFixture{
		convRepo:    new(MockConversationRepo),
		msgRepo:     new(MockMessageRepo),
		embRepo:     new(MockEmbeddingRepo),
		embProvider: new(MockEmbeddingProvider),
		docProvider: new(MockDocumentProvider),
		llmProvider: new(MockLLMProvider),
	}
	f.embUseCase = NewEmbeddingUseCase(f.embRepo, f.embProvider, f.docProvider, nil, "test-model")
	pp := &ragPersonalityProvider{}
	f.chatUseCase = NewChatUseCase(
		f.convRepo, f.msgRepo, f.embRepo, f.embUseCase,
		f.llmProvider, pp, nil,
	)
	req := &dto.SendMessageRequest{Content: "test question"}

	f.convRepo.On("Create", mock.Anything).Return(nil)
	f.msgRepo.On("Create", mock.AnythingOfType("*entities.Message")).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(0).(*entities.Message)
		if m.Role == entities.MessageRoleAssistant {
			m.ID = 100
		}
	})
	f.msgRepo.On("GetByConversationID", mock.Anything, 10, (*int64)(nil)).Return(
		[]entities.Message{{Role: entities.MessageRoleUser, Content: "test question"}}, false, nil,
	)
	sources := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "data"}, SimilarityScore: 0.9},
	}
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sources, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)
	f.llmProvider.On("GenerateResponseStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("Streamed with context", 20, nil)
	f.msgRepo.On("CreateMessageSource", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := f.chatUseCase.ChatStream(context.Background(), 1, req, func(string) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, "Streamed with context", resp.Message.Content)
}
