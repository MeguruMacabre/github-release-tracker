package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/MeguruMacabre/github-release-tracker/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ProcessRelease(ctx context.Context, pool *pgxpool.Pool, release model.Release) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var releaseID int64

	err = tx.QueryRow(ctx, `
		INSERT INTO releases (repository, tag, url, published_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (repository, tag) DO NOTHING
		RETURNING id
	`, release.Repository, release.Tag, release.URL, release.PublishedAt).Scan(&releaseID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("save release: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO notifications (release_id, chat_id)
		SELECT $1, chat_id
		FROM subscriptions
		WHERE repository = $2
		ON CONFLICT (release_id, chat_id) DO NOTHING
	`, releaseID, release.Repository)
	if err != nil {
		return false, fmt.Errorf("create notifications: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return false, fmt.Errorf("commit transaction: %w", err)
	}

	return true, nil
}
