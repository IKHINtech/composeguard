package telegram_test

import (
	"testing"

	"github.com/IKHINtech/composeguard/internal/config"
	"github.com/IKHINtech/composeguard/internal/notifier/telegram"
)

func TestSend(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg     config.TelegramConfig
		message string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := telegram.Send(tt.cfg, tt.message)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Send() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Send() succeeded unexpectedly")
			}
		})
	}
}
