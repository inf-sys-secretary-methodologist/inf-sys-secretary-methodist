package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter creates a gin engine in test mode.
func setupRouter() *gin.Engine {
	return gin.New()
}

// withAuth returns middleware that sets user_id and role in context.
func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// withUserRole returns middleware that sets user_id and user_role in context.
func withUserRole(userID int64, userRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_role", userRole)
		c.Next()
	}
}

// performRequest executes an HTTP request against the router and returns the recorder.
func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// parseResponseBody parses the JSON response body into a map.
func parseResponseBody(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	return result
}
