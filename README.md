# solid-spoon

Telegram-бот на Go, который приветствует пользователей по имени.

## Возможности

- Команда `/start` — бот приветствует пользователя по имени

## Локальный запуск

```bash
export TELEGRAM_BOT_TOKEN="токен_от_BotFather"
go run main.go
```

## Docker

```bash
docker build -t solid-spoon .
docker run -e TELEGRAM_BOT_TOKEN="токен" solid-spoon
```

## CI/CD

При push в `main`:
1. Запускаются тесты
2. Собирается Docker-образ → `ghcr.io/ar2r/solid-spoon:main`
3. Автоматический деплой на сервер через SSH

### GitHub Secrets

| Secret | Описание |
|--------|----------|
| `SERVER_HOST` | IP или домен сервера |
| `SERVER_USER` | SSH пользователь |
| `SERVER_SSH_KEY` | Приватный SSH ключ |
| `TELEGRAM_BOT_TOKEN` | Токен бота от @BotFather |
