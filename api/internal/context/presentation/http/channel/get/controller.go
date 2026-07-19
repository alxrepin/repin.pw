package get

import (
	"context"

	"repin/internal/context/domain"
	"repin/internal/pkg/httpx"
)

type channelService interface {
	Get(ctx context.Context) (*domain.Channel, error)
}

type Request struct{}

type Controller struct {
	channel  channelService
	mediaURL string
}

func NewController(channel channelService, mediaURL string) *Controller {
	return &Controller{channel: channel, mediaURL: mediaURL}
}

func (c *Controller) Handle(ctx context.Context, _ Request) (httpx.APIResponse[Data, any], error) {
	channel, err := c.channel.Get(ctx)
	if err != nil {
		return httpx.APIResponse[Data, any]{}, err
	}

	return toResponse(channel, c.mediaURL), nil
}
