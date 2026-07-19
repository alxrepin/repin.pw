package telegram

import (
	"time"

	"github.com/gotd/td/tg"

	"repin/internal/context/domain"
)

type RawMessageFactory struct{}

func NewRawMessageFactory() *RawMessageFactory { return &RawMessageFactory{} }

func (f *RawMessageFactory) Create(msg *tg.Message) domain.RawMessage {
	raw := domain.RawMessage{
		ID:          msg.ID,
		Date:        time.Unix(int64(msg.Date), 0),
		GroupID:     msg.GroupedID,
		InvertMedia: msg.InvertMedia,
	}

	if msg.Message != "" {
		raw.Text = &msg.Message
	}

	if msg.Media != nil {
		if media := f.extractMedia(msg.Media); media != nil {
			raw.Media = media
		}
	}

	if msg.Entities != nil {
		raw.Entities = f.convertEntities(msg.Entities)
	}

	return raw
}

func (f *RawMessageFactory) extractMedia(media tg.MessageMediaClass) *domain.Media {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := m.Photo.(*tg.Photo)
		if !ok || photo == nil {
			return nil
		}

		return extractPhoto(photo)

	case *tg.MessageMediaDocument:
		doc, ok := m.Document.(*tg.Document)
		if !ok || doc == nil {
			return nil
		}

		return extractDocument(doc)
	}

	return nil
}

func extractPhoto(photo *tg.Photo) *domain.Media {
	var (
		sizeType string
		w, h     int
	)

	for _, s := range photo.Sizes {
		switch ps := s.(type) {
		case *tg.PhotoSize:
			if ps.W*ps.H > w*h {
				sizeType, w, h = ps.Type, ps.W, ps.H
			}
		case *tg.PhotoSizeProgressive:
			if ps.W*ps.H > w*h {
				sizeType, w, h = ps.Type, ps.W, ps.H
			}
		}
	}

	if sizeType == "" {
		return nil
	}

	return &domain.Media{
		Type:          domain.MediaTypePhoto,
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: append([]byte(nil), photo.FileReference...),
		PhotoSizeType: sizeType,
		MimeType:      "image/jpeg",
		Width:         w,
		Height:        h,
	}
}

func extractDocument(doc *tg.Document) *domain.Media {
	media := domain.Media{
		Type:          domain.MediaTypeDocument,
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: append([]byte(nil), doc.FileReference...),
		MimeType:      doc.MimeType,
		Size:          doc.Size,
	}

	animated := false

	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeSticker:
			return nil // stickers are decorative, not post content
		case *tg.DocumentAttributeVideo:
			media.Type = domain.MediaTypeVideo
			if a.RoundMessage {
				media.Type = domain.MediaTypeVideoNote
			}

			media.Width, media.Height, media.Duration = a.W, a.H, a.Duration
		case *tg.DocumentAttributeAudio:
			media.Type = domain.MediaTypeAudio
			if a.Voice {
				media.Type = domain.MediaTypeVoice
			}

			media.Duration = float64(a.Duration)
		case *tg.DocumentAttributeAnimated:
			animated = true
		case *tg.DocumentAttributeFilename:
			name := a.FileName
			media.FileName = &name
		case *tg.DocumentAttributeImageSize:
			media.Width, media.Height = a.W, a.H
		}
	}

	// Telegram GIFs are mp4 documents carrying both Animated and Video
	// attributes; Animated wins over the plain video classification.
	if animated {
		media.Type = domain.MediaTypeGIF
	}

	return &media
}

func (f *RawMessageFactory) convertEntities(entities []tg.MessageEntityClass) []domain.RawMessageEntity {
	result := make([]domain.RawMessageEntity, 0, len(entities))

	for _, e := range entities {
		var entity domain.RawMessageEntity

		switch ent := e.(type) {
		case *tg.MessageEntityTextURL:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeTextLink, Offset: ent.Offset, Length: ent.Length, URL: &ent.URL}
		case *tg.MessageEntityURL:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeURL, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntityBold:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeBold, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntityItalic:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeItalic, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntityUnderline:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeUnderline, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntityStrike:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeStrike, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntitySpoiler:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeSpoiler, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntityBlockquote:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeBlockquote, Offset: ent.Offset, Length: ent.Length, Collapsed: ent.Collapsed}
		case *tg.MessageEntityCode:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeCode, Offset: ent.Offset, Length: ent.Length}
		case *tg.MessageEntityPre:
			entity = domain.RawMessageEntity{Type: domain.EntityTypePre, Offset: ent.Offset, Length: ent.Length}
			if ent.Language != "" {
				lang := ent.Language
				entity.Language = &lang
			}
		case *tg.MessageEntityCustomEmoji:
			entity = domain.RawMessageEntity{Type: domain.EntityTypeCustomEmoji, Offset: ent.Offset, Length: ent.Length, CustomEmojiID: &ent.DocumentID}
		default:
			continue
		}

		result = append(result, entity)
	}

	return result
}
