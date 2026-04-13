package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToConversationResponse(t *testing.T) {
	now := time.Now()
	lastMsg := now.Add(-time.Hour)
	conv := &entities.Conversation{
		ID:            1,
		UserID:        42,
		Title:         "Test Chat",
		Model:         "gpt-4",
		MessageCount:  5,
		LastMessageAt: &lastMsg,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	resp := ToConversationResponse(conv)

	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, int64(42), resp.UserID)
	assert.Equal(t, "Test Chat", resp.Title)
	assert.Equal(t, "gpt-4", resp.Model)
	assert.Equal(t, 5, resp.MessageCount)
	assert.Equal(t, &lastMsg, resp.LastMessageAt)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, now, resp.UpdatedAt)
}

func TestToConversationResponse_NilLastMessage(t *testing.T) {
	conv := &entities.Conversation{
		ID:    2,
		Title: "Empty",
	}

	resp := ToConversationResponse(conv)
	assert.Nil(t, resp.LastMessageAt)
}

func TestToMessageResponse(t *testing.T) {
	now := time.Now()
	tokens := 100
	model := "gpt-4"
	pageNum := 3

	msg := &entities.Message{
		ID:             10,
		ConversationID: 1,
		Role:           entities.MessageRoleAssistant,
		Content:        "Hello!",
		Sources: []entities.MessageSource{
			{
				ID:              1,
				DocumentID:      5,
				DocumentTitle:   "Doc Title",
				ChunkText:       "chunk text",
				SimilarityScore: 0.95,
				PageNumber:      &pageNum,
			},
		},
		TokensUsed:   &tokens,
		Model:        &model,
		ErrorMessage: nil,
		CreatedAt:    now,
	}

	resp := ToMessageResponse(msg)

	require.NotNil(t, resp)
	assert.Equal(t, int64(10), resp.ID)
	assert.Equal(t, "assistant", resp.Role)
	assert.Equal(t, "Hello!", resp.Content)
	assert.Equal(t, "complete", resp.Status)
	assert.Equal(t, &tokens, resp.TokensUsed)
	assert.Equal(t, &model, resp.Model)
	assert.Nil(t, resp.ErrorMessage)
	require.Len(t, resp.Sources, 1)
	assert.Equal(t, int64(5), resp.Sources[0].DocumentID)
	assert.Equal(t, "Doc Title", resp.Sources[0].DocumentTitle)
	assert.Equal(t, 0.95, resp.Sources[0].SimilarityScore)
	assert.Equal(t, &pageNum, resp.Sources[0].PageNumber)
}

func TestToMessageResponse_ErrorStatus(t *testing.T) {
	errMsg := "something went wrong"
	msg := &entities.Message{
		ID:           1,
		Role:         entities.MessageRoleAssistant,
		Content:      "",
		ErrorMessage: &errMsg,
		CreatedAt:    time.Now(),
	}

	resp := ToMessageResponse(msg)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, &errMsg, resp.ErrorMessage)
}

func TestToMessageResponse_NoSources(t *testing.T) {
	msg := &entities.Message{
		ID:        1,
		Role:      entities.MessageRoleUser,
		Content:   "Hi",
		CreatedAt: time.Now(),
	}

	resp := ToMessageResponse(msg)
	assert.Empty(t, resp.Sources)
}

func TestToSearchResultResponse(t *testing.T) {
	pageNum := 5
	chunk := &entities.ChunkWithScore{
		Chunk: &entities.DocumentChunk{
			DocumentID: 10,
			ChunkText:  "some text",
			PageNumber: &pageNum,
		},
		DocumentTitle:   "My Doc",
		SimilarityScore: 0.88,
	}

	resp := ToSearchResultResponse(chunk)

	require.NotNil(t, resp)
	assert.Equal(t, int64(10), resp.DocumentID)
	assert.Equal(t, "My Doc", resp.DocumentTitle)
	assert.Equal(t, "some text", resp.ChunkText)
	assert.Equal(t, 0.88, resp.SimilarityScore)
	assert.Equal(t, &pageNum, resp.PageNumber)
}

func TestToSearchResultResponse_NilPageNumber(t *testing.T) {
	chunk := &entities.ChunkWithScore{
		Chunk: &entities.DocumentChunk{
			DocumentID: 1,
			ChunkText:  "text",
		},
		DocumentTitle:   "Title",
		SimilarityScore: 0.5,
	}

	resp := ToSearchResultResponse(chunk)
	assert.Nil(t, resp.PageNumber)
}
