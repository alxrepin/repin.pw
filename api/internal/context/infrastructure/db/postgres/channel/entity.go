package channel

import (
	"time"

	"repin/internal/context/domain"
)

type channel struct {
	ID            int64     `db:"id"`
	Name          string    `db:"name"`
	Title         string    `db:"title"`
	Description   *string   `db:"description"`
	Avatar        *string   `db:"avatar"`
	Subscriptions int64     `db:"subscriptions"`
	LastMessageID int64     `db:"last_message_id"`
	CreatedAt     time.Time `db:"created_at"`
}

func (c channel) ToDomain() *domain.Channel {
	return &domain.Channel{
		ID:            c.ID,
		Name:          c.Name,
		Title:         c.Title,
		Description:   c.Description,
		Avatar:        c.Avatar,
		Subscriptions: c.Subscriptions,
		LastMessageID: c.LastMessageID,
		CreatedAt:     c.CreatedAt,
	}
}

func fromDomain(c *domain.Channel) channel {
	return channel{
		ID:            c.ID,
		Name:          c.Name,
		Title:         c.Title,
		Description:   c.Description,
		Avatar:        c.Avatar,
		Subscriptions: c.Subscriptions,
		LastMessageID: c.LastMessageID,
		CreatedAt:     c.CreatedAt,
	}
}
