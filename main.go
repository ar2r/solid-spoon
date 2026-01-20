package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getUserName(firstName, userName string) string {
	if firstName != "" {
		return firstName
	}
	return userName
}

func formatGreeting(userName string) string {
	return "Привет, " + userName + "! Рад тебя видеть! 👋"
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() && update.Message.Command() == "start" {
			userName := getUserName(update.Message.From.FirstName, update.Message.From.UserName)
			greeting := formatGreeting(userName)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, greeting)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	}
}
