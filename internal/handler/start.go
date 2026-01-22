package handler

import (
	"log"

	"github.com/artur/solid-spoon/internal/database/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartHandler struct {
	userRepo  *repository.UserRepository
	statsRepo *repository.StatsRepository
}

func NewStartHandler(userRepo *repository.UserRepository, statsRepo *repository.StatsRepository) *StartHandler {
	return &StartHandler{
		userRepo:  userRepo,
		statsRepo: statsRepo,
	}
}

func (h *StartHandler) CanHandle(update tgbotapi.Update) bool {
	return update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start"
}

func (h *StartHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userName := getUserName(update.Message.From.FirstName, update.Message.From.UserName)
	greeting := formatGreeting(userName)

	log.Printf("[START] Greeting user: %s", userName)

	// Сохраняем пользователя в БД
	user, err := h.userRepo.UpsertFromTelegram(update.Message.From)
	if err != nil {
		log.Printf("[START] Failed to upsert user: %v", err)
	} else {
		// Записываем статистику команды
		if err := h.statsRepo.RecordCommand(user.ID, "start"); err != nil {
			log.Printf("[START] Failed to record command: %v", err)
		}
	}

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
