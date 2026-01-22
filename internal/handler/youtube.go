package handler

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/artur/solid-spoon/internal/database/models"
	"github.com/artur/solid-spoon/internal/database/repository"
	"github.com/artur/solid-spoon/internal/downloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type YouTubeHandler struct {
	downloader downloader.Downloader
	userRepo   *repository.UserRepository
	statsRepo  *repository.StatsRepository
	videoRepo  *repository.VideoRepository
}

func NewYouTubeHandler(
	dl downloader.Downloader,
	userRepo *repository.UserRepository,
	statsRepo *repository.StatsRepository,
	videoRepo *repository.VideoRepository,
) *YouTubeHandler {
	return &YouTubeHandler{
		downloader: dl,
		userRepo:   userRepo,
		statsRepo:  statsRepo,
		videoRepo:  videoRepo,
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
	messageID := update.Message.MessageID

	log.Printf("[YOUTUBE] Processing video ID: %s for chat: %d", videoID, chatID)

	// Сохраняем пользователя в БД
	user, err := h.userRepo.UpsertFromTelegram(update.Message.From)
	if err != nil {
		log.Printf("[YOUTUBE] Failed to upsert user: %v", err)
	} else {
		// Записываем статистику команды
		if err := h.statsRepo.RecordCommand(user.ID, "youtube"); err != nil {
			log.Printf("[YOUTUBE] Failed to record command: %v", err)
		}
	}

	// Показываем действие "печатает"
	actionCfg := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	bot.Send(actionCfg)

	// Получаем доступные форматы
	log.Printf("[YOUTUBE] Fetching available formats for: %s", videoID)
	formats, err := h.downloader.GetAvailableFormats(videoID)
	if err != nil {
		log.Printf("[YOUTUBE] Failed to get formats: %v", err)
		errMsg := tgbotapi.NewMessage(chatID, "❌ Ошибка: "+err.Error())
		bot.Send(errMsg)
		return
	}

	log.Printf("[YOUTUBE] Found %d formats for: %s", len(formats), videoID)

	// Создаём кнопки выбора качества
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, f := range formats {
		callbackData := fmt.Sprintf("yt:%s:%s", videoID, f.Quality)
		btn := tgbotapi.NewInlineKeyboardButtonData(f.Description, callbackData)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
		log.Printf("[YOUTUBE] Added quality option: %s", f.Description)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(chatID, "🎬 Выберите качество видео:")
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("[YOUTUBE] Failed to send quality selection: %v", err)
	}

	// Удаляем сообщение пользователя с ссылкой
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	if _, err := bot.Send(deleteMsg); err != nil {
		log.Printf("[YOUTUBE] Failed to delete user message: %v", err)
	}
}

func (h *YouTubeHandler) handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback := update.CallbackQuery
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID

	// Парсим данные: yt:videoID:quality
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 3 {
		log.Printf("[YOUTUBE] Invalid callback data: %s", callback.Data)
		return
	}
	videoID := parts[1]
	quality := downloader.Quality(parts[2])

	log.Printf("[YOUTUBE] Callback: downloading %s in %s quality", videoID, quality)

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
	log.Printf("[YOUTUBE] Starting download: %s (%s)", videoID, quality)
	videoInfo, err := h.downloader.DownloadWithQualityInfo(videoID, quality)
	if err != nil {
		log.Printf("[YOUTUBE] Download failed: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Ошибка: "+err.Error())
		bot.Send(editMsg)
		return
	}
	defer func() {
		if err := os.Remove(videoInfo.FilePath); err != nil {
			log.Printf("[YOUTUBE] Failed to remove temp file %s: %v", videoInfo.FilePath, err)
		} else {
			log.Printf("[YOUTUBE] Temp file removed: %s", videoInfo.FilePath)
		}
	}()

	log.Printf("[YOUTUBE] Download complete: %s, sending to chat", videoInfo.FilePath)
	log.Printf("[YOUTUBE] Video metadata - Title: %s, Size: %dx%d, Duration: %ds, Compressed: %v",
		videoInfo.Title, videoInfo.Width, videoInfo.Height, videoInfo.Duration, videoInfo.Compressed)

	// Проверяем размер скачанного файла
	fileInfo, err := os.Stat(videoInfo.FilePath)
	if err != nil {
		log.Printf("[YOUTUBE] Failed to get file info: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Ошибка при проверке файла")
		bot.Send(editMsg)
		return
	}

	const maxTelegramBotAPI = 50 * 1024 * 1024 // 50 МБ - реальный лимит Bot API
	const maxTelegramDocSize = 2000 * 1024 * 1024 // 2 ГБ - теоретический лимит (работает только с локальным API)

	sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
	log.Printf("[YOUTUBE] File size: %.2f MB", sizeMB)

	// Если файл больше 2 ГБ - отказываем
	if fileInfo.Size() > maxTelegramDocSize {
		log.Printf("[YOUTUBE] File too large: %.2f MB (max 2000 MB)", sizeMB)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID,
			fmt.Sprintf("❌ Видео слишком большое (%.1f ГБ). Максимум 2 ГБ.\n\nВыберите качество пониже.", sizeMB/1024))
		bot.Send(editMsg)
		return
	}

	// Если файл уже сжат в downloader, но всё равно больше 50 МБ - что-то пошло не так
	if videoInfo.Compressed && fileInfo.Size() > maxTelegramBotAPI {
		log.Printf("[YOUTUBE] Compressed file still too large: %.2f MB", sizeMB)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID,
			fmt.Sprintf("❌ Не удалось сжать видео до 50 МБ (получилось %.1f МБ).\n\nПопробуйте более низкое качество.", sizeMB))
		bot.Send(editMsg)
		return
	}

	// Обновляем действие перед отправкой
	uploadAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatUploadDocument)
	bot.Send(uploadAction)

	// Отправляем видео как документ (файл)
	videoFile := tgbotapi.FilePath(videoInfo.FilePath)
	docMsg := tgbotapi.NewDocument(chatID, videoFile)

	// Формируем caption с заголовком и описанием
	caption := videoInfo.Title
	if videoInfo.Description != "" {
		// Ограничиваем описание до 200 символов
		desc := videoInfo.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		caption += "\n\n" + desc
	}

	// Добавляем информацию о сжатии
	if videoInfo.Compressed {
		caption += "\n\n⚙️ Видео сжато для отправки через Telegram"
	}

	// Telegram caption limit is 1024 characters
	if len(caption) > 1024 {
		caption = caption[:1021] + "..."
	}
	docMsg.Caption = caption

	if _, err := bot.Send(docMsg); err != nil {
		log.Printf("[YOUTUBE] Failed to send document: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Не удалось отправить видео: "+err.Error())
		bot.Send(editMsg)
		return
	}

	log.Printf("[YOUTUBE] Video sent successfully: %s", videoID)

	// Записываем скачивание в БД
	if user, err := h.userRepo.GetByTelegramID(callback.From.ID); err == nil && user != nil {
		download := &models.VideoDownload{
			UserID:        user.ID,
			VideoID:       videoID,
			VideoURL:      fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
			VideoTitle:    videoInfo.Title,
			Quality:       string(quality),
			Compressed:    videoInfo.Compressed,
			FileSizeBytes: fileInfo.Size(),
			ExecutedAt:    time.Now(),
		}
		if err := h.videoRepo.RecordDownload(download); err != nil {
			log.Printf("[YOUTUBE] Failed to record download: %v", err)
		}
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
