package media

import (
	"time"

	"repin/internal/context/domain"
)

type media struct {
	TgMessageID int64      `db:"tg_message_id"`
	PostID      int64      `db:"post_id"`
	FileID      int64      `db:"file_id"`
	Type        string     `db:"type"`
	ObjectKey   string     `db:"object_key"`
	MimeType    *string    `db:"mime_type"`
	FileName    *string    `db:"file_name"`
	SizeBytes   *int64     `db:"size_bytes"`
	Width       *int       `db:"width"`
	Height      *int       `db:"height"`
	Duration    *float64   `db:"duration"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

func (m media) ToDomain() *domain.PostMedia {
	return &domain.PostMedia{
		TgMessageID: m.TgMessageID,
		PostID:      m.PostID,
		FileID:      m.FileID,
		Type:        domain.MediaType(m.Type),
		ObjectKey:   m.ObjectKey,
		MimeType:    m.MimeType,
		FileName:    m.FileName,
		Size:        m.SizeBytes,
		Width:       m.Width,
		Height:      m.Height,
		Duration:    m.Duration,
		CreatedAt:   m.CreatedAt,
	}
}

func fromDomain(m *domain.PostMedia) media {
	return media{
		TgMessageID: m.TgMessageID,
		PostID:      m.PostID,
		FileID:      m.FileID,
		Type:        string(m.Type),
		ObjectKey:   m.ObjectKey,
		MimeType:    m.MimeType,
		FileName:    m.FileName,
		SizeBytes:   m.Size,
		Width:       m.Width,
		Height:      m.Height,
		Duration:    m.Duration,
		CreatedAt:   m.CreatedAt,
	}
}
