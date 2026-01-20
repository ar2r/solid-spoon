package handler

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartHandler struct{}

func NewStartHandler() *StartHandler {
	return &StartHandler{}
}

func (h *StartHandler) CanHandle(update tgbotapi.Update) bool {
	return update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start"
}

func (h *StartHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userName := getUserName(update.Message.From.FirstName, update.Message.From.UserName)
	greeting := formatGreeting(userName)

	log.Printf("[START] Greeting user: %s", userName)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, greeting)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("[START] Failed to send message: %v", err)
	}
}

func getUserName(firstName, userName string) string {
	if firstName != "" {
		return firstName
	}
	return userName
}

func formatGreeting(userName string) string {
	return "Привет, " + userName + "! Рад тебя видеть! 👋"
}
