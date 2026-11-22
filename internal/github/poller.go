package github

import (
	"context"
	"log"
	"time"

	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CheckOnce(ctx context.Context, pool *pgxpool.Pool) error {
	repositories, err := storage.ListRepositories(ctx, pool)
	if err != nil {
		return err
	}

	log.Printf("GitHub: checking %d repositories", len(repositories))

	for _, repository := range repositories {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		release, err := LatestRelease(ctx, repository)
		if err != nil {
			log.Printf("GitHub: %s: %v", repository, err)
			continue
		}

		isNew, err := storage.ProcessRelease(ctx, pool, release)
		if err != nil {
			log.Printf("GitHub: save %s: %v", repository, err)
			continue
		}

		if isNew {
			log.Printf("GitHub: new release %s %s", repository, release.Tag)
		}
	}

	return nil
}

func Run(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Minute
	}

	check := func() {
		if err := CheckOnce(ctx, pool); err != nil && ctx.Err() == nil {
			log.Printf("GitHub poller: %v", err)
		}
	}

	check()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		}
	}
}
