package domain

import "time"

type Channel struct {
	ID            int64
	Name          string
	Title         string
	Description   *string
	Avatar        *string
	Subscriptions int64
	LastMessageID int64
	CreatedAt     time.Time
}
