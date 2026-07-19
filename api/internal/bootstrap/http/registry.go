package http

import (
	"context"

	"github.com/rs/zerolog"

	"repin/internal/bootstrap"
	"repin/internal/context/application/service"
	"repin/internal/context/infrastructure/db/postgres/channel"
	"repin/internal/context/infrastructure/db/postgres/media"
	"repin/internal/context/infrastructure/db/postgres/post"
	channelget "repin/internal/context/presentation/http/channel/get"
	postsget "repin/internal/context/presentation/http/posts/get"
	postslist "repin/internal/context/presentation/http/posts/list"
	"repin/internal/pkg/config"
	"repin/internal/pkg/db"
	"repin/internal/pkg/httpx"
	"repin/internal/pkg/logger"
	"repin/internal/pkg/validator"
)

type registry struct {
	cfg *bootstrap.APIConfig
	log *zerolog.Logger
	db  *db.Client

	repos struct {
		post    *post.Repository
		channel *channel.Repository
		media   *media.Repository
	}
	services struct {
		post    *service.PostService
		channel *service.ChannelService
	}
	controllers struct {
		posts   *postslist.Controller
		post    *postsget.Controller
		channel *channelget.Controller
	}
	middleware struct {
		log  *httpx.Log
		cors *httpx.CORS
	}
	router *httpx.Router
}

func newRegistry(ctx context.Context) *registry {
	r := new(registry)
	if err := r.load(ctx); err != nil {
		r.cleanup()
		panic(err)
	}

	return r
}

func (r *registry) load(ctx context.Context) error {
	r.cfg = config.MustLoad(bootstrap.APIConfig{})
	r.log = logger.MustLoad(r.cfg.Logger.Config())
	r.db = db.MustLoad(ctx, r.cfg.Database.Config())

	r.repos.post = post.NewRepository(r.db)
	r.repos.channel = channel.NewRepository(r.db)
	r.repos.media = media.NewRepository(r.db)

	r.services.post = service.NewPostService(r.repos.post, r.repos.media)
	r.services.channel = service.NewChannelService(r.repos.channel)

	r.controllers.posts = postslist.NewController(r.services.post)
	r.controllers.post = postsget.NewController(r.services.post)
	r.controllers.channel = channelget.NewController(r.services.channel)

	r.middleware.log = httpx.NewLog(r.log)
	r.middleware.cors = httpx.NewCORS()

	r.router = newRouter(r.middleware.log, r.middleware.cors, r.controllers.posts, r.controllers.post, r.controllers.channel)

	return validator.ValidateStructDependencies(r)
}

func (r *registry) cleanup() {
	if r.db != nil {
		r.db.Close()
	}
}
