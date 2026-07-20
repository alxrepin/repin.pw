package feeds

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type snapshot struct {
	sitemap     []byte
	rss         []byte
	llms        []byte
	llmsFull    []byte
	generatedAt time.Time
}

const refreshTimeout = time.Minute

func (c *Controller) Refresh(ctx context.Context) error {
	snap := &snapshot{generatedAt: time.Now().UTC()}

	var err error

	if snap.sitemap, err = c.renderSitemap(ctx); err != nil {
		return fmt.Errorf("refresh sitemap: %w", err)
	}

	if snap.rss, err = c.buildRSS(ctx); err != nil {
		return fmt.Errorf("refresh rss: %w", err)
	}

	if snap.llms, err = c.buildLLMs(ctx); err != nil {
		return fmt.Errorf("refresh llms: %w", err)
	}

	if snap.llmsFull, err = c.buildLLMsFull(ctx); err != nil {
		return fmt.Errorf("refresh llms-full: %w", err)
	}

	c.cache.Store(snap)

	return nil
}

func (c *Controller) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	log := zerolog.Ctx(ctx)

	refresh := func() {
		rctx, cancel := context.WithTimeout(ctx, refreshTimeout)
		defer cancel()

		if err := c.Refresh(rctx); err != nil {
			log.Error().Err(err).Msg("feeds refresh failed")
			return
		}

		log.Debug().Msg("feeds refreshed")
	}

	refresh()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh()
		}
	}
}
