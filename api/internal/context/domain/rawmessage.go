package domain

import "time"

type MediaType string

const (
	MediaTypePhoto     MediaType = "photo"
	MediaTypeVideo     MediaType = "video"
	MediaTypeAudio     MediaType = "audio"
	MediaTypeVoice     MediaType = "voice"
	MediaTypeVideoNote MediaType = "video_note"
	MediaTypeGIF       MediaType = "gif"
	MediaTypeDocument  MediaType = "document"
)

type Media struct {
	Type          MediaType
	ID            int64
	AccessHash    int64
	FileReference []byte
	PhotoSizeType string
	MimeType      string
	FileName      *string
	Size          int64
	Width         int
	Height        int
	Duration      float64 // seconds
}

type RawMessageEntityType string

const (
	EntityTypeURL         RawMessageEntityType = "url"
	EntityTypeBold        RawMessageEntityType = "bold"
	EntityTypeItalic      RawMessageEntityType = "italic"
	EntityTypeUnderline   RawMessageEntityType = "underline"
	EntityTypeStrike      RawMessageEntityType = "strike"
	EntityTypeSpoiler     RawMessageEntityType = "spoiler"
	EntityTypeBlockquote  RawMessageEntityType = "blockquote"
	EntityTypeCode        RawMessageEntityType = "code"
	EntityTypePre         RawMessageEntityType = "pre"
	EntityTypeTextLink    RawMessageEntityType = "text_link"
	EntityTypeCustomEmoji RawMessageEntityType = "custom_emoji"
)

type RawMessageEntity struct {
	Type          RawMessageEntityType
	Offset        int
	Length        int
	URL           *string
	User          *int64
	CustomEmojiID *int64
	Language      *string // pre: programming language of the code block
	Collapsed     bool    // blockquote: collapsed by default
}

type RawMessage struct {
	ID          int
	Text        *string
	Date        time.Time
	GroupID     int64
	Media       *Media
	Entities    []RawMessageEntity
	InvertMedia bool
}
