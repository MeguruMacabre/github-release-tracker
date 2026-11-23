# GitHub Release Tracker

Сервис на Go для отслеживания новых релизов GitHub-репозиториев
и отправки уведомлений через Telegram.

## Стек

- Go
- PostgreSQL
- GitHub REST API
- Telegram Bot API
- Docker Compose

## Запуск

На текущем этапе:

```bash
docker compose up -d db  

go run ./cmd/app