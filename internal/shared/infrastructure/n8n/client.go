// Package n8n provides a client for triggering n8n workflow webhooks.
package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Client sends events to n8n webhooks to trigger workflows.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *logging.Logger
	enabled    bool
}

// Config holds n8n client configuration.
type Config struct {
	WebhookURL string
	Timeout    time.Duration
	Enabled    bool
}

// NewClient creates a new n8n webhook client.
func NewClient(cfg Config, logger *logging.Logger) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL: cfg.WebhookURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:  logger,
		enabled: cfg.Enabled,
	}
}

// TriggerWorkflow sends a POST request to an n8n webhook endpoint.
// The path is appended to the base webhook URL (e.g. "document-created").
func (c *Client) TriggerWorkflow(ctx context.Context, path string, payload map[string]any) error {
	if !c.enabled || c.baseURL == "" {
		return nil
	}

	url := c.baseURL + "/webhook/" + path

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("n8n: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("n8n: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("n8n: webhook request failed: %w", err)
	}
	defer resp.Body.Close()
	// Drain body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("n8n: webhook returned status %d for path %s", resp.StatusCode, path)
	}

	c.logger.Debug("n8n webhook triggered", map[string]any{
		"path":   path,
		"status": resp.StatusCode,
	})

	return nil
}

// TriggerAsync sends a webhook in a goroutine, logging errors instead of returning them.
func (c *Client) TriggerAsync(path string, payload map[string]any) {
	if !c.enabled || c.baseURL == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := c.TriggerWorkflow(ctx, path, payload); err != nil {
			c.logger.Warn("n8n async webhook failed", map[string]any{
				"path":  path,
				"error": err.Error(),
			})
		}
	}()
}

// IsEnabled returns whether the n8n integration is active.
func (c *Client) IsEnabled() bool {
	return c.enabled && c.baseURL != ""
}
