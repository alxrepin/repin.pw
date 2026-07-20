package telegram

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"

	"repin/internal/context/domain"
)

type Storage interface {
	Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
}

type RawMessageRepository struct {
	client     *Client
	factory    *RawMessageFactory
	storage    Storage
	faviconDir string
}

func NewRawMessageRepository(client *Client, storage Storage, faviconDir string) *RawMessageRepository {
	return &RawMessageRepository{client: client, factory: NewRawMessageFactory(), storage: storage, faviconDir: faviconDir}
}

func (r *RawMessageRepository) WithSession(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.client.WithSession(ctx, fn)
}

func (r *RawMessageRepository) run(ctx context.Context, fn func(ctx context.Context, client *telegram.Client) error) error {
	if client, ok := clientFrom(ctx); ok {
		return fn(ctx, client)
	}

	return r.client.Run(ctx, fn)
}

func (r *RawMessageRepository) GetAll(ctx context.Context, username string, minID int) ([]domain.RawMessage, error) {
	var messages []domain.RawMessage

	err := r.run(ctx, func(ctx context.Context, client *telegram.Client) error {
		msgs, err := r.client.FetchMessages(ctx, client, username, minID)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, r.factory.Create(m))
			}
		}

		return nil
	})

	return messages, err
}

func (r *RawMessageRepository) GetByID(ctx context.Context, username string, id int) (*domain.RawMessage, error) {
	var message *domain.RawMessage

	err := r.run(ctx, func(ctx context.Context, client *telegram.Client) error {
		msg, err := r.client.FetchMessage(ctx, client, username, id)
		if err != nil {
			return err
		}

		raw := r.factory.Create(msg)
		message = &raw

		return nil
	})

	return message, err
}

func (r *RawMessageRepository) GetByIDs(ctx context.Context, username string, ids []int) ([]domain.RawMessage, error) {
	var messages []domain.RawMessage

	err := r.run(ctx, func(ctx context.Context, client *telegram.Client) error {
		msgs, err := r.client.FetchMessagesByIDs(ctx, client, username, ids)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, r.factory.Create(m))
			}
		}

		return nil
	})

	return messages, err
}

func (r *RawMessageRepository) DownloadMedia(ctx context.Context, media domain.Media) ([]byte, error) {
	var data []byte

	err := r.run(ctx, func(ctx context.Context, client *telegram.Client) error {
		var location tg.InputFileLocationClass

		switch media.Type {
		case domain.MediaTypePhoto:
			location = &tg.InputPhotoFileLocation{
				ID:            media.ID,
				AccessHash:    media.AccessHash,
				FileReference: media.FileReference,
				ThumbSize:     media.PhotoSizeType,
			}
		case domain.MediaTypeVideo, domain.MediaTypeAudio, domain.MediaTypeVoice,
			domain.MediaTypeVideoNote, domain.MediaTypeGIF, domain.MediaTypeDocument:
			location = &tg.InputDocumentFileLocation{
				ID:            media.ID,
				AccessHash:    media.AccessHash,
				FileReference: media.FileReference,
			}
		default:
			return fmt.Errorf("unsupported media type: %s", media.Type)
		}

		buf := &bytes.Buffer{}
		if _, err := downloader.NewDownloader().Download(client.API(), location).Stream(ctx, buf); err != nil {
			return fmt.Errorf("stream media: %w", err)
		}

		data = buf.Bytes()

		return nil
	})

	return data, err
}

func (r *RawMessageRepository) GetChannelInfo(ctx context.Context, username string) (*domain.Channel, error) {
	var channel *domain.Channel

	err := r.run(ctx, func(ctx context.Context, client *telegram.Client) error {
		tgChannel, about, subscribers, err := r.client.FetchChannelInfo(ctx, client, username)
		if err != nil {
			return err
		}

		avatar, err := r.uploadAvatar(ctx, client, tgChannel)
		if err != nil {
			return err
		}

		var description *string
		if about != "" {
			description = &about
		}

		channel = &domain.Channel{
			ID:            tgChannel.ID,
			Name:          username,
			Title:         tgChannel.Title,
			Description:   description,
			Avatar:        avatar,
			Subscriptions: subscribers,
			CreatedAt:     time.Unix(int64(tgChannel.Date), 0),
		}

		return nil
	})

	return channel, err
}

func (r *RawMessageRepository) uploadAvatar(ctx context.Context, client *telegram.Client, tgChannel *tg.Channel) (*string, error) {
	photo, ok := tgChannel.Photo.(*tg.ChatPhoto)
	if !ok {
		return nil, nil
	}

	buf := &bytes.Buffer{}
	location := &tg.InputPeerPhotoFileLocation{
		Peer:    &tg.InputPeerChannel{ChannelID: tgChannel.ID, AccessHash: tgChannel.AccessHash},
		PhotoID: photo.PhotoID,
	}

	if _, err := downloader.NewDownloader().Download(client.API(), location).Stream(ctx, buf); err != nil {
		return nil, fmt.Errorf("download channel photo: %w", err)
	}

	object := fmt.Sprintf("channel_avatars/%d.jpg", tgChannel.ID)

	url, err := r.storage.Upload(ctx, object, buf.Bytes(), "image/jpeg")
	if err != nil {
		return nil, fmt.Errorf("upload channel photo: %w", err)
	}

	if err := r.writeFavicon(buf.Bytes()); err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *RawMessageRepository) writeFavicon(data []byte) error {
	if r.faviconDir == "" {
		return nil
	}

	if err := os.MkdirAll(r.faviconDir, 0o755); err != nil { //nolint:gosec // must stay readable for the web container user
		return fmt.Errorf("create favicon dir: %w", err)
	}

	target := filepath.Join(r.faviconDir, "favicon.jpg")
	tmp := target + ".tmp"

	if err := os.WriteFile(tmp, data, 0o644); err != nil { //nolint:gosec // world-readable static asset
		return fmt.Errorf("write favicon: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		return fmt.Errorf("replace favicon: %w", err)
	}

	return nil
}
