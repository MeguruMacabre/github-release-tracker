package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/MeguruMacabre/github-release-tracker/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSubscriptionExists = errors.New("subscription already exists")

func AddSubscription(ctx context.Context, pool *pgxpool.Pool, chatID int64, repository string) (model.Subscription, error) {
	var subscription model.Subscription

	err := pool.QueryRow(ctx, `
		INSERT INTO subscriptions (chat_id, repository)
		VALUES ($1, $2)
		ON CONFLICT (chat_id, repository) DO NOTHING
		RETURNING id, chat_id, repository, created_at
	`, chatID, repository).Scan(
		&subscription.ID,
		&subscription.ChatID,
		&subscription.Repository,
		&subscription.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Subscription{}, ErrSubscriptionExists
	}
	if err != nil {
		return model.Subscription{}, fmt.Errorf("add subscription: %w", err)
	}

	return subscription, nil
}

func GetSubscriptions(ctx context.Context, pool *pgxpool.Pool, chatID int64) ([]model.Subscription, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, chat_id, repository, created_at
		FROM subscriptions
		WHERE chat_id = $1
		ORDER BY id
	`, chatID)
	if err != nil {
		return nil, fmt.Errorf("get subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)

	for rows.Next() {
		var subscription model.Subscription

		err = rows.Scan(
			&subscription.ID,
			&subscription.ChatID,
			&subscription.Repository,
			&subscription.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read subscriptions: %w", err)
	}

	return subscriptions, nil
}

func DeleteSubscription(ctx context.Context, pool *pgxpool.Pool, chatID int64, repository string) (bool, error) {
	result, err := pool.Exec(ctx, `
		DELETE FROM subscriptions
		WHERE chat_id = $1 AND repository = $2
	`, chatID, repository)
	if err != nil {
		return false, fmt.Errorf("delete subscription: %w", err)
	}

	return result.RowsAffected() > 0, nil
}
