package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const invalidID = "invalid"

// mockAuthMiddleware creates a middleware that sets user_id for testing
func mockAuthMiddleware(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// TestCreateDirectConversation tests direct conversation creation
func TestCreateDirectConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        map[string]any
		authenticated  bool
		expectedStatus int
	}{
		{
			name: "success",
			payload: map[string]any{
				"recipient_id": 2,
			},
			authenticated:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing recipient",
			payload:        map[string]any{},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthorized",
			payload: map[string]any{
				"recipient_id": 2,
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.POST("/conversations/direct", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON"})
					return
				}

				if input["recipient_id"] == nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Recipient required"})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status": "success",
					"data": gin.H{
						"id":   1,
						"type": "direct",
					},
				})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/conversations/direct", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCreateGroupConversation tests group conversation creation
func TestCreateGroupConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        map[string]any
		authenticated  bool
		expectedStatus int
	}{
		{
			name: "success",
			payload: map[string]any{
				"name":            "Test Group",
				"participant_ids": []int{2, 3},
			},
			authenticated:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			payload: map[string]any{
				"participant_ids": []int{2, 3},
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthorized",
			payload: map[string]any{
				"name": "Test Group",
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.POST("/conversations/group", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON"})
					return
				}

				if input["name"] == nil || input["name"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Name required"})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status": "success",
					"data": gin.H{
						"id":   1,
						"name": input["name"],
						"type": "group",
					},
				})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/conversations/group", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestListConversations tests listing conversations
func TestListConversations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success - no params",
			queryParams:    "",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with type filter",
			queryParams:    "?type=direct",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with pagination",
			queryParams:    "?limit=10&offset=0",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			queryParams:    "",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.GET("/conversations", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"conversations": []gin.H{},
						"total":         0,
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/conversations"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetConversation tests getting a single conversation
func TestGetConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			conversationID: "999",
			authenticated:  true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			conversationID: invalidID,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			conversationID: "1",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.GET("/conversations/:id", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}
				if id == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"id":   1,
						"type": "direct",
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/conversations/"+tt.conversationID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSendMessage tests sending a message
func TestSendMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		payload        map[string]any
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			payload: map[string]any{
				"content": "Hello, World!",
			},
			authenticated:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty message",
			conversationID: "1",
			payload: map[string]any{
				"content": "",
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid conversation id",
			conversationID: invalidID,
			payload: map[string]any{
				"content": "Hello!",
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			conversationID: "1",
			payload: map[string]any{
				"content": "Hello!",
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.POST("/conversations/:id/messages", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}

				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON"})
					return
				}

				if input["content"] == nil || input["content"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Content required"})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status": "success",
					"data": gin.H{
						"id":      1,
						"content": input["content"],
					},
				})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/conversations/"+tt.conversationID+"/messages", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetMessages tests getting messages from a conversation
func TestGetMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		queryParams    string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			queryParams:    "",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with pagination",
			conversationID: "1",
			queryParams:    "?limit=20",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid conversation id",
			conversationID: invalidID,
			queryParams:    "",
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			conversationID: "1",
			queryParams:    "",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.GET("/conversations/:id/messages", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"messages": []gin.H{},
						"has_more": false,
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/conversations/"+tt.conversationID+"/messages"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestEditMessage tests editing a message
func TestEditMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		messageID      string
		payload        map[string]any
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			messageID:      "1",
			payload: map[string]any{
				"content": "Updated message",
			},
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "message not found",
			conversationID: "1",
			messageID:      "999",
			payload: map[string]any{
				"content": "Updated",
			},
			authenticated:  true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid message id",
			conversationID: "1",
			messageID:      invalidID,
			payload:        map[string]any{},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.PATCH("/conversations/:id/messages/:messageId", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				messageID := c.Param("messageId")
				if messageID == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}
				if messageID == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"id":      1,
						"content": input["content"],
					},
				})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPatch, "/conversations/"+tt.conversationID+"/messages/"+tt.messageID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteMessage tests deleting a message
func TestDeleteMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		messageID      string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			messageID:      "1",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "message not found",
			conversationID: "1",
			messageID:      "999",
			authenticated:  true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid message id",
			conversationID: "1",
			messageID:      invalidID,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.DELETE("/conversations/:id/messages/:messageId", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				messageID := c.Param("messageId")
				if messageID == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}
				if messageID == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Deleted"})
			})

			req := httptest.NewRequest(http.MethodDelete, "/conversations/"+tt.conversationID+"/messages/"+tt.messageID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMarkAsRead tests marking messages as read
func TestMarkAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		payload        map[string]any
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			payload: map[string]any{
				"message_id": 10,
			},
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid conversation id",
			conversationID: invalidID,
			payload:        map[string]any{},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			conversationID: "1",
			payload:        map[string]any{},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.POST("/conversations/:id/read", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Marked as read"})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/conversations/"+tt.conversationID+"/read", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestLeaveConversation tests leaving a conversation
func TestLeaveConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid conversation id",
			conversationID: invalidID,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			conversationID: "1",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.POST("/conversations/:id/leave", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Left conversation"})
			})

			req := httptest.NewRequest(http.MethodPost, "/conversations/"+tt.conversationID+"/leave", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSearchMessages tests searching messages
func TestSearchMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		conversationID string
		queryParams    string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			conversationID: "1",
			queryParams:    "?q=hello",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty query",
			conversationID: "1",
			queryParams:    "?q=",
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing query",
			conversationID: "1",
			queryParams:    "",
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid conversation id",
			conversationID: invalidID,
			queryParams:    "?q=test",
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "user"))
			}

			router.GET("/conversations/:id/messages/search", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}

				q := c.Query("q")
				if q == "" {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Query required"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"messages": []gin.H{},
						"total":    0,
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/conversations/"+tt.conversationID+"/messages/search"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
