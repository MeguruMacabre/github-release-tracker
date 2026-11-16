package storage

import (
	"context"
	"fmt"

	"github.com/MeguruMacabre/github-release-tracker/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetPendingNotifications(ctx context.Context, pool *pgxpool.Pool) ([]model.Notification, error) {
	rows, err := pool.Query(ctx, `
		SELECT n.id, n.chat_id, r.id, r.repository, r.tag, r.url, r.published_at
		FROM notifications n
		JOIN releases r ON r.id = n.release_id
		WHERE n.sent_at IS NULL
		ORDER BY n.id
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("get notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]model.Notification, 0)

	for rows.Next() {
		var notification model.Notification

		err = rows.Scan(
			&notification.ID,
			&notification.ChatID,
			&notification.Release.ID,
			&notification.Release.Repository,
			&notification.Release.Tag,
			&notification.Release.URL,
			&notification.Release.PublishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		notifications = append(notifications, notification)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read notifications: %w", err)
	}

	return notifications, nil
}

func MarkNotificationSent(ctx context.Context, pool *pgxpool.Pool, id int64) (bool, error) {
	result, err := pool.Exec(ctx, `
		UPDATE notifications
		SET sent_at = NOW()
		WHERE id = $1 AND sent_at IS NULL
	`, id)
	if err != nil {
		return false, fmt.Errorf("mark notification sent: %w", err)
	}

	return result.RowsAffected() == 1, nil
}
