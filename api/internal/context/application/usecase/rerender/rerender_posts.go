package rerender

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/text"
)

const batchSize = 500

type postStore interface {
	ListAfter(ctx context.Context, afterID int64, limit int) ([]domain.Post, error)
	Upsert(ctx context.Context, post *domain.Post) error
}

type channelStore interface {
	Get(ctx context.Context) (*domain.Channel, error)
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
	channels channelStore
	tx       txRunner
	renderer *text.Renderer
}

func NewRerenderPosts(posts postStore, channels channelStore, tx txRunner) *RerenderPosts {
	return &RerenderPosts{posts: posts, channels: channels, tx: tx, renderer: text.NewRenderer()}
}

func (uc *RerenderPosts) Execute(ctx context.Context) (Stats, error) {
	log := zerolog.Ctx(ctx)

	channel, err := uc.channels.Get(ctx)
	if err != nil {
		return Stats{}, fmt.Errorf("load channel: %w", err)
	}

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

		err = uc.tx.RunInTx(ctx, func(ctx context.Context) error {
			for i := range posts {
				post := &posts[i]

				if post.RawText == nil {
					stats.Skipped++
					continue
				}

				title, body := uc.renderer.Render(*post.RawText, post.Entities, channel.Name)

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

		afterID = posts[len(posts)-1].ID
	}

	log.Info().Int("rendered", stats.Rendered).Int("skipped", stats.Skipped).Msg("posts re-rendered")

	return stats, nil
}
