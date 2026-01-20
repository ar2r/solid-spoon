package main

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
