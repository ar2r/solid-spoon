package handler

import (
	"testing"
)

func TestGetUserName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		userName  string
		expected  string
	}{
		{
			name:      "returns first name when available",
			firstName: "Артур",
			userName:  "artur123",
			expected:  "Артур",
		},
		{
			name:      "returns username when first name is empty",
			firstName: "",
			userName:  "artur123",
			expected:  "artur123",
		},
		{
			name:      "returns empty string when both are empty",
			firstName: "",
			userName:  "",
			expected:  "",
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

func TestFormatGreeting(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		expected string
	}{
		{
			name:     "formats greeting with name",
			userName: "Артур",
			expected: "Привет, Артур! Рад тебя видеть! 👋",
		},
		{
			name:     "formats greeting with empty name",
			userName: "",
			expected: "Привет, ! Рад тебя видеть! 👋",
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

func TestExtractYouTubeID(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "extracts ID from youtube.com/watch",
			text:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "extracts ID from youtu.be",
			text:     "https://youtu.be/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "extracts ID from youtube.com/shorts",
			text:     "https://youtube.com/shorts/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "returns empty for non-youtube URL",
			text:     "https://example.com/video",
			expected: "",
		},
		{
			name:     "returns empty for empty string",
			text:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractYouTubeID(tt.text)
			if result != tt.expected {
				t.Errorf("extractYouTubeID(%q) = %q, want %q",
					tt.text, result, tt.expected)
			}
		})
	}
}
