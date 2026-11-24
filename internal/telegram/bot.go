package telegram

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(token string, pool *pgxpool.Pool) (*bot.Bot, error) {
	b, err := bot.New(token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"start",
		bot.MatchTypeCommandStartOnly,
		startHandler,
	)

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"help",
		bot.MatchTypeCommandStartOnly,
		helpHandler,
	)

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"add",
		bot.MatchTypeCommandStartOnly,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			addHandler(ctx, b, u, pool)
		},
	)

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"remove",
		bot.MatchTypeCommandStartOnly,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			removeHandler(ctx, b, u, pool)
		},
	)

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"list",
		bot.MatchTypeCommandStartOnly,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			listHandler(ctx, b, u, pool)
		},
	)

	return b, nil
}
