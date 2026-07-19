package list

import (
	"time"

	"repin/internal/context/application/service"
	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/text"
	"repin/internal/context/presentation/http/media"
	"repin/internal/pkg/httpx"
)

const excerptRunes = 200

type Item struct {
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

func toResponse(page service.PostPage, mediaURL string) httpx.APIResponse[any, Item] {
	items := make([]Item, len(page.Posts))
	for i, post := range page.Posts {
		items[i] = toItem(post, mediaURL)
	}

	paginate := &httpx.Paginate{Page: page.Page, Limit: page.Limit, Total: page.Total}

	return httpx.NewAPIResponse[any](nil, items, paginate, nil)
}

func toItem(post domain.Post, mediaURL string) Item {
	item := Item{
		ID:        post.ID,
		GroupID:   post.GroupID,
		Title:     post.Title,
		URL:       post.URL,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}

	if post.Text != nil {
		excerpt := text.Excerpt(*post.Text, excerptRunes)
		item.Text = &excerpt
	}

	for _, m := range post.Media {
		if m.Type == domain.MediaTypePhoto {
			item.Cover = &Cover{URL: media.URL(mediaURL, m.ObjectKey), Width: m.Width, Height: m.Height}
			break
		}
	}

	return item
}
