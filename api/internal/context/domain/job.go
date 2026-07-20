package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type JobKind string

const (
	JobKindMediaDownload JobKind = "media.download"
	JobKindGenerateSEO   JobKind = "post.seo"
)

// Job is always a claimed one: Claim is the only thing that builds it, and the
// same statement flips the row to "running". The queue's states live in SQL —
// the partial indexes on jobs are defined in terms of them and every transition
// is an atomic UPDATE — so nothing here mirrors them.
type Job struct {
	ID          int64
	Kind        JobKind
	DedupKey    string
	Payload     json.RawMessage
	Attempts    int
	MaxAttempts int
	RunAt       time.Time
	LastError   *string
	CreatedAt   time.Time
}

func (j Job) Decode(out any) error {
	if err := json.Unmarshal(j.Payload, out); err != nil {
		return fmt.Errorf("decode payload of job %d (%s): %w", j.ID, j.Kind, err)
	}

	return nil
}

type MediaDownloadPayload struct {
	PostID    int64 `json:"post_id"`
	MessageID int64 `json:"message_id"`
}

func (p MediaDownloadPayload) DedupKey() string {
	return fmt.Sprintf("%s:%d", JobKindMediaDownload, p.MessageID)
}

type GenerateSEOPayload struct {
	PostID int64 `json:"post_id"`
}

func (p GenerateSEOPayload) DedupKey() string {
	return fmt.Sprintf("%s:%d", JobKindGenerateSEO, p.PostID)
}

type PostSEO struct {
	Title       string
	Description string
	Keywords    string
}
