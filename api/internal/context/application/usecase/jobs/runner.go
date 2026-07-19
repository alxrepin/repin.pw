package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"repin/internal/context/domain"
)

type Handler func(ctx context.Context, job domain.Job) error

type queue interface {
	Claim(ctx context.Context, kinds []domain.JobKind) (*domain.Job, error)
	Complete(ctx context.Context, id int64) error
	Retry(ctx context.Context, id int64, runAt time.Time, cause string) error
	Bury(ctx context.Context, id int64, cause string) error
	RequeueStale(ctx context.Context, lease time.Duration) (int64, error)
}

type RunnerConfig struct {
	Concurrency     int
	PollInterval    time.Duration
	Lease           time.Duration
	RetryBackoff    time.Duration
	RetryBackoffMax time.Duration
}

func DefaultRunnerConfig() RunnerConfig {
	return RunnerConfig{
		Concurrency:     2,
		PollInterval:    2 * time.Second,
		Lease:           30 * time.Minute,
		RetryBackoff:    30 * time.Second,
		RetryBackoffMax: 30 * time.Minute,
	}
}

type Runner struct {
	queue    queue
	cfg      RunnerConfig
	handlers map[domain.JobKind]Handler
}

func NewRunner(q queue, cfg RunnerConfig) *Runner {
	return &Runner{queue: q, cfg: cfg, handlers: make(map[domain.JobKind]Handler)}
}

func (r *Runner) Handle(kind domain.JobKind, handler Handler) {
	r.handlers[kind] = handler
}

func (r *Runner) Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	kinds := make([]domain.JobKind, 0, len(r.handlers))
	for kind := range r.handlers {
		kinds = append(kinds, kind)
	}

	if len(kinds) == 0 {
		return errors.New("job runner: no handlers registered")
	}

	log.Info().
		Int("concurrency", r.cfg.Concurrency).
		Interface("kinds", kinds).
		Msg("job runner started")

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error { return r.reap(gctx) })

	for i := range r.cfg.Concurrency {
		g.Go(func() error { return r.work(gctx, i, kinds) })
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (r *Runner) work(ctx context.Context, id int, kinds []domain.JobKind) error {
	log := zerolog.Ctx(ctx).With().Int("worker", id).Logger()
	ctx = log.WithContext(ctx)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		job, err := r.queue.Claim(ctx, kinds)

		switch {
		case errors.Is(err, domain.ErrNoJobs):
			if err := wait(ctx, r.cfg.PollInterval); err != nil {
				return err
			}

			continue

		case err != nil:
			if ctx.Err() != nil {
				return ctx.Err()
			}

			log.Error().Err(err).Msg("claim job failed")

			if err := wait(ctx, r.cfg.PollInterval); err != nil {
				return err
			}

			continue
		}

		r.process(ctx, *job)
	}
}

func (r *Runner) process(ctx context.Context, job domain.Job) {
	log := zerolog.Ctx(ctx).With().Int64("job", job.ID).Str("kind", string(job.Kind)).Logger()
	started := time.Now()

	book := context.WithoutCancel(ctx)

	handler, ok := r.handlers[job.Kind]
	if !ok {
		r.bury(book, log, job, errors.New("no handler registered"))
		return
	}

	err := handler(ctx, job)
	if err == nil {
		if err := r.queue.Complete(book, job.ID); err != nil {
			log.Error().Err(err).Msg("complete job failed")
			return
		}

		log.Info().Dur("elapsed", time.Since(started)).Int("attempt", job.Attempts).Msg("job done")

		return
	}

	if ctx.Err() != nil {
		if err := r.queue.Retry(book, job.ID, time.Now(), "interrupted by shutdown"); err != nil {
			log.Error().Err(err).Msg("requeue interrupted job failed")
		}

		return
	}

	if job.Attempts >= job.MaxAttempts {
		r.bury(book, log, job, err)
		return
	}

	runAt := time.Now().Add(r.backoff(job.Attempts))

	if retryErr := r.queue.Retry(book, job.ID, runAt, err.Error()); retryErr != nil {
		log.Error().Err(retryErr).Msg("schedule retry failed")
		return
	}

	log.Warn().
		Err(err).
		Int("attempt", job.Attempts).
		Int("max_attempts", job.MaxAttempts).
		Time("retry_at", runAt).
		Msg("job failed, retrying")
}

func (r *Runner) bury(ctx context.Context, log zerolog.Logger, job domain.Job, cause error) {
	if err := r.queue.Bury(ctx, job.ID, cause.Error()); err != nil {
		log.Error().Err(err).Msg("bury job failed")
		return
	}

	log.Error().Err(cause).Int("attempts", job.Attempts).Msg("job buried after exhausting attempts")
}

func (r *Runner) backoff(attempts int) time.Duration {
	delay := r.cfg.RetryBackoff
	for range max(0, attempts-1) {
		delay *= 2
		if delay >= r.cfg.RetryBackoffMax {
			return r.cfg.RetryBackoffMax
		}
	}

	return delay
}

func (r *Runner) reap(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	ticker := time.NewTicker(r.cfg.Lease / 2)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			n, err := r.queue.RequeueStale(ctx, r.cfg.Lease)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				log.Error().Err(err).Msg("requeue stale jobs failed")

				continue
			}

			if n > 0 {
				log.Warn().Int64("jobs", n).Msg("reclaimed jobs from expired leases")
			}
		}
	}
}

func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type Enqueuer interface {
	Enqueue(ctx context.Context, kind domain.JobKind, dedupKey string, payload any) error
}

func EnqueueMediaDownload(ctx context.Context, q Enqueuer, postID, messageID int64) error {
	payload := domain.MediaDownloadPayload{PostID: postID, MessageID: messageID}

	if err := q.Enqueue(ctx, domain.JobKindMediaDownload, payload.DedupKey(), payload); err != nil {
		return fmt.Errorf("enqueue media download of message %d: %w", messageID, err)
	}

	return nil
}

func EnqueueGenerateSEO(ctx context.Context, q Enqueuer, postID int64) error {
	payload := domain.GenerateSEOPayload{PostID: postID}

	if err := q.Enqueue(ctx, domain.JobKindGenerateSEO, payload.DedupKey(), payload); err != nil {
		return fmt.Errorf("enqueue seo generation of post %d: %w", postID, err)
	}

	return nil
}
