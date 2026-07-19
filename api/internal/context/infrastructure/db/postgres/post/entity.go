package post

import (
	"encoding/json"
	"fmt"
	"time"

	"repin/internal/context/domain"
)

type post struct {
	ID             int64      `db:"id"`
	GroupID        int64      `db:"group_id"`
	Title          *string    `db:"title"`
	URL            *string    `db:"url"`
	Text           *string    `db:"text"`
	RawText        *string    `db:"raw_text"`
	Entities       *string    `db:"entities"`
	InvertMedia    bool       `db:"invert_media"`
	SEOTitle       *string    `db:"seo_title"`
	SEODescription *string    `db:"seo_description"`
	SEOKeywords    *string    `db:"seo_keywords"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
}

type entity struct {
	Type          string  `json:"type"`
	Offset        int     `json:"offset"`
	Length        int     `json:"length"`
	URL           *string `json:"url,omitempty"`
	User          *int64  `json:"user,omitempty"`
	CustomEmojiID *int64  `json:"custom_emoji_id,omitempty"`
	Language      *string `json:"language,omitempty"`
	Collapsed     bool    `json:"collapsed,omitempty"`
}

func (p post) ToDomain() (*domain.Post, error) {
	var entities []domain.RawMessageEntity

	if p.Entities != nil {
		var rows []entity
		if err := json.Unmarshal([]byte(*p.Entities), &rows); err != nil {
			return nil, fmt.Errorf("decode entities of post %d: %w", p.ID, err)
		}

		entities = make([]domain.RawMessageEntity, 0, len(rows))
		for _, e := range rows {
			entities = append(entities, domain.RawMessageEntity{
				Type:          domain.RawMessageEntityType(e.Type),
				Offset:        e.Offset,
				Length:        e.Length,
				URL:           e.URL,
				User:          e.User,
				CustomEmojiID: e.CustomEmojiID,
				Language:      e.Language,
				Collapsed:     e.Collapsed,
			})
		}
	}

	return &domain.Post{
		ID:             p.ID,
		GroupID:        p.GroupID,
		Title:          p.Title,
		URL:            p.URL,
		Text:           p.Text,
		RawText:        p.RawText,
		Entities:       entities,
		InvertMedia:    p.InvertMedia,
		SEOTitle:       p.SEOTitle,
		SEODescription: p.SEODescription,
		SEOKeywords:    p.SEOKeywords,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}, nil
}

func fromDomain(p *domain.Post) (post, error) {
	row := post{
		ID:             p.ID,
		GroupID:        p.GroupID,
		Title:          p.Title,
		URL:            p.URL,
		Text:           p.Text,
		RawText:        p.RawText,
		InvertMedia:    p.InvertMedia,
		SEOTitle:       p.SEOTitle,
		SEODescription: p.SEODescription,
		SEOKeywords:    p.SEOKeywords,
		CreatedAt:      p.CreatedAt,
	}

	if p.Entities != nil {
		rows := make([]entity, 0, len(p.Entities))
		for _, e := range p.Entities {
			rows = append(rows, entity{
				Type:          string(e.Type),
				Offset:        e.Offset,
				Length:        e.Length,
				URL:           e.URL,
				User:          e.User,
				CustomEmojiID: e.CustomEmojiID,
				Language:      e.Language,
				Collapsed:     e.Collapsed,
			})
		}

		encoded, err := json.Marshal(rows)
		if err != nil {
			return post{}, fmt.Errorf("encode entities of post %d: %w", p.ID, err)
		}

		s := string(encoded)
		row.Entities = &s
	}

	return row, nil
}
