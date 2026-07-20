package regenseo

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
)

const batchSize = 500

type postStore interface {
	ListAfter(ctx context.Context, afterID int64, limit int) ([]domain.Post, error)
}

type jobQueue interface {
	Enqueue(ctx context.Context, kind domain.JobKind, dedupKey string, payload any) error
}

type Stats struct {
	Enqueued int
	Skipped  int // no raw text to describe, or SEO already present with --missing-only
}

type RegenerateSEO struct {
	posts postStore
	jobs  jobQueue
}

func NewRegenerateSEO(posts postStore, jobs jobQueue) *RegenerateSEO {
	return &RegenerateSEO{posts: posts, jobs: jobs}
}

func (uc *RegenerateSEO) Execute(ctx context.Context, missingOnly bool) (Stats, error) {
	var (
		stats   Stats
		afterID int64
	)

	for {
		posts, err := uc.posts.ListAfter(ctx, afterID, batchSize)
		if err != nil {
			return stats, fmt.Errorf("load posts after %d: %w", afterID, err)
		}

		if len(posts) == 0 {
			break
		}

		for i := range posts {
			post := &posts[i]

			if !wantsSEO(post, missingOnly) {
				stats.Skipped++
				continue
			}

			payload := domain.GenerateSEOPayload{PostID: post.ID}
			if err := uc.jobs.Enqueue(ctx, domain.JobKindGenerateSEO, payload.DedupKey(), payload); err != nil {
				return stats, fmt.Errorf("enqueue seo job for post %d: %w", post.ID, err)
			}

			stats.Enqueued++
		}

		afterID = posts[len(posts)-1].ID
	}

	zerolog.Ctx(ctx).Info().
		Int("enqueued", stats.Enqueued).Int("skipped", stats.Skipped).
		Msg("seo regeneration queued")

	return stats, nil
}

func wantsSEO(post *domain.Post, missingOnly bool) bool {
	if post.RawText == nil || *post.RawText == "" {
		return false
	}

	if !missingOnly {
		return true
	}

	return post.SEOTitle == nil || post.SEODescription == nil || post.SEOKeywords == nil
}
