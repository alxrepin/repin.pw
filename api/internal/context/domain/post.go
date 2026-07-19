package domain

import "time"

type Post struct {
	ID             int64
	GroupID        int64
	Title          *string
	URL            *string
	Text           *string
	RawText        *string
	Entities       []RawMessageEntity
	InvertMedia    bool // Telegram "caption above media": render media below the text
	SEOTitle       *string
	SEODescription *string
	SEOKeywords    *string
	Media          []PostMedia
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
