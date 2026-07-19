package rerender

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/text"
)

type postStore interface {
	All(ctx context.Context) ([]domain.Post, error)
	Upsert(ctx context.Context, post *domain.Post) error
}

type txRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Stats struct {
	Rendered int
	Skipped  int // posts imported before raw captions were stored
}

type RerenderPosts struct {
	posts    postStore
	tx       txRunner
	renderer *text.Renderer
}

func NewRerenderPosts(posts postStore, tx txRunner) *RerenderPosts {
	return &RerenderPosts{posts: posts, tx: tx, renderer: text.NewRenderer()}
}

func (uc *RerenderPosts) Execute(ctx context.Context) (Stats, error) {
	log := zerolog.Ctx(ctx)

	posts, err := uc.posts.All(ctx)
	if err != nil {
		return Stats{}, fmt.Errorf("load posts: %w", err)
	}

	var stats Stats

	err = uc.tx.RunInTx(ctx, func(ctx context.Context) error {
		for i := range posts {
			post := &posts[i]

			if post.RawText == nil {
				stats.Skipped++
				continue
			}

			title, body := uc.renderer.Render(*post.RawText, post.Entities)

			post.Text = &body

			post.Title = nil
			if title != "" {
				post.Title = &title
			}

			if err := uc.posts.Upsert(ctx, post); err != nil {
				return fmt.Errorf("save post %d: %w", post.ID, err)
			}

			stats.Rendered++
		}

		return nil
	})
	if err != nil {
		return stats, err
	}

	log.Info().Int("rendered", stats.Rendered).Int("skipped", stats.Skipped).Msg("posts re-rendered")

	return stats, nil
}
