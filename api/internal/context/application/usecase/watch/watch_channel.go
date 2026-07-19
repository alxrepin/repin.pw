package watch

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/rs/zerolog"

	syncuc "repin/internal/context/application/usecase/sync"
	"repin/internal/context/domain"
)

const albumOverlap = 10

type rawMessages interface {
	GetByIDs(ctx context.Context, username string, ids []int) ([]domain.RawMessage, error)
	GetChannelInfo(ctx context.Context, username string) (*domain.Channel, error)
}

type importer interface {
	ImportMessages(ctx context.Context, messages []domain.RawMessage) syncuc.Stats
}

type postStore interface {
	Delete(ctx context.Context, id int64) error
}

type mediaStore interface {
	GetByMessageID(ctx context.Context, id int64) (*domain.PostMedia, error)
	ListByPostIDs(ctx context.Context, ids []int64) (map[int64][]domain.PostMedia, error)
	DeleteByMessageID(ctx context.Context, id int64) error
	DeleteStale(ctx context.Context, postID int64, keepIDs []int64) error
}

type channelStore interface {
	Get(ctx context.Context) (*domain.Channel, error)
	Upsert(ctx context.Context, channel *domain.Channel) error
}

type txRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type WatchChannel struct {
	raw      rawMessages
	importer importer
	posts    postStore
	media    mediaStore
	channels channelStore
	tx       txRunner
}

func NewWatchChannel(
	raw rawMessages,
	importer importer,
	posts postStore,
	media mediaStore,
	channels channelStore,
	tx txRunner,
) *WatchChannel {
	return &WatchChannel{
		raw:      raw,
		importer: importer,
		posts:    posts,
		media:    media,
		channels: channels,
		tx:       tx,
	}
}

func (uc *WatchChannel) SyncPosts(ctx context.Context, username string, ids []int) error {
	log := zerolog.Ctx(ctx)

	if len(ids) == 0 {
		return nil
	}

	fetched, err := uc.raw.GetByIDs(ctx, username, expandAlbumRange(ids))
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}

	relevant := relatedMessages(fetched, ids)
	if len(relevant) == 0 {
		return nil
	}

	stats := uc.importer.ImportMessages(ctx, relevant)
	if stats.Errors > 0 {
		return fmt.Errorf("import messages %v: %d group(s) failed", ids, stats.Errors)
	}

	if err := uc.bumpLastMessageID(ctx, relevant); err != nil {
		return err
	}

	log.Info().Ints("ids", ids).Int("posts", stats.Posts).Msg("posts synced")

	return nil
}

func (uc *WatchChannel) DeletePosts(ctx context.Context, username string, ids []int) error {
	log := zerolog.Ctx(ctx)

	if len(ids) == 0 {
		return nil
	}

	deleted := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		deleted[int64(id)] = struct{}{}
	}

	survivors, err := uc.survivingAlbumMembers(ctx, deleted)
	if err != nil {
		return err
	}

	err = uc.tx.RunInTx(ctx, func(ctx context.Context) error {
		for id := range deleted {
			if err := uc.media.DeleteStale(ctx, id, nil); err != nil {
				return fmt.Errorf("delete media of post %d: %w", id, err)
			}

			if err := uc.media.DeleteByMessageID(ctx, id); err != nil {
				return fmt.Errorf("delete media row %d: %w", id, err)
			}

			if err := uc.posts.Delete(ctx, id); err != nil {
				return fmt.Errorf("delete post %d: %w", id, err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.Info().Ints("ids", ids).Msg("posts deleted")

	if len(survivors) == 0 {
		return nil
	}

	return uc.SyncPosts(ctx, username, survivors)
}

func (uc *WatchChannel) RefreshChannel(ctx context.Context, username string) error {
	log := zerolog.Ctx(ctx)

	channel, err := uc.raw.GetChannelInfo(ctx, username)
	if err != nil {
		return fmt.Errorf("fetch channel info: %w", err)
	}

	if err := uc.channels.Upsert(ctx, channel); err != nil {
		return fmt.Errorf("save channel: %w", err)
	}

	log.Info().Str("title", channel.Title).Int64("subscribers", channel.Subscriptions).Msg("channel info refreshed")

	return nil
}

func (uc *WatchChannel) survivingAlbumMembers(ctx context.Context, deleted map[int64]struct{}) ([]int, error) {
	touched := make(map[int64]struct{})

	for id := range deleted {
		row, err := uc.media.GetByMessageID(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrMediaNotFound) {
				continue
			}

			return nil, fmt.Errorf("check media of %d: %w", id, err)
		}

		touched[row.PostID] = struct{}{}
	}

	if len(touched) == 0 {
		return nil, nil
	}

	postIDs := make([]int64, 0, len(touched))
	for id := range touched {
		postIDs = append(postIDs, id)
	}

	byPost, err := uc.media.ListByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, fmt.Errorf("list album media: %w", err)
	}

	var survivors []int

	for _, rows := range byPost {
		for _, row := range rows {
			if _, gone := deleted[row.TgMessageID]; !gone {
				survivors = append(survivors, int(row.TgMessageID))
			}
		}
	}

	sort.Ints(survivors)

	return survivors, nil
}

func (uc *WatchChannel) bumpLastMessageID(ctx context.Context, messages []domain.RawMessage) error {
	channel, err := uc.channels.Get(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrChannelNotFound) {
			return nil // no channel row yet: the first full sync will create it
		}

		return fmt.Errorf("get channel: %w", err)
	}

	last := channel.LastMessageID
	for _, m := range messages {
		last = max(last, int64(m.ID))
	}

	if last == channel.LastMessageID {
		return nil
	}

	channel.LastMessageID = last

	if err := uc.channels.Upsert(ctx, channel); err != nil {
		return fmt.Errorf("save sync mark: %w", err)
	}

	return nil
}

func expandAlbumRange(ids []int) []int {
	seen := make(map[int]struct{})

	var out []int

	for _, id := range ids {
		for i := max(1, id-albumOverlap); i <= id+albumOverlap; i++ {
			if _, ok := seen[i]; ok {
				continue
			}

			seen[i] = struct{}{}
			out = append(out, i)
		}
	}

	sort.Ints(out)

	return out
}

func relatedMessages(messages []domain.RawMessage, ids []int) []domain.RawMessage {
	wanted := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}

	groups := make(map[int64]struct{})

	for _, m := range messages {
		if _, ok := wanted[m.ID]; ok && m.GroupID != 0 {
			groups[m.GroupID] = struct{}{}
		}
	}

	var out []domain.RawMessage

	for _, m := range messages {
		if _, ok := wanted[m.ID]; ok {
			out = append(out, m)
			continue
		}

		if m.GroupID != 0 {
			if _, ok := groups[m.GroupID]; ok {
				out = append(out, m)
			}
		}
	}

	return out
}
