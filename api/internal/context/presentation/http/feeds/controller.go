package feeds

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
)

const (
	rssLimit      = 50   // feed readers only care about the newest entries
	llmsLimit     = 1000 // llms.txt is an index, not a database dump
	llmsFullLimit = 200  // full texts inline — keep the file ingestible
)

type postService interface {
	Latest(ctx context.Context, n int) ([]domain.Post, error)
	ListAfter(ctx context.Context, afterID int64, limit int) ([]domain.Post, error)
}

type channelService interface {
	Get(ctx context.Context) (*domain.Channel, error)
}

type Controller struct {
	posts    postService
	channel  channelService
	siteURL  string
	mediaURL string

	cache atomic.Pointer[snapshot]
}

func NewController(posts postService, channel channelService, siteURL, mediaURL string) *Controller {
	return &Controller{posts: posts, channel: channel, siteURL: siteURL, mediaURL: mediaURL}
}

func (c *Controller) siteDescription(src *source) string {
	if src.Channel.Description != nil && *src.Channel.Description != "" {
		return *src.Channel.Description
	}

	return "Зеркало Telegram-канала @" + src.Channel.Name
}

type source struct {
	Channel *domain.Channel
	Posts   []domain.Post
}

func (s *source) head(n int) *source {
	if len(s.Posts) <= n {
		return s
	}

	return &source{Channel: s.Channel, Posts: s.Posts[:n]}
}

func (c *Controller) load(ctx context.Context, limit int) (*source, error) {
	channel, err := c.channel.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("feed channel: %w", err)
	}

	posts, err := c.posts.Latest(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("feed posts: %w", err)
	}

	return &source{Channel: channel, Posts: posts}, nil
}

func (c *Controller) buildRSS(ctx context.Context) ([]byte, error) {
	src, err := c.load(ctx, rssLimit)
	if err != nil {
		return nil, err
	}

	return c.renderRSS(src)
}

func (c *Controller) buildLLMs(ctx context.Context) ([]byte, error) {
	src, err := c.load(ctx, llmsLimit)
	if err != nil {
		return nil, err
	}

	return c.renderLLMs(src)
}

func (c *Controller) buildLLMsFull(ctx context.Context) ([]byte, error) {
	src, err := c.load(ctx, llmsFullLimit)
	if err != nil {
		return nil, err
	}

	return c.renderLLMsFull(src)
}

func (c *Controller) handle(contentType string, pick func(*snapshot) []byte, live func(ctx context.Context) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if snap := c.cache.Load(); snap != nil {
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=300")
			w.Header().Set("Last-Modified", snap.generatedAt.Format(http.TimeFormat))

			if notModifiedSince(r, snap.generatedAt) {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			_, _ = w.Write(pick(snap))

			return
		}

		body, err := live(r.Context())
		if err != nil {
			zerolog.Ctx(r.Context()).Error().Err(err).Str("path", r.URL.Path).Msg("feed failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(body)
	}
}

func notModifiedSince(r *http.Request, generatedAt time.Time) bool {
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" {
		return false
	}

	since, err := http.ParseTime(ims)
	if err != nil {
		return false
	}

	return !generatedAt.Truncate(time.Second).After(since)
}

func (c *Controller) Sitemap() http.HandlerFunc {
	return c.handle("application/xml; charset=utf-8",
		func(s *snapshot) []byte { return s.sitemap }, c.renderSitemap)
}

func (c *Controller) RSS() http.HandlerFunc {
	return c.handle("application/rss+xml; charset=utf-8",
		func(s *snapshot) []byte { return s.rss }, c.buildRSS)
}

func (c *Controller) LLMs() http.HandlerFunc {
	return c.handle("text/plain; charset=utf-8",
		func(s *snapshot) []byte { return s.llms }, c.buildLLMs)
}

func (c *Controller) LLMsFull() http.HandlerFunc {
	return c.handle("text/plain; charset=utf-8",
		func(s *snapshot) []byte { return s.llmsFull }, c.buildLLMsFull)
}

func (c *Controller) postURL(p *domain.Post) string {
	slug := strconv.FormatInt(p.ID, 10)
	if p.URL != nil && *p.URL != "" {
		slug = *p.URL
	}

	return c.siteURL + "/posts/" + slug
}

func postTitle(p *domain.Post) string {
	if p.SEOTitle != nil && *p.SEOTitle != "" {
		return *p.SEOTitle
	}

	if p.Title != nil && *p.Title != "" {
		return *p.Title
	}

	return "Пост №" + strconv.FormatInt(p.ID, 10)
}

func postDescription(p *domain.Post) string {
	if p.SEODescription != nil {
		return *p.SEODescription
	}

	return ""
}

func firstPhoto(p *domain.Post) *domain.PostMedia {
	for i := range p.Media {
		if p.Media[i].Type == domain.MediaTypePhoto {
			return &p.Media[i]
		}
	}

	return nil
}
