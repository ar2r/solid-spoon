package bot

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MockHandler implements Handler interface for testing
type MockHandler struct {
	canHandleFunc func(update tgbotapi.Update) bool
	handleFunc    func(bot *tgbotapi.BotAPI, update tgbotapi.Update)
}

func (m *MockHandler) CanHandle(update tgbotapi.Update) bool {
	if m.canHandleFunc != nil {
		return m.canHandleFunc(update)
	}
	return false
}

func (m *MockHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if m.handleFunc != nil {
		m.handleFunc(bot, update)
	}
}

func TestBot_RegisterHandler(t *testing.T) {
	// This test doesn't need a real token, but NewBotAPI will fail without it
	// So we'll test the registration logic separately
	bot := &Bot{
		handlers: make([]Handler, 0),
	}

	// Initially no handlers
	if len(bot.handlers) != 0 {
		t.Errorf("Expected 0 handlers initially, got %d", len(bot.handlers))
	}

	// Register first handler
	handler1 := &MockHandler{}
	bot.RegisterHandler(handler1)

	if len(bot.handlers) != 1 {
		t.Errorf("Expected 1 handler after first registration, got %d", len(bot.handlers))
	}

	// Register second handler
	handler2 := &MockHandler{}
	bot.RegisterHandler(handler2)

	if len(bot.handlers) != 2 {
		t.Errorf("Expected 2 handlers after second registration, got %d", len(bot.handlers))
	}

	// Verify order is preserved
	if bot.handlers[0] != handler1 {
		t.Error("First handler should be handler1")
	}
	if bot.handlers[1] != handler2 {
		t.Error("Second handler should be handler2")
	}
}

func TestBot_HandlerExecution(t *testing.T) {
	bot := &Bot{
		handlers: make([]Handler, 0),
	}

	// Create a mock handler that can handle specific updates
	handlerCalled := false
	handler := &MockHandler{
		canHandleFunc: func(update tgbotapi.Update) bool {
			return update.Message != nil && update.Message.Text == "test"
		},
		handleFunc: func(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
			handlerCalled = true
		},
	}

	bot.RegisterHandler(handler)

	// Create a test update
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "test",
		},
	}

	// Check if handler can handle the update
	canHandle := false
	for _, h := range bot.handlers {
		if h.CanHandle(update) {
			canHandle = true
			h.Handle(nil, update)
			break
		}
	}

	if !canHandle {
		t.Error("Handler should be able to handle the update")
	}

	if !handlerCalled {
		t.Error("Handler should have been called")
	}
}

func TestBot_MultipleHandlers(t *testing.T) {
	bot := &Bot{
		handlers: make([]Handler, 0),
	}

	handler1Called := false
	handler2Called := false

	// First handler handles "command1"
	handler1 := &MockHandler{
		canHandleFunc: func(update tgbotapi.Update) bool {
			return update.Message != nil && update.Message.Text == "command1"
		},
		handleFunc: func(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
			handler1Called = true
		},
	}

	// Second handler handles "command2"
	handler2 := &MockHandler{
		canHandleFunc: func(update tgbotapi.Update) bool {
			return update.Message != nil && update.Message.Text == "command2"
		},
		handleFunc: func(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
			handler2Called = true
		},
	}

	bot.RegisterHandler(handler1)
	bot.RegisterHandler(handler2)

	// Test update for handler1
	update1 := tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "command1"},
	}

	for _, h := range bot.handlers {
		if h.CanHandle(update1) {
			h.Handle(nil, update1)
			break
		}
	}

	if !handler1Called {
		t.Error("Handler1 should have been called")
	}
	if handler2Called {
		t.Error("Handler2 should not have been called")
	}

	// Reset
	handler1Called = false
	handler2Called = false

	// Test update for handler2
	update2 := tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "command2"},
	}

	for _, h := range bot.handlers {
		if h.CanHandle(update2) {
			h.Handle(nil, update2)
			break
		}
	}

	if handler1Called {
		t.Error("Handler1 should not have been called")
	}
	if !handler2Called {
		t.Error("Handler2 should have been called")
	}
}
