package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestAIHandler() *AIHandler {
	return &AIHandler{}
}

func TestAIHandler_Chat_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.POST("/ai/chat", handler.Chat)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/chat", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestAIHandler_Chat_InvalidJSON(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.POST("/ai/chat", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Chat(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/chat", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestAIHandler_ChatStream_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/chat/stream", handler.ChatStream)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/chat/stream", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_ChatStream_MissingContent(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/chat/stream", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.ChatStream(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/chat/stream", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "content is required")
}

func TestAIHandler_ListConversations_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/conversations", handler.ListConversations)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/conversations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_CreateConversation_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.POST("/ai/conversations", handler.CreateConversation)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/conversations", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_GetConversation_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/conversations/:id", handler.GetConversation)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/conversations/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_GetConversation_InvalidID(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/conversations/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.GetConversation(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/conversations/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid conversation ID")
}

func TestAIHandler_UpdateConversation_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.PATCH("/ai/conversations/:id", handler.UpdateConversation)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/ai/conversations/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_UpdateConversation_InvalidID(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.PATCH("/ai/conversations/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.UpdateConversation(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/ai/conversations/abc", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_UpdateConversation_InvalidJSON(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.PATCH("/ai/conversations/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.UpdateConversation(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/ai/conversations/1", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_DeleteConversation_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.DELETE("/ai/conversations/:id", handler.DeleteConversation)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/ai/conversations/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_DeleteConversation_InvalidID(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.DELETE("/ai/conversations/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.DeleteConversation(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/ai/conversations/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_GetMessages_Unauthorized(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/conversations/:id/messages", handler.GetMessages)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/conversations/1/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAIHandler_GetMessages_InvalidID(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/conversations/:id/messages", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.GetMessages(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/conversations/abc/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_Search_InvalidJSON(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.POST("/ai/search", handler.Search)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/search", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_IndexDocument_InvalidID(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.POST("/ai/index/:documentId", handler.IndexDocument)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/index/abc", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_IndexDocumentsBatch_InvalidJSON(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.POST("/ai/index/batch", handler.IndexDocumentsBatch)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/index/batch", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAIHandler_GetMood_ServiceUnavailable(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/mood", handler.GetMood)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/mood", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "mood service not available")
}

func TestAIHandler_GetFact_ServiceUnavailable(t *testing.T) {
	handler := newTestAIHandler()
	r := gin.New()
	r.GET("/ai/fact", handler.GetFact)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai/fact", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "fun facts not available")
}
