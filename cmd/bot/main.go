package main

import (
	"log"
	"os"

	"github.com/artur/solid-spoon/internal/bot"
	"github.com/artur/solid-spoon/internal/downloader"
	"github.com/artur/solid-spoon/internal/handler"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	b, err := bot.New(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Регистрируем обработчики
	b.RegisterHandler(handler.NewStartHandler())
	b.RegisterHandler(handler.NewYouTubeHandler(downloader.NewYouTubeDownloader()))

	// Отправляем уведомление о запуске
	b.SendStartupNotification()

	// Запускаем бота
	b.Run()
}
