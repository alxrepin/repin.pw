package worker

import (
	"context"
	"errors"
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

	err := r.raw.WithSession(ctx, func(ctx context.Context) error {
		return r.runner.Run(ctx)
	})
	if errors.Is(err, context.Canceled) {
		return nil // graceful shutdown
	}

	return err
}

func (a *App) Stop() {
	a.registry.cleanup()
}
