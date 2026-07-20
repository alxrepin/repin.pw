package sync

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"repin/internal/context/application/usecase/jobs"
	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/text"
)

const albumOverlap = 10

type rawMessages interface {
	WithSession(ctx context.Context, fn func(ctx context.Context) error) error
	GetAll(ctx context.Context, username string, minID int) ([]domain.RawMessage, error)
	GetChannelInfo(ctx context.Context, username string) (*domain.Channel, error)
}

type postWriter interface {
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	Upsert(ctx context.Context, post *domain.Post) error
}

type channelStore interface {
	Get(ctx context.Context) (*domain.Channel, error)
	Upsert(ctx context.Context, channel *domain.Channel) error
}

type mediaStore interface {
	DeleteStale(ctx context.Context, postID int64, keepIDs []int64) error
}

type txRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Stats struct {
	Posts       int
	MediaQueued int
	SEOQueued   int
	Errors      int
}

type SyncChannel struct {
	raw      rawMessages
	posts    postWriter
	channels channelStore
	media    mediaStore
	jobs     jobs.Enqueuer
	tx       txRunner

	renderer *text.Renderer
}

func NewSyncChannel(
	raw rawMessages,
	posts postWriter,
	channels channelStore,
	media mediaStore,
	queue jobs.Enqueuer,
	tx txRunner,
) *SyncChannel {
	return &SyncChannel{
		raw:      raw,
		posts:    posts,
		channels: channels,
		media:    media,
		jobs:     queue,
		tx:       tx,
		renderer: text.NewRenderer(),
	}
}

func (uc *SyncChannel) Execute(ctx context.Context, username string) error {
	return uc.raw.WithSession(ctx, func(ctx context.Context) error {
		return uc.sync(ctx, username)
	})
}

func (uc *SyncChannel) sync(ctx context.Context, username string) error {
	log := zerolog.Ctx(ctx)
	started := time.Now()

	log.Info().Str("channel", username).Msg("sync started")

	channel, err := uc.raw.GetChannelInfo(ctx, username)
	if err != nil {
		return fmt.Errorf("fetch channel info: %w", err)
	}

	lastID, err := uc.lastMessageID(ctx)
	if err != nil {
		return err
	}

	channel.LastMessageID = int64(lastID)
	if err := uc.channels.Upsert(ctx, channel); err != nil {
		return fmt.Errorf("save channel: %w", err)
	}

	log.Info().Str("title", channel.Title).Int64("subscribers", channel.Subscriptions).Msg("channel saved")

	minID := max(0, lastID-albumOverlap)

	log.Info().Int("min_id", minID).Msg("fetching history")

	messages, err := uc.raw.GetAll(ctx, username, minID)
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}

	log.Info().Int("messages", len(messages)).Msg("history fetched")

	stats := uc.ImportMessages(ctx, messages, username)

	for _, m := range messages {
		channel.LastMessageID = max(channel.LastMessageID, int64(m.ID))
	}

	if err := uc.channels.Upsert(ctx, channel); err != nil {
		return fmt.Errorf("save sync mark: %w", err)
	}

	log.Info().
		Str("channel", username).
		Int("posts", stats.Posts).
		Int("media_queued", stats.MediaQueued).
		Int("seo_queued", stats.SEOQueued).
		Int("errors", stats.Errors).
		Dur("elapsed", time.Since(started)).
		Msg("channel synced")

	return nil
}

func (uc *SyncChannel) lastMessageID(ctx context.Context) (int, error) {
	channel, err := uc.channels.Get(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrChannelNotFound) {
			return 0, nil
		}

		return 0, fmt.Errorf("get last message id: %w", err)
	}

	return int(channel.LastMessageID), nil
}

func (uc *SyncChannel) ImportMessages(ctx context.Context, messages []domain.RawMessage, channel string) Stats {
	log := zerolog.Ctx(ctx)

	var stats Stats

	for _, group := range groupMessages(messages) {
		result, err := uc.importGroup(ctx, group, channel)
		if err != nil {
			stats.Errors++

			log.Error().Err(err).Int("message_id", group[0].ID).Msg("import failed")

			continue
		}

		if !result.imported {
			continue
		}

		stats.Posts++
		stats.MediaQueued += result.media
		stats.SEOQueued += result.seo
	}

	return stats
}

func groupMessages(messages []domain.RawMessage) [][]domain.RawMessage {
	albums := make(map[int64][]domain.RawMessage)

	var groups [][]domain.RawMessage

	for _, m := range messages {
		if m.GroupID == 0 {
			groups = append(groups, []domain.RawMessage{m})
			continue
		}

		albums[m.GroupID] = append(albums[m.GroupID], m)
	}

	for _, album := range albums {
		sort.Slice(album, func(i, j int) bool { return album[i].ID < album[j].ID })
		groups = append(groups, album)
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i][0].ID < groups[j][0].ID })

	return groups
}

type queued struct {
	imported bool
	media    int
	seo      int
}

func (uc *SyncChannel) importGroup(ctx context.Context, group []domain.RawMessage, channel string) (queued, error) {
	log := zerolog.Ctx(ctx)

	postID := int64(group[0].ID)

	post := uc.buildPost(postID, group, channel)
	if post == nil {
		return queued{}, nil // nothing publishable: no title and no media
	}

	existing, err := uc.posts.GetByID(ctx, postID)
	if err != nil && !errors.Is(err, domain.ErrPostNotFound) {
		return queued{}, fmt.Errorf("load post %d: %w", postID, err)
	}

	if existing != nil {
		post.SEOTitle = existing.SEOTitle
		post.SEODescription = existing.SEODescription
		post.SEOKeywords = existing.SEOKeywords
	}

	mediaIDs := mediaMessageIDs(group)
	wantSEO := needsSEO(existing, post)

	err = uc.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := uc.posts.Upsert(ctx, post); err != nil {
			return fmt.Errorf("save post: %w", err)
		}

		if err := uc.media.DeleteStale(ctx, postID, mediaIDs); err != nil {
			return fmt.Errorf("delete stale media: %w", err)
		}

		for _, id := range mediaIDs {
			if err := jobs.EnqueueMediaDownload(ctx, uc.jobs, postID, id); err != nil {
				return err
			}
		}

		if wantSEO {
			if err := jobs.EnqueueGenerateSEO(ctx, uc.jobs, postID); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return queued{}, err
	}

	result := queued{imported: true, media: len(mediaIDs)}
	if wantSEO {
		result.seo = 1
	}

	log.Info().
		Int64("post_id", postID).
		Str("title", stringOr(post.Title, "")).
		Int("media_queued", result.media).
		Bool("seo_queued", wantSEO).
		Msg("post imported")

	return result, nil
}

func (uc *SyncChannel) buildPost(postID int64, group []domain.RawMessage, channel string) *domain.Post {
	post := &domain.Post{
		ID:        postID,
		GroupID:   group[0].GroupID,
		CreatedAt: group[0].Date,
	}

	hasMedia := false

	var caption *domain.RawMessage

	for i := range group {
		if group[i].Media != nil {
			hasMedia = true
		}

		post.InvertMedia = post.InvertMedia || group[i].InvertMedia

		if group[i].Text == nil {
			continue
		}

		if caption == nil || len(*group[i].Text) > len(*caption.Text) {
			caption = &group[i]
		}
	}

	if caption != nil {
		post.RawText = caption.Text
		post.Entities = caption.Entities

		title, body := uc.renderer.Render(*caption.Text, caption.Entities, channel)

		post.Text = &body

		if title != "" {
			post.Title = &title
		}
	}

	if post.Title == nil && !hasMedia {
		return nil
	}

	url := strconv.FormatInt(postID, 10)
	if post.Title != nil {
		if slug := text.Slug(*post.Title); slug != "" {
			url += "-" + slug
		}
	}

	post.URL = &url

	return post
}

func needsSEO(existing, post *domain.Post) bool {
	if post.RawText == nil || *post.RawText == "" {
		return false
	}

	if existing == nil {
		return true
	}

	if existing.SEOTitle == nil || existing.SEODescription == nil || existing.SEOKeywords == nil {
		return true
	}

	return !equalStrings(existing.RawText, post.RawText)
}

func mediaMessageIDs(group []domain.RawMessage) []int64 {
	ids := make([]int64, 0, len(group))

	for _, msg := range group {
		if msg.Media != nil {
			ids = append(ids, int64(msg.ID))
		}
	}

	return ids
}

func equalStrings(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func stringOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}

	return *s
}
