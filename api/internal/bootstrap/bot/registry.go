package bot

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"repin/internal/bootstrap"
	syncuc "repin/internal/context/application/usecase/sync"
	watchuc "repin/internal/context/application/usecase/watch"
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
	bot     *telegram.Bot
	watch   *watchuc.WatchChannel
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

	if r.cfg.TelegramBotToken() == "" {
		return errors.New("TELEGRAM_BOT_TOKEN is required")
	}

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

	client := telegram.NewBotClient(
		r.cfg.Telegram.AppID,
		r.cfg.Telegram.AppHash,
		r.cfg.TelegramBotToken(),
		r.cfg.ProxyURL(),
	)
	rawRepo := telegram.NewRawMessageRepository(client, r.storage, r.cfg.FaviconDir())

	postRepo := post.NewRepository(r.db)
	channelRepo := channel.NewRepository(r.db)
	mediaRepo := media.NewRepository(r.db)
	tx := db.NewTxRunner(r.db)

	importer := syncuc.NewSyncChannel(rawRepo, postRepo, channelRepo, mediaRepo, job.NewRepository(r.db), tx)

	r.watch = watchuc.NewWatchChannel(rawRepo, importer, postRepo, mediaRepo, channelRepo, tx)
	r.bot = telegram.NewBot(client, r.cfg.Telegram.Channel)

	return validator.ValidateStructDependencies(r)
}

func (r *registry) cleanup() {
	if r.db != nil {
		r.db.Close()
	}
}
