package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ListRepositories(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT repository
		FROM subscriptions
		ORDER BY repository
	`)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	defer rows.Close()

	repositories := make([]string, 0)

	for rows.Next() {
		var repository string

		if err := rows.Scan(&repository); err != nil {
			return nil, fmt.Errorf("scan repository: %w", err)
		}

		repositories = append(repositories, repository)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read repositories: %w", err)
	}

	return repositories, nil
}
