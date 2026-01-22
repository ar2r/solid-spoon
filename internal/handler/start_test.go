package handler

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestStartHandler_CanHandle(t *testing.T) {
	handler := NewStartHandler(nil, nil)

	tests := []struct {
		name     string
		update   tgbotapi.Update
		expected bool
	}{
		{
			name: "handles /start command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/start",
					Entities: []tgbotapi.MessageEntity{
						{Type: "bot_command", Offset: 0, Length: 6},
					},
				},
			},
			expected: true,
		},
		{
			name: "ignores regular message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "Hello",
				},
			},
			expected: false,
		},
		{
			name: "ignores other commands",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/help",
					Entities: []tgbotapi.MessageEntity{
						{Type: "bot_command", Offset: 0, Length: 5},
					},
				},
			},
			expected: false,
		},
		{
			name:     "ignores nil message",
			update:   tgbotapi.Update{},
			expected: false,
		},
		{
			name: "ignores callback query",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "some_data",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set Command field manually for test
			if tt.update.Message != nil {
				for _, entity := range tt.update.Message.Entities {
					if entity.Type == "bot_command" {
						tt.update.Message.Text = tt.update.Message.Text[entity.Offset : entity.Offset+entity.Length]
					}
				}
			}

			result := handler.CanHandle(tt.update)
			if result != tt.expected {
				t.Errorf("CanHandle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUserName_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		userName  string
		expected  string
	}{
		{
			name:      "prefers first name over username",
			firstName: "John",
			userName:  "john_doe",
			expected:  "John",
		},
		{
			name:      "uses username when first name empty",
			firstName: "",
			userName:  "john_doe",
			expected:  "john_doe",
		},
		{
			name:      "handles both empty",
			firstName: "",
			userName:  "",
			expected:  "",
		},
		{
			name:      "handles unicode in first name",
			firstName: "Артур",
			userName:  "artur",
			expected:  "Артур",
		},
		{
			name:      "handles special characters in username",
			firstName: "",
			userName:  "user_123",
			expected:  "user_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUserName(tt.firstName, tt.userName)
			if result != tt.expected {
				t.Errorf("getUserName(%q, %q) = %q, want %q",
					tt.firstName, tt.userName, result, tt.expected)
			}
		})
	}
}

func TestFormatGreeting_Various(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		expected string
	}{
		{
			name:     "formats with regular name",
			userName: "Alice",
			expected: "Привет, Alice! Рад тебя видеть! 👋",
		},
		{
			name:     "formats with unicode name",
			userName: "Мария",
			expected: "Привет, Мария! Рад тебя видеть! 👋",
		},
		{
			name:     "formats with empty name",
			userName: "",
			expected: "Привет, ! Рад тебя видеть! 👋",
		},
		{
			name:     "formats with long name",
			userName: "VeryLongUserNameThatIsUnusual",
			expected: "Привет, VeryLongUserNameThatIsUnusual! Рад тебя видеть! 👋",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGreeting(tt.userName)
			if result != tt.expected {
				t.Errorf("formatGreeting(%q) = %q, want %q",
					tt.userName, result, tt.expected)
			}
		})
	}
}
