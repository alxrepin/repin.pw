package http

import (
	"net/http"

	channelget "repin/internal/context/presentation/http/channel/get"
	mediaget "repin/internal/context/presentation/http/media/get"
	postsget "repin/internal/context/presentation/http/posts/get"
	postslist "repin/internal/context/presentation/http/posts/list"
	"repin/internal/pkg/httpx"
)

func newRouter(logMW *httpx.Log, corsMW *httpx.CORS, postsC *postslist.Controller, postC *postsget.Controller, channelC *channelget.Controller, mediaC *mediaget.Controller) *httpx.Router {
	r := httpx.NewRouter().WithErrorMapper(errMap).WithHealthCheck()

	r.Use(corsMW.Handle)
	r.Use(logMW.Handle)

	api := r.PathPrefix("/api/v1").Subrouter()

	httpx.HandleRoute(api, "/posts", postsC.Handle).Methods(http.MethodGet)
	httpx.HandleRoute(api, "/posts/{slug}", postC.Handle).Methods(http.MethodGet)
	httpx.HandleRoute(api, "/channel", channelC.Handle).Methods(http.MethodGet)

	// Raw handler: streams bytes, not JSON, so it bypasses the typed wrapper.
	api.Router.HandleFunc("/media/{key:.*}", mediaC.Handle).Methods(http.MethodGet, http.MethodHead)

	return r
}
