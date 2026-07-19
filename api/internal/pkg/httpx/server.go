package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Host string
	Port string
}

type Server struct {
	instance *http.Server
}

func New(r *Router, cfg Config) *Server {
	return &Server{
		instance: &http.Server{
			Addr:         net.JoinHostPort(cfg.Host, cfg.Port),
			Handler:      r.Router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	if err := s.instance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.instance.Shutdown(ctx)
}
