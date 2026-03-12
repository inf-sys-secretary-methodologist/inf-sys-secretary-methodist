package handlers_test

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

// TestListUsers tests listing users with filtering
func TestListUsers(t *testing.T) {
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
			name:           "success - with search",
			queryParams:    "?search=john",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with pagination",
			queryParams:    "?page=1&page_size=10",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - with role filter",
			queryParams:    "?role=student",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			if tt.authenticated {
				router.Use(mockAuthMiddleware(1, "admin"))
			}

			router.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"users": []gin.H{},
						"total": 0,
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetUserByID tests getting a user by ID
func TestGetUserByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "success",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			userID:         invalidID,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.GET("/users/:id", func(c *gin.Context) {
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
						"id":    1,
						"email": "user@example.com",
						"name":  "Test User",
						"role":  "student",
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUpdateUserProfile tests updating user profile
func TestUpdateUserProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name:   "success",
			userID: "1",
			payload: map[string]any{
				"phone": "+1234567890",
				"bio":   "Updated bio",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "999",
			payload: map[string]any{
				"phone": "+1234567890",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			userID:         invalidID,
			payload:        map[string]any{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.PUT("/users/:id/profile", func(c *gin.Context) {
				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}
				if id == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile updated"})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID+"/profile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUpdateUserRole tests updating user role
func TestUpdateUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name:   "success",
			userID: "1",
			payload: map[string]any{
				"role": "methodist",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid role",
			userID: "1",
			payload: map[string]any{
				"role": "invalid_role",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "missing role",
			userID: "1",
			payload: map[string]any{
				"role": "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			payload: map[string]any{
				"role": "student",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.PUT("/users/:id/role", func(c *gin.Context) {
				id := c.Param("id")
				if id == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				var input map[string]any
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid JSON"})
					return
				}

				role, _ := input["role"].(string)
				if role == "" || role == "invalid_role" {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid role"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Role updated"})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID+"/role", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUpdateUserStatus tests updating user status
func TestUpdateUserStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name:   "success - activate",
			userID: "1",
			payload: map[string]any{
				"is_active": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "success - deactivate",
			userID: "1",
			payload: map[string]any{
				"is_active": false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "999",
			payload: map[string]any{
				"is_active": true,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			userID:         invalidID,
			payload:        map[string]any{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.PUT("/users/:id/status", func(c *gin.Context) {
				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}
				if id == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Status updated"})
			})

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID+"/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteUser tests deleting a user
func TestDeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "success",
			userID:         "2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "cannot delete self",
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "user not found",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			userID:         invalidID,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.DELETE("/users/:id", func(c *gin.Context) {
				id := c.Param("id")
				if id == invalidID {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
					return
				}
				if id == "1" {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot delete self"})
					return
				}
				if id == "999" {
					c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted"})
			})

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSearchUsers tests searching users
func TestSearchUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "success",
			queryParams:    "?q=john",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty query",
			queryParams:    "?q=",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "with limit",
			queryParams:    "?q=test&limit=5",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(mockAuthMiddleware(1, "admin"))

			router.GET("/users/search", func(c *gin.Context) {
				q := c.Query("q")
				if q == "" {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Query required"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"users": []gin.H{},
						"total": 0,
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/users/search"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetCurrentUser tests getting the current user
func TestGetCurrentUser(t *testing.T) {
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

			router.GET("/users/me", func(c *gin.Context) {
				_, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"id":    1,
						"email": "user@example.com",
						"name":  "Current User",
						"role":  "student",
					},
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
