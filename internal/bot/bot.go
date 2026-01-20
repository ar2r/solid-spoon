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

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:      api,
		handlers: make([]Handler, 0),
	}, nil
}

func (b *Bot) RegisterHandler(h Handler) {
	b.handlers = append(b.handlers, h)
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		// Пропускаем только если нет ни сообщения, ни callback
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		for _, handler := range b.handlers {
			if handler.CanHandle(update) {
				go handler.Handle(b.api, update)
				break
			}
		}
	}
}
