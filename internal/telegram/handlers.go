package telegram

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var repositoryPattern = regexp.MustCompile(
	`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`,
)

func sendText(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		log.Printf("Telegram send message: %v", err)
	}
}

func parseRepository(text string) (string, error) {
	parts := strings.Fields(text)

	if len(parts) != 2 {
		return "", errors.New("expected repository argument")
	}

	repository := strings.ToLower(parts[1])

	if !repositoryPattern.MatchString(repository) {
		return "", errors.New("invalid repository")
	}

	return repository, nil
}

func startHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil {
		return
	}

	sendText(ctx, b, u.Message.Chat.ID,
		"Привет! Я GitHub Release Tracker.\n\n"+
			"Я отслеживаю новые релизы GitHub-репозиториев.\n\n"+
			"Используй /help для просмотра команд.",
	)
}

func helpHandler(ctx context.Context, b *bot.Bot, u *models.Update) {
	if u.Message == nil {
		return
	}

	text := "Доступные команды:\n\n" +
		"/add owner/repo — добавить подписку\n" +
		"/remove owner/repo — удалить подписку\n" +
		"/list — мои подписки\n" +
		"/help — помощь"

	sendText(ctx, b, u.Message.Chat.ID, text)
}

func addHandler(
	ctx context.Context,
	b *bot.Bot,
	u *models.Update,
	pool *pgxpool.Pool,
) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID

	repository, err := parseRepository(u.Message.Text)
	if err != nil {
		sendText(ctx, b, chatID, "Использование: /add owner/repo")
		return
	}

	_, err = storage.AddSubscription(ctx, pool, chatID, repository)

	if errors.Is(err, storage.ErrSubscriptionExists) {
		sendText(ctx, b, chatID, "Ты уже подписан на "+repository)
		return
	}

	if err != nil {
		log.Printf("Add subscription: %v", err)
		sendText(ctx, b, chatID, "Не удалось добавить подписку.")
		return
	}

	sendText(ctx, b, chatID, "Подписка добавлена: "+repository)
}

func removeHandler(
	ctx context.Context,
	b *bot.Bot,
	u *models.Update,
	pool *pgxpool.Pool,
) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID

	repository, err := parseRepository(u.Message.Text)
	if err != nil {
		sendText(ctx, b, chatID, "Использование: /remove owner/repo")
		return
	}

	deleted, err := storage.DeleteSubscription(
		ctx, pool, chatID, repository,
	)

	if err != nil {
		log.Printf("Remove subscription: %v", err)
		sendText(ctx, b, chatID, "Не удалось удалить подписку.")
		return
	}

	if !deleted {
		sendText(ctx, b, chatID, "Подписка не найдена.")
		return
	}

	sendText(ctx, b, chatID, "Подписка удалена: "+repository)
}

func listHandler(
	ctx context.Context,
	b *bot.Bot,
	u *models.Update,
	pool *pgxpool.Pool,
) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID

	subscriptions, err := storage.GetSubscriptions(ctx, pool, chatID)
	if err != nil {
		log.Printf("List subscriptions: %v", err)
		sendText(ctx, b, chatID, "Не удалось получить подписки.")
		return
	}

	if len(subscriptions) == 0 {
		sendText(ctx, b, chatID, "У тебя пока нет подписок.")
		return
	}

	// Отправляем не больше 30 подписок в одном сообщении.
	for i := 0; i < len(subscriptions); i += 30 {
		end := i + 30
		if end > len(subscriptions) {
			end = len(subscriptions)
		}

		var text strings.Builder
		text.WriteString("Твои подписки:\n\n")

		for _, sub := range subscriptions[i:end] {
			text.WriteString("• ")
			text.WriteString(sub.Repository)
			text.WriteString("\n")
		}

		sendText(ctx, b, chatID, text.String())
	}
}
