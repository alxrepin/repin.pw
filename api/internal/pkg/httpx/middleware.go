package httpx

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Log struct {
	log *zerolog.Logger
}

func NewLog(log *zerolog.Logger) *Log { return &Log{log: log} }

func (m *Log) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := m.log.WithContext(r.Context())

		zerolog.Ctx(ctx).Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("request")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type CORS struct{}

func NewCORS() *CORS { return &CORS{} }

func (m *CORS) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
