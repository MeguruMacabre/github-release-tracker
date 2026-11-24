package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MeguruMacabre/github-release-tracker/internal/model"
	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5/pgxpool"
)

func formatNotification(n model.Notification) string {
	return fmt.Sprintf(
		"Новый релиз GitHub!\n\n"+
			"Репозиторий: %s\n"+
			"Версия: %s\n"+
			"Ссылка: %s",
		n.Release.Repository,
		n.Release.Tag,
		n.Release.URL,
	)
}

func SendPending(
	ctx context.Context,
	b *bot.Bot,
	pool *pgxpool.Pool,
) error {
	notifications, err := storage.GetPendingNotifications(ctx, pool)
	if err != nil {
		return fmt.Errorf("get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := ctx.Err(); err != nil {
			return err
		}

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: notification.ChatID,
			Text:   formatNotification(notification),
		})
		if err != nil {
			log.Printf(
				"Telegram: send notification %d: %v",
				notification.ID,
				err,
			)
			continue
		}

		_, err = storage.MarkNotificationSent(
			ctx, pool, notification.ID,
		)
		if err != nil {
			log.Printf(
				"Telegram: mark notification %d: %v",
				notification.ID,
				err,
			)
			continue
		}

		log.Printf(
			"Telegram: notification %d sent",
			notification.ID,
		)
	}

	return nil
}

func RunNotifications(
	ctx context.Context,
	b *bot.Bot,
	pool *pgxpool.Pool,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := SendPending(ctx, b, pool); err != nil && ctx.Err() == nil {
			log.Printf("Telegram: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
