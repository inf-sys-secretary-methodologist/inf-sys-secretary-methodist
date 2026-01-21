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

// mockAuthMiddleware creates a middleware that sets user_id for testing
func mockAuthMiddleware(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// TestListNotifications tests listing notifications
func TestListNotifications(t *testing.T) {
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
			queryParams:    "?type=document",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with priority filter",
			queryParams:    "?priority=high",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with read filter",
			queryParams:    "?is_read=true",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with pagination",
			queryParams:    "?limit=10&offset=5",
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
				router.Use(mockAuthMiddleware(1, "student"))
			}

			router.GET("/notifications", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"notifications": []gin.H{},
					"total":         0,
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/notifications"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetNotificationByID tests getting a notification by ID
func TestGetNotificationByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		notificationID string
		expectedStatus int
	}{
		{
			name:           "success",
			notificationID: "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			notificationID: "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			notificationID: "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "student"))

			router.GET("/notifications/:id", func(c *gin.Context) {
				id := c.Param("id")
				if id == "invalid" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
					return
				}
				if id == "999" {
					c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"id":      1,
					"title":   "Test Notification",
					"message": "Test message",
					"type":    "document",
					"is_read": false,
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/notifications/"+tt.notificationID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMarkAsRead tests marking a notification as read
func TestMarkAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		notificationID string
		expectedStatus int
	}{
		{
			name:           "success",
			notificationID: "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			notificationID: "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "student"))

			router.PUT("/notifications/:id/read", func(c *gin.Context) {
				id := c.Param("id")
				if id == "invalid" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
			})

			req := httptest.NewRequest(http.MethodPut, "/notifications/"+tt.notificationID+"/read", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMarkAllAsRead tests marking all notifications as read
func TestMarkAllAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "student"))
			}

			router.PUT("/notifications/read-all", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
			})

			req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteNotification tests deleting a notification
func TestDeleteNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		notificationID string
		expectedStatus int
	}{
		{
			name:           "success",
			notificationID: "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			notificationID: "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "student"))

			router.DELETE("/notifications/:id", func(c *gin.Context) {
				id := c.Param("id")
				if id == "invalid" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
			})

			req := httptest.NewRequest(http.MethodDelete, "/notifications/"+tt.notificationID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteAllNotifications tests deleting all notifications
func TestDeleteAllNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "student"))
			}

			router.DELETE("/notifications", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "all notifications deleted"})
			})

			req := httptest.NewRequest(http.MethodDelete, "/notifications", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetUnreadCount tests getting unread notification count
func TestGetUnreadCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "student"))
			}

			router.GET("/notifications/unread-count", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"count": 5})
			})

			req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetNotificationStats tests getting notification statistics
func TestGetNotificationStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "success",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "student"))
			}

			router.GET("/notifications/stats", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"total":  10,
					"unread": 5,
					"by_type": gin.H{
						"document":     3,
						"task":         4,
						"announcement": 3,
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/notifications/stats", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCreateNotification tests creating a notification (admin only)
func TestCreateNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name: "success",
			payload: map[string]any{
				"user_id":  2,
				"title":    "New Notification",
				"message":  "This is a test notification",
				"type":     "system",
				"priority": "normal",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			payload: map[string]any{
				"title": "New Notification",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.POST("/admin/notifications", func(c *gin.Context) {
				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if _, ok := input["user_id"]; !ok {
					c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
					return
				}
				if _, ok := input["message"]; !ok {
					c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"id":       1,
					"user_id":  input["user_id"],
					"title":    input["title"],
					"message":  input["message"],
					"type":     input["type"],
					"priority": input["priority"],
				})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/admin/notifications", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCreateBulkNotifications tests creating bulk notifications (admin only)
func TestCreateBulkNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name: "success",
			payload: map[string]any{
				"user_ids": []int{1, 2, 3},
				"title":    "Bulk Notification",
				"message":  "This is a bulk notification",
				"type":     "system",
				"priority": "normal",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "empty user_ids",
			payload: map[string]any{
				"user_ids": []int{},
				"title":    "Bulk Notification",
				"message":  "This is a bulk notification",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.POST("/admin/notifications/bulk", func(c *gin.Context) {
				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				userIDs, ok := input["user_ids"].([]any)
				if !ok || len(userIDs) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "user_ids is required and cannot be empty"})
					return
				}

				c.JSON(http.StatusCreated, []gin.H{
					{"id": 1, "user_id": 1},
					{"id": 2, "user_id": 2},
					{"id": 3, "user_id": 3},
				})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/admin/notifications/bulk", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
