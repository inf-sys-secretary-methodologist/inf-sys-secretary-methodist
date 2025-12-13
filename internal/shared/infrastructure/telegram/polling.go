// Package telegram provides Telegram Bot API integration.
package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// PollingService handles Telegram long polling for local development
type PollingService struct {
	botToken   string
	httpClient *http.Client
	logger     *slog.Logger
	handler    UpdateHandler
	running    bool
	mu         sync.Mutex
	stopCh     chan struct{}
	offset     int64
}

// UpdateHandler is a function that handles Telegram updates
type UpdateHandler func(update *Update)

// NewPollingService creates a new polling service
func NewPollingService(botToken string, logger *slog.Logger) *PollingService {
	return &PollingService{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Long polling timeout
		},
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// SetHandler sets the update handler
func (p *PollingService) SetHandler(handler UpdateHandler) {
	p.handler = handler
}

// Start starts the polling loop
func (p *PollingService) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.mu.Unlock()

	p.logger.Info("Starting Telegram polling service")

	// Delete any existing webhook first
	if err := p.deleteWebhook(); err != nil {
		p.logger.Warn("Failed to delete webhook", "error", err)
	}

	go p.pollLoop(ctx)
	return nil
}

// Stop stops the polling loop
func (p *PollingService) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	p.running = false
	close(p.stopCh)
	p.logger.Info("Stopping Telegram polling service")
}

// deleteWebhook removes any existing webhook
func (p *PollingService) deleteWebhook() error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", p.botToken)
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// pollLoop continuously polls for updates
func (p *PollingService) pollLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
			updates, err := p.getUpdates()
			if err != nil {
				p.logger.Error("Failed to get updates", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, update := range updates {
				if p.handler != nil {
					p.handler(&update)
				}
				// Update offset to acknowledge the update
				if update.UpdateID >= p.offset {
					p.offset = update.UpdateID + 1
				}
			}
		}
	}
}

// getUpdates fetches updates from Telegram
func (p *PollingService) getUpdates() ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", p.botToken, p.offset)

	resp, err := p.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !response.OK {
		return nil, fmt.Errorf("telegram API returned not ok")
	}

	if len(response.Result) > 0 {
		p.logger.Debug("Received updates", "count", len(response.Result))
	}

	return response.Result, nil
}
