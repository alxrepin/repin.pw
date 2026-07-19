package sync

import "context"

type App struct {
	registry *registry
}

func New(ctx context.Context) *App {
	return &App{registry: newRegistry(ctx)}
}

func (a *App) Run(ctx context.Context) error {
	r := a.registry
	ctx = r.log.WithContext(ctx)

	return r.usecase.Execute(ctx, r.cfg.Telegram.Channel)
}

func (a *App) Stop() {
	a.registry.cleanup()
}
