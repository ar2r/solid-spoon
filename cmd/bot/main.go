package main

import (
	"log"
	"os"

	"github.com/artur/solid-spoon/internal/bot"
	"github.com/artur/solid-spoon/internal/database"
	"github.com/artur/solid-spoon/internal/database/repository"
	"github.com/artur/solid-spoon/internal/downloader"
	"github.com/artur/solid-spoon/internal/handler"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	// Инициализация базы данных
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/bot.db"
	}

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Запускаем миграции
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Создаём репозитории
	userRepo := repository.NewUserRepository(db.DB)
	statsRepo := repository.NewStatsRepository(db.DB)
	videoRepo := repository.NewVideoRepository(db.DB)

	b, err := bot.New(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Регистрируем обработчики с репозиториями
	b.RegisterHandler(handler.NewStartHandler(userRepo, statsRepo))
	b.RegisterHandler(handler.NewYouTubeHandler(downloader.NewYouTubeDownloader(), userRepo, statsRepo, videoRepo))

	// Отправляем уведомление о запуске
	b.SendStartupNotification()

	// Запускаем бота
	b.Run()
}
