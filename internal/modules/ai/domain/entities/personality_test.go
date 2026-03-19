package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTimeOfDay(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		expected string
	}{
		{"midnight", 0, "night"},
		{"early night", 5, "night"},
		{"morning start", 6, "morning"},
		{"late morning", 11, "morning"},
		{"afternoon start", 12, "afternoon"},
		{"late afternoon", 16, "afternoon"},
		{"evening start", 17, "evening"},
		{"late evening", 21, "evening"},
		{"night start", 22, "night"},
		{"before midnight", 23, "night"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimeOfDay(tt.hour)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoodStateConstants(t *testing.T) {
	assert.Equal(t, MoodState("happy"), MoodHappy)
	assert.Equal(t, MoodState("content"), MoodContent)
	assert.Equal(t, MoodState("worried"), MoodWorried)
	assert.Equal(t, MoodState("stressed"), MoodStressed)
	assert.Equal(t, MoodState("panicking"), MoodPanicking)
	assert.Equal(t, MoodState("relaxed"), MoodRelaxed)
	assert.Equal(t, MoodState("inspired"), MoodInspired)
}

func TestNewConversation(t *testing.T) {
	userID := int64(42)
	title := "Test Conversation"
	model := "gpt-4"

	conv := NewConversation(userID, title, model)

	assert.Equal(t, userID, conv.UserID)
	assert.Equal(t, title, conv.Title)
	assert.Equal(t, model, conv.Model)
	assert.Equal(t, 0, conv.MessageCount)
	assert.Nil(t, conv.LastMessageAt)
	assert.False(t, conv.CreatedAt.IsZero())
	assert.False(t, conv.UpdatedAt.IsZero())
}

func TestNewDocumentChunk(t *testing.T) {
	documentID := int64(10)
	index := 3
	text := "some chunk text"
	tokens := 150

	chunk := NewDocumentChunk(documentID, index, text, tokens)

	assert.Equal(t, documentID, chunk.DocumentID)
	assert.Equal(t, index, chunk.ChunkIndex)
	assert.Equal(t, text, chunk.ChunkText)
	assert.NotNil(t, chunk.ChunkTokens)
	assert.Equal(t, tokens, *chunk.ChunkTokens)
	assert.NotNil(t, chunk.Metadata)
	assert.False(t, chunk.CreatedAt.IsZero())
}

func TestNewEmbedding(t *testing.T) {
	chunkID := int64(5)
	vec := []float32{0.1, 0.2, 0.3}
	model := "text-embedding-3-small"

	emb := NewEmbedding(chunkID, vec, model)

	assert.Equal(t, chunkID, emb.ChunkID)
	assert.Equal(t, vec, emb.Embedding)
	assert.Equal(t, model, emb.Model)
	assert.False(t, emb.CreatedAt.IsZero())
}

func TestEmbeddingDimension(t *testing.T) {
	assert.Equal(t, 1536, EmbeddingDimension)
}

func TestIndexStatusConstants(t *testing.T) {
	assert.Equal(t, IndexStatus("pending"), IndexStatusPending)
	assert.Equal(t, IndexStatus("indexing"), IndexStatusIndexing)
	assert.Equal(t, IndexStatus("indexed"), IndexStatusIndexed)
	assert.Equal(t, IndexStatus("failed"), IndexStatusFailed)
}

func TestMessageRoleConstants(t *testing.T) {
	assert.Equal(t, MessageRole("user"), MessageRoleUser)
	assert.Equal(t, MessageRole("assistant"), MessageRoleAssistant)
	assert.Equal(t, MessageRole("system"), MessageRoleSystem)
}

func TestNewUserMessage(t *testing.T) {
	conversationID := int64(1)
	content := "Hello, AI!"

	msg := NewUserMessage(conversationID, content)

	assert.Equal(t, conversationID, msg.ConversationID)
	assert.Equal(t, MessageRoleUser, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Nil(t, msg.TokensUsed)
	assert.Nil(t, msg.Model)
	assert.False(t, msg.CreatedAt.IsZero())
}

func TestNewAssistantMessage(t *testing.T) {
	conversationID := int64(1)
	content := "Hello, human!"
	model := "gpt-4"
	tokensUsed := 42

	msg := NewAssistantMessage(conversationID, content, model, tokensUsed)

	assert.Equal(t, conversationID, msg.ConversationID)
	assert.Equal(t, MessageRoleAssistant, msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.NotNil(t, msg.Model)
	assert.Equal(t, model, *msg.Model)
	assert.NotNil(t, msg.TokensUsed)
	assert.Equal(t, tokensUsed, *msg.TokensUsed)
	assert.False(t, msg.CreatedAt.IsZero())
}

func TestMoodContext_Struct(t *testing.T) {
	ctx := MoodContext{
		State:             MoodHappy,
		Intensity:         0.8,
		Reason:            "all docs submitted",
		OverdueDocuments:  0,
		AtRiskStudents:    2,
		UpcomingDeadlines: 3,
		TimeOfDay:         "morning",
		DayOfWeek:         "Monday",
		AttendanceTrend:   "improving",
	}

	assert.Equal(t, MoodHappy, ctx.State)
	assert.Equal(t, 0.8, ctx.Intensity)
	assert.Equal(t, "all docs submitted", ctx.Reason)
	assert.Equal(t, 0, ctx.OverdueDocuments)
	assert.Equal(t, 2, ctx.AtRiskStudents)
	assert.Equal(t, 3, ctx.UpcomingDeadlines)
	assert.Equal(t, "morning", ctx.TimeOfDay)
	assert.Equal(t, "Monday", ctx.DayOfWeek)
	assert.Equal(t, "improving", ctx.AttendanceTrend)
}

func TestChunkWithScore_Struct(t *testing.T) {
	chunk := &DocumentChunk{ID: 1, ChunkText: "test"}
	cws := ChunkWithScore{
		Chunk:           chunk,
		DocumentTitle:   "Doc Title",
		SimilarityScore: 0.95,
	}

	assert.Equal(t, chunk, cws.Chunk)
	assert.Equal(t, "Doc Title", cws.DocumentTitle)
	assert.Equal(t, 0.95, cws.SimilarityScore)
}

func TestDocumentIndexStatus_Struct(t *testing.T) {
	dis := DocumentIndexStatus{
		DocumentID:  1,
		Status:      IndexStatusPending,
		ChunksCount: 5,
	}

	assert.Equal(t, int64(1), dis.DocumentID)
	assert.Equal(t, IndexStatusPending, dis.Status)
	assert.Equal(t, 5, dis.ChunksCount)
}

func TestMessageSource_Struct(t *testing.T) {
	ms := MessageSource{
		ID:              1,
		MessageID:       2,
		ChunkID:         3,
		DocumentID:      4,
		DocumentTitle:   "Title",
		ChunkText:       "text",
		SimilarityScore: 0.85,
	}

	assert.Equal(t, int64(1), ms.ID)
	assert.Equal(t, int64(2), ms.MessageID)
	assert.Equal(t, "Title", ms.DocumentTitle)
	assert.Equal(t, 0.85, ms.SimilarityScore)
}

func TestFunFact_Struct(t *testing.T) {
	ff := FunFact{
		ID:         1,
		Content:    "Fun fact content",
		Category:   "science",
		Source:     "Wikipedia",
		SourceURL:  "https://en.wikipedia.org",
		Language:   "en",
		IsApproved: true,
		UsedCount:  5,
	}

	assert.Equal(t, int64(1), ff.ID)
	assert.Equal(t, "Fun fact content", ff.Content)
	assert.Equal(t, "science", ff.Category)
	assert.Equal(t, "Wikipedia", ff.Source)
	assert.Equal(t, "https://en.wikipedia.org", ff.SourceURL)
	assert.Equal(t, "en", ff.Language)
	assert.True(t, ff.IsApproved)
	assert.Equal(t, 5, ff.UsedCount)
}
