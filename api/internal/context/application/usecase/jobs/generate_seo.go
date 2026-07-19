package jobs

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/openrouter"
)

type postStore interface {
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	UpdateSEO(ctx context.Context, id int64, seo domain.PostSEO) error
}

type seoGenerator interface {
	Generate(ctx context.Context, post *domain.Post) (*domain.PostSEO, error)
}

type GenerateSEO struct {
	posts     postStore
	generator seoGenerator
}

func NewGenerateSEO(posts postStore, generator seoGenerator) *GenerateSEO {
	return &GenerateSEO{posts: posts, generator: generator}
}

func (uc *GenerateSEO) Handle(ctx context.Context, job domain.Job) error {
	log := zerolog.Ctx(ctx)

	var payload domain.GenerateSEOPayload
	if err := job.Decode(&payload); err != nil {
		return err
	}

	post, err := uc.posts.GetByID(ctx, payload.PostID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			log.Info().Int64("post_id", payload.PostID).Msg("seo job skipped: post is gone")

			return nil
		}

		return fmt.Errorf("load post %d: %w", payload.PostID, err)
	}

	seo, err := uc.generator.Generate(ctx, post)
	if err != nil {
		if errors.Is(err, openrouter.ErrNothingToDescribe) {
			log.Info().Int64("post_id", payload.PostID).Msg("seo job skipped: post has no text")

			return nil
		}

		return fmt.Errorf("generate seo for post %d: %w", payload.PostID, err)
	}

	if err := uc.posts.UpdateSEO(ctx, payload.PostID, *seo); err != nil {
		return fmt.Errorf("save seo of post %d: %w", payload.PostID, err)
	}

	log.Info().
		Int64("post_id", payload.PostID).
		Str("seo_title", seo.Title).
		Msg("seo generated")

	return nil
}
