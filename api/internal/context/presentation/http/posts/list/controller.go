package list

import (
	"context"

	"repin/internal/context/application/service"
	"repin/internal/pkg/httpx"
)

type postService interface {
	List(ctx context.Context, page, limit int) (service.PostPage, error)
}

type Controller struct {
	posts    postService
	mediaURL string
}

func NewController(posts postService, mediaURL string) *Controller {
	return &Controller{posts: posts, mediaURL: mediaURL}
}

func (c *Controller) Handle(ctx context.Context, req Request) (httpx.APIResponse[any, Item], error) {
	page, err := c.posts.List(ctx, req.Page, req.Limit)
	if err != nil {
		return httpx.APIResponse[any, Item]{}, err
	}

	return toResponse(page, c.mediaURL), nil
}
