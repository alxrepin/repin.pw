package sync

import (
	"context"

	"github.com/rs/zerolog"

	"repin/internal/bootstrap"
	syncuc "repin/internal/context/application/usecase/sync"
	"repin/internal/context/infrastructure/db/postgres/channel"
	"repin/internal/context/infrastructure/db/postgres/job"
	"repin/internal/context/infrastructure/db/postgres/media"
	"repin/internal/context/infrastructure/db/postgres/post"
	"repin/internal/context/infrastructure/storage/minio"
	"repin/internal/context/infrastructure/telegram"
	"repin/internal/pkg/config"
	"repin/internal/pkg/db"
	"repin/internal/pkg/logger"
	"repin/internal/pkg/validator"
)

type registry struct {
	cfg     *bootstrap.Config
	log     *zerolog.Logger
	db      *db.Client
	storage *minio.Client
	usecase *syncuc.SyncChannel
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
	r.cfg = config.MustLoad(bootstrap.Config{})
	r.log = logger.MustLoad(r.cfg.LoggerConfig())
	r.db = db.MustLoad(ctx, r.cfg.PGConfig())

	storage, err := minio.NewClient(
		r.cfg.Storage.Endpoint,
		r.cfg.Storage.AccessKey,
		r.cfg.Storage.SecretKey,
		r.cfg.Storage.Bucket,
	)
	if err != nil {
		return err
	}

	r.storage = storage

	client := telegram.NewClient(r.cfg.Telegram.AppID, r.cfg.Telegram.AppHash, r.cfg.Telegram.Phone, r.cfg.ProxyURL())
	rawRepo := telegram.NewRawMessageRepository(client, r.storage, r.cfg.FaviconDir())

	r.usecase = syncuc.NewSyncChannel(
		rawRepo,
		post.NewRepository(r.db),
		channel.NewRepository(r.db),
		media.NewRepository(r.db),
		job.NewRepository(r.db),
		db.NewTxRunner(r.db),
	)

	return validator.ValidateStructDependencies(r)
}

func (r *registry) cleanup() {
	if r.db != nil {
		r.db.Close()
	}
}
