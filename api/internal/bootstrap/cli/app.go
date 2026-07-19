package cli

import "context"

type App struct {
	registry *registry
}

func New(ctx context.Context) *App {
	return &App{registry: newRegistry(ctx)}
}

func (a *App) Run(ctx context.Context) error {
	ctx = a.registry.log.WithContext(ctx)
	return a.registry.root.ExecuteContext(ctx)
}

func (a *App) Stop() {
	a.registry.cleanup()
}
