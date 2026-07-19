package service

import (
	"context"
	"fmt"

	"repin/internal/context/domain"
)

type channelGetter interface {
	Get(ctx context.Context) (*domain.Channel, error)
}

type ChannelService struct {
	channel channelGetter
}

func NewChannelService(channel channelGetter) *ChannelService {
	return &ChannelService{channel: channel}
}

func (s *ChannelService) Get(ctx context.Context) (*domain.Channel, error) {
	channel, err := s.channel.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	return channel, nil
}
