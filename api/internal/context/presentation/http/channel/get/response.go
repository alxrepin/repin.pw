package get

import (
	"repin/internal/context/domain"
	"repin/internal/pkg/httpx"
)

type Data struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	Subscriptions int64   `json:"subscriptions"`
}

func toResponse(channel *domain.Channel) httpx.APIResponse[Data, any] {
	data := Data{
		ID:            channel.ID,
		Name:          channel.Name,
		Title:         channel.Title,
		Description:   channel.Description,
		Avatar:        channel.Avatar,
		Subscriptions: channel.Subscriptions,
	}

	return httpx.NewAPIResponse[Data, any](&data, nil, nil, nil)
}
