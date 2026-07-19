package worker

import (
	"context"

	"github.com/rs/zerolog"

	"repin/internal/bootstrap"
	"repin/internal/context/application/usecase/jobs"
	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/db/postgres/job"
	"repin/internal/context/infrastructure/db/postgres/media"
	"repin/internal/context/infrastructure/db/postgres/post"
	"repin/internal/context/infrastructure/openrouter"
	"repin/internal/context/infrastructure/storage/minio"
	"repin/internal/context/infrastructure/telegram"
	"repin/internal/pkg/config"
	"repin/internal/pkg/db"
	"repin/internal/pkg/logger"
	"repin/internal/pkg/proxyx"
	"repin/internal/pkg/validator"
)

type registry struct {
	cfg     *bootstrap.Config
	log     *zerolog.Logger
	db      *db.Client
	storage *minio.Client
	raw     *telegram.RawMessageRepository
	runner  *jobs.Runner
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

	client := telegram.NewWorkerClient(
		r.cfg.Telegram.AppID,
		r.cfg.Telegram.AppHash,
		r.cfg.Telegram.Phone,
		r.cfg.TelegramBotToken(),
		r.cfg.ProxyURL(),
	)
	r.raw = telegram.NewRawMessageRepository(client, r.storage, r.cfg.FaviconDir())

	cfg := jobs.DefaultRunnerConfig()
	cfg.Concurrency = r.cfg.Worker.Concurrency
	cfg.PollInterval = r.cfg.Worker.PollInterval
	cfg.Lease = r.cfg.Worker.JobLease

	r.runner = jobs.NewRunner(job.NewRepository(r.db), cfg)

	r.runner.Handle(domain.JobKindMediaDownload, jobs.NewDownloadMedia(
		r.raw,
		media.NewRepository(r.db),
		r.storage,
		r.cfg.Telegram.Channel,
	).Handle)

	if err := r.registerSEO(); err != nil {
		return err
	}

	return validator.ValidateStructDependencies(r)
}

func (r *registry) registerSEO() error {
	if r.cfg.OpenRouterKey() == "" {
		r.log.Warn().Msg("OPENROUTER_API_KEY is not set: seo jobs will not be processed")

		return nil
	}

	httpClient, err := proxyx.HTTPClient(r.cfg.ProxyURL(), r.cfg.OpenRouter.Timeout)
	if err != nil {
		return err
	}

	client := openrouter.NewClient(openrouter.Config{
		APIKey:        r.cfg.OpenRouterKey(),
		Model:         r.cfg.OpenRouter.Model,
		FallbackModel: r.cfg.OpenRouter.FallbackModel,
		MaxRetries:    r.cfg.OpenRouter.MaxRetries,
		Referer:       r.cfg.OpenRouterReferer(),
	}, httpClient)

	r.runner.Handle(domain.JobKindGenerateSEO, jobs.NewGenerateSEO(
		post.NewRepository(r.db),
		openrouter.NewSEOGenerator(client),
	).Handle)

	r.log.Info().
		Str("model", r.cfg.OpenRouter.Model).
		Str("fallback_model", r.cfg.OpenRouter.FallbackModel).
		Msg("seo generation enabled")

	return nil
}

func (r *registry) cleanup() {
	if r.db != nil {
		r.db.Close()
	}
}
