// Package composio provides a client for the Composio API.
package composio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// BaseURL is the Composio API base URL.
	BaseURL = "https://backend.composio.dev/api"

	// ActionGmailSendEmail is the Gmail send email action ID.
	ActionGmailSendEmail = "GMAIL_SEND_EMAIL"
	// ActionGmailCreateDraft is the Gmail create draft action ID.
	ActionGmailCreateDraft = "GMAIL_CREATE_EMAIL_DRAFT"
	// ActionGmailReplyToThread is the Gmail reply to thread action ID.
	ActionGmailReplyToThread = "GMAIL_REPLY_TO_EMAIL_THREAD"
	// ActionGmailFetchEmails is the Gmail fetch emails action ID.
	ActionGmailFetchEmails = "GMAIL_FETCH_EMAILS"
)

// Client represents a Composio API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Composio client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
	}
}

// ExecuteActionRequest represents a request to execute an action
type ExecuteActionRequest struct {
	ConnectedAccountID string                 `json:"connectedAccountId,omitempty"`
	EntityID           string                 `json:"entityId,omitempty"`
	AppName            string                 `json:"appName,omitempty"`
	Input              map[string]interface{} `json:"input"`
}

// ExecuteActionResponse represents the response from executing an action
type ExecuteActionResponse struct {
	ExecutionID string                 `json:"executionId"`
	Data        map[string]interface{} `json:"data"`
	Successful  bool                   `json:"successful"`
	Error       *ErrorResponse         `json:"error,omitempty"`
}

// ErrorResponse represents an error from the API
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ExecuteAction executes a Composio action
func (c *Client) ExecuteAction(ctx context.Context, actionID string, req *ExecuteActionRequest) (*ExecuteActionResponse, error) {
	url := fmt.Sprintf("%s/v2/actions/%s/execute", c.baseURL, actionID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)

	// Log request details for debugging
	log.Printf("[Composio] Sending request to %s", url)
	log.Printf("[Composio] Request body: %s", string(reqBody))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[Composio] Failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response details for debugging
	log.Printf("[Composio] Response status: %d", resp.StatusCode)
	log.Printf("[Composio] Response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ExecuteActionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Successful && response.Error != nil {
		return nil, fmt.Errorf("action execution failed: %s", response.Error.Message)
	}

	return &response, nil
}

// SendEmailRequest represents a request to send an email via Gmail
type SendEmailRequest struct {
	RecipientEmail string   `json:"recipient_email"`
	Subject        string   `json:"subject"`
	Body           string   `json:"body"`
	CC             []string `json:"cc,omitempty"`
	BCC            []string `json:"bcc,omitempty"`
	IsHTML         bool     `json:"is_html,omitempty"`
}

// SendEmail sends an email via Gmail using Composio
func (c *Client) SendEmail(ctx context.Context, entityID string, email *SendEmailRequest) (*ExecuteActionResponse, error) {
	input := map[string]interface{}{
		"recipient_email": email.RecipientEmail,
		"subject":         email.Subject,
		"body":            email.Body,
	}

	if len(email.CC) > 0 {
		input["cc"] = email.CC
	}
	if len(email.BCC) > 0 {
		input["bcc"] = email.BCC
	}
	if email.IsHTML {
		input["is_html"] = true
	}

	req := &ExecuteActionRequest{
		EntityID: entityID,
		AppName:  "gmail",
		Input:    input,
	}

	return c.ExecuteAction(ctx, ActionGmailSendEmail, req)
}
