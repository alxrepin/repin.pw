package bot

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	"repin/internal/context/infrastructure/telegram"
)

type App struct {
	registry *registry
}

func New(ctx context.Context) *App {
	return &App{registry: newRegistry(ctx)}
}

func (a *App) Run(ctx context.Context) error {
	r := a.registry
	ctx = r.log.WithContext(ctx)
	username := r.cfg.Telegram.Channel

	err := r.bot.Run(ctx, telegram.BotHooks{
		OnPostsChanged: func(ctx context.Context, ids []int) error {
			return r.watch.SyncPosts(ctx, username, ids)
		},
		OnPostsDeleted: func(ctx context.Context, ids []int) error {
			return r.watch.DeletePosts(ctx, username, ids)
		},
		OnReady: a.refreshChannelLoop,
	})
	if errors.Is(err, context.Canceled) {
		return nil // graceful shutdown
	}

	return err
}

func (a *App) refreshChannelLoop(ctx context.Context) error {
	r := a.registry
	log := zerolog.Ctx(ctx)

	refresh := func() {
		if err := r.watch.RefreshChannel(ctx, r.cfg.Telegram.Channel); err != nil {
			log.Error().Err(err).Msg("refresh channel info failed")
		}
	}

	refresh()

	ticker := time.NewTicker(r.cfg.Telegram.ChannelRefresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			refresh()
		}
	}
}

func (a *App) Stop() {
	a.registry.cleanup()
}
