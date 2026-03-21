package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test-token")
	assert.NotNil(t, c)
	assert.Equal(t, "test-token", c.botToken)
	assert.Equal(t, BaseURL, c.baseURL)
}

func TestClient_SendMessage_EmptyChatID(t *testing.T) {
	c := NewClient("token")
	_, err := c.SendMessage(context.Background(), &SendMessageRequest{Text: "hello"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestClient_SendMessage_EmptyText(t *testing.T) {
	c := NewClient("token")
	_, err := c.SendMessage(context.Background(), &SendMessageRequest{ChatID: 123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text is required")
}

func TestClient_SendMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			OK:     true,
			Result: json.RawMessage(`{"message_id": 1, "chat": {"id": 123, "type": "private"}, "text": "hello"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	msg, err := c.SendMessage(context.Background(), &SendMessageRequest{ChatID: 123, Text: "hello"})
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, int64(1), msg.MessageID)
}

func TestClient_SendMessage_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: false, Description: "Bot blocked", ErrorCode: 403}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	_, err := c.SendMessage(context.Background(), &SendMessageRequest{ChatID: 123, Text: "hello"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telegram API error")
}

func TestClient_SendMessage_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	_, err := c.SendMessage(context.Background(), &SendMessageRequest{ChatID: 123, Text: "hello"})
	assert.Error(t, err)
}

func TestClient_GetMe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			OK:     true,
			Result: json.RawMessage(`{"id": 123, "is_bot": true, "first_name": "TestBot", "username": "testbot"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	info, err := c.GetMe(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(123), info.ID)
	assert.True(t, info.IsBot)
}

func TestClient_GetMe_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: false, Description: "Unauthorized"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	_, err := c.GetMe(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid bot token")
}

func TestClient_GetMe_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	_, err := c.GetMe(context.Background())
	assert.Error(t, err)
}

func TestClient_SetWebhook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.SetWebhook(context.Background(), "https://example.com/webhook", "secret")
	assert.NoError(t, err)
}

func TestClient_SetWebhook_NoSecretToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.SetWebhook(context.Background(), "https://example.com/webhook", "")
	assert.NoError(t, err)
}

func TestClient_SetWebhook_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: false, Description: "Bad webhook URL"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.SetWebhook(context.Background(), "invalid", "")
	assert.Error(t, err)
}

func TestClient_SetWebhook_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.SetWebhook(context.Background(), "https://example.com/webhook", "")
	assert.Error(t, err)
}

func TestClient_DeleteWebhook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.DeleteWebhook(context.Background())
	assert.NoError(t, err)
}

func TestClient_DeleteWebhook_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: false, Description: "Failed"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.DeleteWebhook(context.Background())
	assert.Error(t, err)
}

func TestClient_DeleteWebhook_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	err := c.DeleteWebhook(context.Background())
	assert.Error(t, err)
}

func TestClient_IsValidToken_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			OK:     true,
			Result: json.RawMessage(`{"id": 1, "is_bot": true, "first_name": "Bot", "username": "bot"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	assert.True(t, c.IsValidToken(context.Background()))
}

func TestClient_IsValidToken_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{OK: false, Description: "Unauthorized"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	assert.False(t, c.IsValidToken(context.Background()))
}

func TestClient_SendNotification_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			OK:     true,
			Result: json.RawMessage(`{"message_id": 1, "chat": {"id": 123, "type": "private"}, "text": "hello"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("token")
	c.baseURL = server.URL

	priorities := []string{"urgent", "high", "normal", "low", "unknown"}
	for _, p := range priorities {
		err := c.SendNotification(context.Background(), 123, "Test Title", "Test Message", p)
		assert.NoError(t, err)
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<script>", "&lt;script&gt;"},
		{"a > b & c < d", "a &gt; b &amp; c &lt; d"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, escapeHTML(tt.input))
		})
	}
}

func TestClient_SendMessage_ServerDown(t *testing.T) {
	c := NewClient("token")
	c.baseURL = "http://127.0.0.1:1" // unreachable

	_, err := c.SendMessage(context.Background(), &SendMessageRequest{ChatID: 123, Text: "hello"})
	assert.Error(t, err)
}

func TestClient_GetMe_ServerDown(t *testing.T) {
	c := NewClient("token")
	c.baseURL = "http://127.0.0.1:1"

	_, err := c.GetMe(context.Background())
	assert.Error(t, err)
}

func TestClient_SetWebhook_ServerDown(t *testing.T) {
	c := NewClient("token")
	c.baseURL = "http://127.0.0.1:1"

	err := c.SetWebhook(context.Background(), "https://example.com/webhook", "")
	assert.Error(t, err)
}

func TestClient_DeleteWebhook_ServerDown(t *testing.T) {
	c := NewClient("token")
	c.baseURL = "http://127.0.0.1:1"

	err := c.DeleteWebhook(context.Background())
	assert.Error(t, err)
}
