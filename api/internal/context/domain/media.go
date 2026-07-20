package domain

import (
	"fmt"
	"time"
)

type PostMedia struct {
	TgMessageID int64
	PostID      int64
	FileID      int64
	Type        MediaType
	ObjectKey   string
	MimeType    *string
	FileName    *string
	Size        *int64
	Width       *int
	Height      *int
	Duration    *float64 // seconds
	CreatedAt   time.Time
}

func NewPostMedia(postID int64, msg RawMessage) *PostMedia {
	media := *msg.Media

	return &PostMedia{
		TgMessageID: int64(msg.ID),
		PostID:      postID,
		FileID:      media.ID,
		Type:        media.Type,
		MimeType:    nilIfEmpty(media.MimeType),
		FileName:    media.FileName,
		Size:        nilIfZero(media.Size),
		Width:       nilIfZero(media.Width),
		Height:      nilIfZero(media.Height),
		Duration:    nilIfZero(media.Duration),
		CreatedAt:   msg.Date,
	}
}

func (m Media) ContentType() string {
	if m.MimeType != "" {
		return m.MimeType
	}

	switch m.Type {
	case MediaTypePhoto:
		return "image/jpeg"
	case MediaTypeVideo, MediaTypeVideoNote, MediaTypeGIF:
		return "video/mp4"
	case MediaTypeAudio:
		return "audio/mpeg"
	case MediaTypeVoice:
		return "audio/ogg"
	default:
		return "application/octet-stream"
	}
}

func (m Media) ObjectName() string {
	return fmt.Sprintf("media/%d%s", m.ID, m.ext())
}

func (m Media) ext() string {
	switch m.ContentType() {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "audio/mpeg":
		return ".mp3"
	case "audio/ogg":
		return ".ogg"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// nilIfZero maps Telegram's "attribute absent" zero values onto NULL.
func nilIfZero[T int | int64 | float64](v T) *T {
	if v == 0 {
		return nil
	}

	return &v
}
