package http

import (
	"context"

	"repin/internal/pkg/httpx"
)

type App struct {
	registry *registry
	server   *httpx.Server
}

func New(ctx context.Context) *App {
	r := newRegistry(ctx)

	return &App{
		registry: r,
		server:   httpx.New(r.router, r.cfg.HTTP.Config()),
	}
}

func (a *App) Run() error {
	a.registry.log.Info().
		Str("addr", a.registry.cfg.HTTP.Host+":"+a.registry.cfg.HTTP.Port).
		Msg("http server started")

	return a.server.Run()
}

func (a *App) Stop(ctx context.Context) error {
	err := a.server.Stop(ctx)
	a.registry.cleanup()

	return err
}
