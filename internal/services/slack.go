package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// SlackServiceImpl implements SlackService via incoming webhook
type SlackServiceImpl struct {
	webhookURL string
	client    *http.Client
}

// NewSlackService creates a SlackService. If webhookURL is empty, Notify is a no-op.
func NewSlackService(webhookURL string) SlackService {
	return &SlackServiceImpl{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackServiceImpl) Notify(ctx context.Context, message string) error {
	if s.webhookURL == "" {
		slog.Debug("slack notify skipped (no webhook)")
		return nil
	}
	slog.Debug("slack notify", "message", message)
	body, err := json.Marshal(map[string]string{"text": message})
	if err != nil {
		return fmt.Errorf("slack marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}
