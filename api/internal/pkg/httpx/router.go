package httpx

import (
	"context"
	"maps"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Router struct {
	*mux.Router

	errorMapper map[error]ErrorCodes
}

type Route struct {
	*mux.Route

	errorMapper map[error]ErrorCodes
}

func NewRouter() *Router {
	errMap := map[error]ErrorCodes{}
	maps.Copy(errMap, defaultErrMap)

	return &Router{Router: mux.NewRouter(), errorMapper: errMap}
}

func (r *Router) WithErrorMapper(errMap map[error]ErrorCodes) *Router {
	maps.Copy(r.errorMapper, errMap)
	return r
}

func (r *Router) WithHealthCheck() *Router {
	r.Router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet, http.MethodHead)

	return r
}

func (r *Router) PathPrefix(prefix string) *Route {
	return &Route{Route: r.Router.PathPrefix(prefix), errorMapper: r.errorMapper}
}

func (rt *Route) Subrouter() *Router {
	return &Router{Router: rt.Route.Subrouter(), errorMapper: rt.errorMapper}
}

type handler[Request, Data, Item any] func(ctx context.Context, req Request) (APIResponse[Data, Item], error)

func HandleRoute[Request, Data, Item any](r *Router, path string, h handler[Request, Data, Item]) *mux.Route {
	return r.Router.HandleFunc(path, wrap(r, h))
}

func wrap[Request, Data, Item any](r *Router, h handler[Request, Data, Item]) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := withResponseWriter(req.Context(), w)

		defer func() {
			if rec := recover(); rec != nil {
				zerolog.Ctx(ctx).Error().
					Any("recover", rec).Bytes("stack", debug.Stack()).
					Msg("panic while handling request")
				writeError(ctx, ErrInternalServer, r.errorMapper)
			}
		}()

		request, err := decodeRequest[Request](req)
		if err != nil {
			writeError(ctx, err, r.errorMapper)
			return
		}

		response, err := h(ctx, request)
		if err != nil {
			writeError(ctx, err, r.errorMapper)
			return
		}

		writeJSON(ctx, http.StatusOK, response)
	}
}
