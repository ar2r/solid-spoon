package handler

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/artur/solid-spoon/internal/downloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type YouTubeHandler struct {
	downloader downloader.Downloader
}

func NewYouTubeHandler(dl downloader.Downloader) *YouTubeHandler {
	return &YouTubeHandler{
		downloader: dl,
	}
}

func (h *YouTubeHandler) CanHandle(update tgbotapi.Update) bool {
	if update.Message != nil {
		return extractYouTubeID(update.Message.Text) != ""
	}
	if update.CallbackQuery != nil {
		return strings.HasPrefix(update.CallbackQuery.Data, "yt:")
	}
	return false
}

func (h *YouTubeHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Обработка callback от кнопок
	if update.CallbackQuery != nil {
		h.handleCallback(bot, update)
		return
	}

	videoID := extractYouTubeID(update.Message.Text)
	chatID := update.Message.Chat.ID

	// Показываем действие "печатает"
	actionCfg := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	bot.Send(actionCfg)

	// Получаем доступные форматы
	formats, err := h.downloader.GetAvailableFormats(videoID)
	if err != nil {
		log.Printf("Failed to get formats: %v", err)
		errMsg := tgbotapi.NewMessage(chatID, "❌ Ошибка: "+err.Error())
		bot.Send(errMsg)
		return
	}

	// Создаём кнопки выбора качества
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, f := range formats {
		callbackData := fmt.Sprintf("yt:%s:%s", videoID, f.Quality)
		btn := tgbotapi.NewInlineKeyboardButtonData(f.Description, callbackData)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(chatID, "🎬 Выберите качество видео:")
	msg.ReplyMarkup = keyboard

	bot.Send(msg)
}

func (h *YouTubeHandler) handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback := update.CallbackQuery
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID

	// Парсим данные: yt:videoID:quality
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 3 {
		return
	}
	videoID := parts[1]
	quality := downloader.Quality(parts[2])

	// Отвечаем на callback
	callbackCfg := tgbotapi.NewCallback(callback.ID, "Скачиваю "+string(quality)+"...")
	bot.Send(callbackCfg)

	// Редактируем сообщение
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "⏳ Скачиваю видео в качестве "+string(quality)+"...")
	bot.Send(editMsg)

	// Показываем действие "отправляет видео"
	actionCfg := tgbotapi.NewChatAction(chatID, tgbotapi.ChatUploadVideo)
	bot.Send(actionCfg)

	// Скачиваем видео
	filePath, err := h.downloader.DownloadWithQuality(videoID, quality)
	if err != nil {
		log.Printf("Failed to download video: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Ошибка: "+err.Error())
		bot.Send(editMsg)
		return
	}
	defer os.Remove(filePath)

	// Обновляем действие перед отправкой
	bot.Send(actionCfg)

	// Отправляем видео
	videoFile := tgbotapi.FilePath(filePath)
	videoMsg := tgbotapi.NewVideo(chatID, videoFile)
	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("Failed to send video: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Не удалось отправить видео: "+err.Error())
		bot.Send(editMsg)
		return
	}

	// Удаляем сообщение с кнопками
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	bot.Send(deleteMsg)
}

func extractYouTubeID(text string) string {
	patterns := []string{
		`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)([a-zA-Z0-9_-]{11})`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
