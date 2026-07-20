package get

import (
	"time"

	"repin/internal/context/application/service"
	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/text"
	"repin/internal/context/presentation/http/media"
	"repin/internal/pkg/httpx"
)

const excerptRunes = 200

type Data struct {
	ID             int64     `json:"id"`
	GroupID        int64     `json:"group_id"`
	Title          *string   `json:"title,omitempty"`
	URL            *string   `json:"url,omitempty"`
	Text           *string   `json:"text,omitempty"`
	InvertMedia    bool      `json:"invert_media"`
	SEOTitle       *string   `json:"seo_title,omitempty"`
	SEODescription *string   `json:"seo_description,omitempty"`
	SEOKeywords    *string   `json:"seo_keywords,omitempty"`
	Media          []Media   `json:"media"`
	Prev           *Adjacent `json:"prev,omitempty"`
	Next           *Adjacent `json:"next,omitempty"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      *string   `json:"updated_at,omitempty"`
}

type Adjacent struct {
	ID        int64   `json:"id"`
	GroupID   int64   `json:"group_id"`
	Title     *string `json:"title,omitempty"`
	URL       *string `json:"url,omitempty"`
	Text      *string `json:"text,omitempty"`
	Cover     *Cover  `json:"cover,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type Cover struct {
	URL    string `json:"url"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}

type Media struct {
	ID       int64    `json:"id"`
	Type     string   `json:"type"`
	URL      string   `json:"url"`
	MimeType *string  `json:"mime_type,omitempty"`
	FileName *string  `json:"file_name,omitempty"`
	Size     *int64   `json:"size,omitempty"`
	Width    *int     `json:"width,omitempty"`
	Height   *int     `json:"height,omitempty"`
	Duration *float64 `json:"duration,omitempty"`
}

func toMedia(m domain.PostMedia, mediaURL string) Media {
	return Media{
		ID:       m.TgMessageID,
		Type:     string(m.Type),
		URL:      media.URL(mediaURL, m.ObjectKey),
		MimeType: m.MimeType,
		FileName: m.FileName,
		Size:     m.Size,
		Width:    m.Width,
		Height:   m.Height,
		Duration: m.Duration,
	}
}

func toResponse(details service.PostDetails, mediaURL string) httpx.APIResponse[Data, any] {
	post := details.Post

	data := Data{
		ID:             post.ID,
		GroupID:        post.GroupID,
		Title:          post.Title,
		URL:            post.URL,
		Text:           post.Text,
		InvertMedia:    post.InvertMedia,
		SEOTitle:       post.SEOTitle,
		SEODescription: post.SEODescription,
		SEOKeywords:    post.SEOKeywords,
		Media:          make([]Media, 0, len(post.Media)),
		Prev:           toAdjacent(details.Prev, mediaURL),
		Next:           toAdjacent(details.Next, mediaURL),
		CreatedAt:      post.CreatedAt.Format(time.RFC3339),
	}

	if post.UpdatedAt != nil {
		data.UpdatedAt = new(post.UpdatedAt.Format(time.RFC3339))
	}

	for _, m := range post.Media {
		data.Media = append(data.Media, toMedia(m, mediaURL))
	}

	return httpx.NewAPIResponse[Data, any](&data, nil, nil, nil)
}

func toAdjacent(post *domain.Post, mediaURL string) *Adjacent {
	if post == nil {
		return nil
	}

	adj := Adjacent{
		ID:        post.ID,
		GroupID:   post.GroupID,
		Title:     post.Title,
		URL:       post.URL,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}

	if post.Text != nil {
		adj.Text = new(text.Excerpt(*post.Text, excerptRunes))
	}

	for _, m := range post.Media {
		if m.Type == domain.MediaTypePhoto {
			adj.Cover = &Cover{URL: media.URL(mediaURL, m.ObjectKey), Width: m.Width, Height: m.Height}
			break
		}
	}

	return &adj
}
