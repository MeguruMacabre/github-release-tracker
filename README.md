# GitHub Release Tracker

Сервис на Go для отслеживания релизов GitHub-репозиториев
с уведомлениями через Telegram.

## Стек

Go, PostgreSQL, GitHub REST API, Telegram Bot API, Docker Compose.

## Запуск

Создать `.env` по примеру `.env.example`, указать токены GitHub и Telegram.

Запустить:

`docker compose up --build -d`

Завершить:

`docker compose down`

Проверка: http://localhost:8080/healthz

## Команды бота

- /start — приветствие
- /help — справка
- /add owner/repo — подписка
- /remove owner/repo — удаление подписки
- /list — список подписок