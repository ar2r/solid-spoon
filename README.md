# solid-spoon

Telegram-бот на Go для скачивания YouTube видео.

## Возможности

- `/start` — приветствие пользователя по имени
- **YouTube Downloader** — отправьте ссылку на YouTube видео, и бот предложит выбрать качество и скачает его
  - Поддержка youtube.com/watch, youtu.be и YouTube Shorts
  - Выбор качества видео (360p, 480p, 720p, 1080p)
  - Автоматическое сжатие видео больше 50 МБ (требует ffmpeg)
  - Отправка видео как документа с сохранением качества

## Требования

- Go 1.21+
- Docker (для деплоя)
- **ffmpeg** (опционально, для сжатия больших видео)

## Локальный запуск

```bash
export TELEGRAM_BOT_TOKEN="токен_от_BotFather"
export ADMIN_CHAT_ID="ваш_chat_id"  # опционально
go run ./cmd/bot
```

## Docker

```bash
docker build -t solid-spoon .
docker run -e TELEGRAM_BOT_TOKEN="токен" solid-spoon
```

> **Примечание:** Для сжатия видео в Docker-контейнере должен быть установлен ffmpeg.

## CI/CD

При push в `main`:
1. Запускаются тесты
2. Собирается Docker-образ → `ghcr.io/ar2r/solid-spoon:main`
3. Вызывается Portainer webhook → автоматический редеплой контейнера

### GitHub Secrets

| Secret | Описание |
|--------|----------|
| `PORTAINER_WEBHOOK_URL` | Webhook URL из Portainer для автодеплоя |

### Environment Variables (Portainer)

| Переменная | Описание | Обязательная |
|------------|----------|--------------|
| `TELEGRAM_BOT_TOKEN` | Токен бота от @BotFather | ✅ Да |
| `ADMIN_CHAT_ID` | Chat ID для уведомлений о деплое | Нет |
| `APP_VERSION` | Версия приложения (устанавливается автоматически) | Нет |

## Структура проекта

```
├── cmd/bot/           # Точка входа приложения
├── internal/
│   ├── bot/           # Инициализация и запуск бота
│   ├── handler/       # Обработчики команд (start, youtube)
│   └── downloader/    # YouTube downloader
├── .github/workflows/ # CI/CD конфигурация
└── Dockerfile
```

## Разработка

```bash
# Запуск тестов
go test ./...

# Запуск с покрытием
go test -cover ./...

# Форматирование и проверка
go fmt ./... && go vet ./... && go test ./...
```
