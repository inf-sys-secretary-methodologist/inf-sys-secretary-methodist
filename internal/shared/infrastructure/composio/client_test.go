package composio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test-api-key")
	assert.NotNil(t, c)
	assert.Equal(t, "test-api-key", c.apiKey)
	assert.Equal(t, BaseURL, c.baseURL)
	assert.NotNil(t, c.httpClient)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "https://backend.composio.dev/api", BaseURL)
	assert.Equal(t, "GMAIL_SEND_EMAIL", ActionGmailSendEmail)
	assert.Equal(t, "GMAIL_CREATE_EMAIL_DRAFT", ActionGmailCreateDraft)
	assert.Equal(t, "GMAIL_REPLY_TO_EMAIL_THREAD", ActionGmailReplyToThread)
	assert.Equal(t, "GMAIL_FETCH_EMAILS", ActionGmailFetchEmails)
	assert.Equal(t, "TELEGRAM_SEND_MESSAGE", ActionTelegramSendMessage)
	assert.Equal(t, "TELEGRAM_SEND_PHOTO", ActionTelegramSendPhoto)
	assert.Equal(t, "TELEGRAM_SEND_DOCUMENT", ActionTelegramSendDocument)
	assert.Equal(t, "SLACK_SENDS_A_MESSAGE_TO_A_SLACK_CHANNEL", ActionSlackSendMessage)
	assert.Equal(t, "SLACK_SEND_DIRECT_MESSAGE", ActionSlackSendDirectMessage)
}

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient("test-key")
	c.baseURL = srv.URL
	return srv, c
}

func TestExecuteAction_Success(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Contains(t, r.URL.Path, "/v2/actions/TEST_ACTION/execute")

		resp := ExecuteActionResponse{
			ExecutionID: "exec-123",
			Successful:  true,
			Data:        map[string]interface{}{"result": "ok"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	resp, err := c.ExecuteAction(ctx, "TEST_ACTION", &ExecuteActionRequest{
		EntityID: "entity-1",
		AppName:  "test",
		Input:    map[string]interface{}{"key": "value"},
	})

	require.NoError(t, err)
	assert.True(t, resp.Successful)
	assert.Equal(t, "exec-123", resp.ExecutionID)
}

func TestExecuteAction_HTTPError(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})

	ctx := context.Background()
	resp, err := c.ExecuteAction(ctx, "TEST_ACTION", &ExecuteActionRequest{
		Input: map[string]interface{}{},
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500")
}

func TestExecuteAction_ActionFailure(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		resp := ExecuteActionResponse{
			Successful: false,
			Error: &ErrorResponse{
				Message: "action failed",
				Code:    "ACTION_ERROR",
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	resp, err := c.ExecuteAction(ctx, "TEST_ACTION", &ExecuteActionRequest{
		Input: map[string]interface{}{},
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action execution failed: action failed")
}

func TestExecuteAction_InvalidJSON(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})

	ctx := context.Background()
	resp, err := c.ExecuteAction(ctx, "TEST_ACTION", &ExecuteActionRequest{
		Input: map[string]interface{}{},
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestSendEmail(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, ActionGmailSendEmail)

		var req ExecuteActionRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "entity-1", req.EntityID)
		assert.Equal(t, "gmail", req.AppName)
		assert.Equal(t, "test@example.com", req.Input["recipient_email"])
		assert.Equal(t, "Test Subject", req.Input["subject"])
		assert.Equal(t, "Test Body", req.Input["body"])
		assert.Equal(t, true, req.Input["is_html"])

		resp := ExecuteActionResponse{Successful: true, ExecutionID: "email-1"}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	resp, err := c.SendEmail(ctx, "entity-1", &SendEmailRequest{
		RecipientEmail: "test@example.com",
		Subject:        "Test Subject",
		Body:           "Test Body",
		CC:             []string{"cc@example.com"},
		BCC:            []string{"bcc@example.com"},
		IsHTML:         true,
	})

	require.NoError(t, err)
	assert.True(t, resp.Successful)
}

func TestSendEmail_MinimalFields(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req ExecuteActionRequest
		json.NewDecoder(r.Body).Decode(&req)

		// CC, BCC, and is_html should not be present
		_, hasCc := req.Input["cc"]
		_, hasBcc := req.Input["bcc"]
		_, hasIsHTML := req.Input["is_html"]
		assert.False(t, hasCc)
		assert.False(t, hasBcc)
		assert.False(t, hasIsHTML)

		resp := ExecuteActionResponse{Successful: true}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	_, err := c.SendEmail(ctx, "entity-1", &SendEmailRequest{
		RecipientEmail: "test@example.com",
		Subject:        "Hello",
		Body:           "World",
	})
	require.NoError(t, err)
}

func TestSendTelegramMessage(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, ActionTelegramSendMessage)

		var req ExecuteActionRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "entity-1", req.EntityID)
		assert.Equal(t, "telegram", req.AppName)
		assert.Equal(t, "12345", req.Input["chat_id"])
		assert.Equal(t, "Hello!", req.Input["text"])
		assert.Equal(t, "HTML", req.Input["parse_mode"])

		resp := ExecuteActionResponse{Successful: true}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	resp, err := c.SendTelegramMessage(ctx, "entity-1", &SendTelegramMessageRequest{
		ChatID:    "12345",
		Text:      "Hello!",
		ParseMode: "HTML",
	})

	require.NoError(t, err)
	assert.True(t, resp.Successful)
}

func TestSendTelegramMessage_NoParseModeField(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req ExecuteActionRequest
		json.NewDecoder(r.Body).Decode(&req)
		_, hasParseMode := req.Input["parse_mode"]
		assert.False(t, hasParseMode)

		resp := ExecuteActionResponse{Successful: true}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	_, err := c.SendTelegramMessage(ctx, "entity-1", &SendTelegramMessageRequest{
		ChatID: "12345",
		Text:   "Hello!",
	})
	require.NoError(t, err)
}

func TestSendSlackMessage(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, ActionSlackSendMessage)

		var req ExecuteActionRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "entity-1", req.EntityID)
		assert.Equal(t, "slack", req.AppName)
		assert.Equal(t, "#general", req.Input["channel"])
		assert.Equal(t, "Hello Slack!", req.Input["text"])

		resp := ExecuteActionResponse{Successful: true}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	resp, err := c.SendSlackMessage(ctx, "entity-1", &SendSlackMessageRequest{
		Channel: "#general",
		Text:    "Hello Slack!",
	})

	require.NoError(t, err)
	assert.True(t, resp.Successful)
}

func TestSendSlackDirectMessage(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, ActionSlackSendDirectMessage)

		var req ExecuteActionRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "entity-1", req.EntityID)
		assert.Equal(t, "slack", req.AppName)
		assert.Equal(t, "U12345", req.Input["user_id"])
		assert.Equal(t, "Hello DM!", req.Input["text"])

		resp := ExecuteActionResponse{Successful: true}
		json.NewEncoder(w).Encode(resp)
	})

	ctx := context.Background()
	resp, err := c.SendSlackDirectMessage(ctx, "entity-1", &SendSlackDirectMessageRequest{
		UserID: "U12345",
		Text:   "Hello DM!",
	})

	require.NoError(t, err)
	assert.True(t, resp.Successful)
}

func TestExecuteAction_CanceledContext(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		resp := ExecuteActionResponse{Successful: true}
		json.NewEncoder(w).Encode(resp)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	resp, err := c.ExecuteAction(ctx, "TEST", &ExecuteActionRequest{
		Input: map[string]interface{}{},
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
}
