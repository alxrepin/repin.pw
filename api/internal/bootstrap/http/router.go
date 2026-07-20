package http

import (
	"net/http"

	channelget "repin/internal/context/presentation/http/channel/get"
	"repin/internal/context/presentation/http/feeds"
	postsget "repin/internal/context/presentation/http/posts/get"
	postslist "repin/internal/context/presentation/http/posts/list"
	"repin/internal/pkg/httpx"
)

func newRouter(logMW *httpx.Log, corsMW *httpx.CORS, postsC *postslist.Controller, postC *postsget.Controller, channelC *channelget.Controller, feedsC *feeds.Controller) *httpx.Router {
	r := httpx.NewRouter().WithErrorMapper(errMap).WithHealthCheck()

	r.Use(corsMW.Handle)
	r.Use(logMW.Handle)

	api := r.PathPrefix("/api/v1").Subrouter()

	httpx.HandleRoute(api, "/posts", postsC.Handle).Methods(http.MethodGet)
	httpx.HandleRoute(api, "/posts/{slug}", postC.Handle).Methods(http.MethodGet)
	httpx.HandleRoute(api, "/channel", channelC.Handle).Methods(http.MethodGet)

	r.HandleFunc("/sitemap.xml", feedsC.Sitemap()).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/rss.xml", feedsC.RSS()).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/llms.txt", feedsC.LLMs()).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/llms-full.txt", feedsC.LLMsFull()).Methods(http.MethodGet, http.MethodHead)

	return r
}
