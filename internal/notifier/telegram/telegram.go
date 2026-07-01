// Package telegram...
package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IKHINtech/composeguard/internal/config"
)

type sendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func Send(cfg config.TelegramConfig, message string) error {
	if !cfg.Enabled {
		return nil
	}

	botToken := resolveEnv(cfg.BotToken)
	chatID := resolveEnv(cfg.ChatID)

	if botToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}

	if chatID == "" {
		return fmt.Errorf("telegram chat id is empty")
	}
	payload := sendMessageRequest{
		ChatID: chatID,
		Text:   message,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func resolveEnv(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		key := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		return os.Getenv(key)
	}
	return value
}
