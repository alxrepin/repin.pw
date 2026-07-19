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
	channel channelService
}

func NewController(channel channelService) *Controller {
	return &Controller{channel: channel}
}

func (c *Controller) Handle(ctx context.Context, _ Request) (httpx.APIResponse[Data, any], error) {
	channel, err := c.channel.Get(ctx)
	if err != nil {
		return httpx.APIResponse[Data, any]{}, err
	}

	return toResponse(channel), nil
}
