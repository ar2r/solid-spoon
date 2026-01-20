package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler interface {
	CanHandle(update tgbotapi.Update) bool
	Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update)
}

type Bot struct {
	api      *tgbotapi.BotAPI
	handlers []Handler
}

func New(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("[BOT] Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:      api,
		handlers: make([]Handler, 0),
	}, nil
}

func (b *Bot) RegisterHandler(h Handler) {
	b.handlers = append(b.handlers, h)
	log.Printf("[BOT] Registered handler: %T", h)
}

func (b *Bot) Run() {
	log.Printf("[BOT] Starting bot with %d handlers", len(b.handlers))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		// Логируем входящее обновление
		if update.Message != nil {
			log.Printf("[BOT] Message from %s (@%s): %s",
				update.Message.From.FirstName,
				update.Message.From.UserName,
				update.Message.Text)
		}
		if update.CallbackQuery != nil {
			log.Printf("[BOT] Callback from %s (@%s): %s",
				update.CallbackQuery.From.FirstName,
				update.CallbackQuery.From.UserName,
				update.CallbackQuery.Data)
		}

		// Пропускаем только если нет ни сообщения, ни callback
		if update.Message == nil && update.CallbackQuery == nil {
			log.Printf("[BOT] Skipping update: no message or callback")
			continue
		}

		handled := false
		for _, handler := range b.handlers {
			if handler.CanHandle(update) {
				log.Printf("[BOT] Handling with: %T", handler)
				go handler.Handle(b.api, update)
				handled = true
				break
			}
		}

		if !handled {
			log.Printf("[BOT] No handler found for update")
		}
	}
}
