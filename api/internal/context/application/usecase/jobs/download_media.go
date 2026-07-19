package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
)

type rawMessages interface {
	GetByIDs(ctx context.Context, username string, ids []int) ([]domain.RawMessage, error)
	DownloadMedia(ctx context.Context, media domain.Media) ([]byte, error)
}

type mediaStore interface {
	GetByMessageID(ctx context.Context, id int64) (*domain.PostMedia, error)
	Upsert(ctx context.Context, media *domain.PostMedia) error
}

type storage interface {
	Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
}

type DownloadMedia struct {
	raw     rawMessages
	media   mediaStore
	storage storage
	channel string
}

func NewDownloadMedia(raw rawMessages, media mediaStore, storage storage, channel string) *DownloadMedia {
	return &DownloadMedia{raw: raw, media: media, storage: storage, channel: channel}
}

func (uc *DownloadMedia) Handle(ctx context.Context, job domain.Job) error {
	log := zerolog.Ctx(ctx)

	var payload domain.MediaDownloadPayload
	if err := job.Decode(&payload); err != nil {
		return err
	}

	messages, err := uc.raw.GetByIDs(ctx, uc.channel, []int{int(payload.MessageID)})
	if err != nil {
		return fmt.Errorf("fetch message %d: %w", payload.MessageID, err)
	}

	if len(messages) == 0 || messages[0].Media == nil {
		log.Info().Int64("message_id", payload.MessageID).Msg("media job skipped: message is gone")

		return nil
	}

	msg := messages[0]
	media := *msg.Media

	stored, err := uc.media.GetByMessageID(ctx, payload.MessageID)
	if err != nil && !errors.Is(err, domain.ErrMediaNotFound) {
		return fmt.Errorf("check stored media: %w", err)
	}

	if err == nil && stored.FileID == media.ID {
		log.Debug().Int64("message_id", payload.MessageID).Msg("media unchanged, skipping")

		return nil
	}

	started := time.Now()

	data, err := uc.raw.DownloadMedia(ctx, media)
	if err != nil {
		return fmt.Errorf("download media of message %d: %w", payload.MessageID, err)
	}

	key, err := uc.storage.Upload(ctx, media.ObjectName(), data, media.ContentType())
	if err != nil {
		return fmt.Errorf("upload media of message %d: %w", payload.MessageID, err)
	}

	row := domain.NewPostMedia(payload.PostID, msg)
	row.ObjectKey = key

	if err := uc.media.Upsert(ctx, row); err != nil {
		return fmt.Errorf("save media of message %d: %w", payload.MessageID, err)
	}

	log.Info().
		Int64("post_id", payload.PostID).
		Int64("message_id", payload.MessageID).
		Str("type", string(media.Type)).
		Int("bytes", len(data)).
		Dur("elapsed", time.Since(started)).
		Msg("media stored")

	return nil
}
