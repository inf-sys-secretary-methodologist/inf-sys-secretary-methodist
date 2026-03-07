package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

const (
	maxRetries    = 3
	retryDelay    = 2 * time.Second
	requestTimeout = 30 * time.Second
)

// APIResponse is the standard server response envelope.
type APIResponse struct {
	Status  string          `json:"status"`
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

// Client is an HTTP client for the server API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// doRequest executes an HTTP request with auth and retry logic.
func (c *Client) doRequest(ctx context.Context, method, path string, agent *agent.Agent, body any) (*APIResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		resp, err := c.executeRequest(ctx, method, path, agent, body)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.Status == "error" && resp.Code >= 500 {
			lastErr = fmt.Errorf("server error %d: %s", resp.Code, resp.Message)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) executeRequest(ctx context.Context, method, path string, a *agent.Agent, body any) (*APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if a != nil && a.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.AccessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response (status %d, body: %s): %w", resp.StatusCode, truncate(string(respBody), 200), err)
	}

	if apiResp.Status == "error" {
		return &apiResp, fmt.Errorf("API error %d: %s", apiResp.Code, apiResp.Message)
	}

	return &apiResp, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, a *agent.Agent) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodGet, path, a, nil)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, a *agent.Agent, body any) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, path, a, body)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, a *agent.Agent, body any) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPut, path, a, body)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, a *agent.Agent, body any) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPatch, path, a, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, a *agent.Agent) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, path, a, nil)
}

// ParseData unmarshals the Data field of an APIResponse into the target.
func ParseData(resp *APIResponse, target any) error {
	if resp.Data == nil {
		return fmt.Errorf("response data is nil")
	}
	return json.Unmarshal(resp.Data, target)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
