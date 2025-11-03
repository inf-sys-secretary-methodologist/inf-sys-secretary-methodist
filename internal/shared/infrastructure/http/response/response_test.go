package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccess(t *testing.T) {
	data := map[string]string{"message": "test"}
	resp := Success(data)

	assert.True(t, resp.Success)
	assert.Equal(t, data, resp.Data)
	assert.Nil(t, resp.Error)
	assert.NotEmpty(t, resp.Meta.Timestamp)
}

func TestList(t *testing.T) {
	items := []string{"item1", "item2", "item3"}
	pagination := Pagination{
		Page:       1,
		PerPage:    10,
		Total:      3,
		TotalPages: 1,
	}

	resp := List(items, pagination)

	assert.True(t, resp.Success)
	assert.Equal(t, items, resp.Data)
	assert.NotNil(t, resp.Meta.Pagination)
	assert.Equal(t, 1, resp.Meta.Pagination.Page)
	assert.Equal(t, 3, resp.Meta.Pagination.Total)
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse("TEST_ERROR", "Test error message")

	assert.False(t, resp.Success)
	assert.Nil(t, resp.Data)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "TEST_ERROR", resp.Error.Code)
	assert.Equal(t, "Test error message", resp.Error.Message)
}

func TestValidationErrorResponse(t *testing.T) {
	fields := map[string][]string{
		"email":    {"Email is required", "Invalid email format"},
		"password": {"Password must be at least 8 characters"},
	}

	resp := ValidationErrorResponse(fields)

	assert.False(t, resp.Success)
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	assert.NotNil(t, resp.Error.Details)

	details, ok := resp.Error.Details.(ValidationError)
	require.True(t, ok)
	assert.Equal(t, fields, details.Fields)
}

func TestNotFound(t *testing.T) {
	resp := NotFound("User")

	assert.False(t, resp.Success)
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	assert.Equal(t, "User not found", resp.Error.Message)
}

func TestUnauthorized(t *testing.T) {
	tests := []struct{
		name     string
		message  string
		expected string
	}{
		{
			name:     "with custom message",
			message:  "Invalid token",
			expected: "Invalid token",
		},
		{
			name:     "with empty message",
			message:  "",
			expected: "Unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := Unauthorized(tt.message)
			assert.False(t, resp.Success)
			assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
			assert.Equal(t, tt.expected, resp.Error.Message)
		})
	}
}

func TestForbidden(t *testing.T) {
	resp := Forbidden("You don't have permission")

	assert.False(t, resp.Success)
	assert.Equal(t, "FORBIDDEN", resp.Error.Code)
	assert.Equal(t, "You don't have permission", resp.Error.Message)
}

func TestInternalError(t *testing.T) {
	resp := InternalError("")

	assert.False(t, resp.Success)
	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
	assert.Equal(t, "Internal server error", resp.Error.Message)
}

func TestBadRequest(t *testing.T) {
	resp := BadRequest("Invalid JSON")

	assert.False(t, resp.Success)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
	assert.Equal(t, "Invalid JSON", resp.Error.Message)
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"test": "value"}

	err := WriteJSON(w, http.StatusOK, data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	err := WriteSuccess(w, data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestWriteCreated(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	err := WriteCreated(w, data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestWriteNoContent(t *testing.T) {
	w := httptest.NewRecorder()

	err := WriteNoContent(w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	errorResp := NotFound("User")

	err := WriteError(w, http.StatusNotFound, errorResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}

func TestWithRequestID(t *testing.T) {
	requestID := "req-123-456"
	resp := Success(map[string]string{"test": "data"})

	respWithID := WithRequestID(resp, requestID)

	assert.Equal(t, requestID, respWithID.Meta.RequestID)
}

func TestErrorWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field": "email",
		"value": "invalid",
	}

	resp := ErrorWithDetails("CUSTOM_ERROR", "Custom error message", details)

	assert.False(t, resp.Success)
	assert.Equal(t, "CUSTOM_ERROR", resp.Error.Code)
	assert.Equal(t, "Custom error message", resp.Error.Message)
	assert.Equal(t, details, resp.Error.Details)
}
