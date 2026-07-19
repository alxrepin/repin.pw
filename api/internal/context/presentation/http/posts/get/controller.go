package get

import (
	"context"

	"repin/internal/context/application/service"
	"repin/internal/pkg/httpx"
)

type postService interface {
	Get(ctx context.Context, slug string) (service.PostDetails, error)
}

type Request struct {
	Slug string `json:"slug"`
}

type Controller struct {
	posts    postService
	mediaURL string
}

func NewController(posts postService, mediaURL string) *Controller {
	return &Controller{posts: posts, mediaURL: mediaURL}
}

func (c *Controller) Handle(ctx context.Context, req Request) (httpx.APIResponse[Data, any], error) {
	details, err := c.posts.Get(ctx, req.Slug)
	if err != nil {
		return httpx.APIResponse[Data, any]{}, err
	}

	return toResponse(details, c.mediaURL), nil
}
