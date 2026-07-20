package job

import (
	"encoding/json"
	"time"

	"repin/internal/context/domain"
)

type job struct {
	ID          int64     `db:"id"`
	Kind        string    `db:"kind"`
	DedupKey    string    `db:"dedup_key"`
	Payload     []byte    `db:"payload"`
	Attempts    int       `db:"attempts"`
	MaxAttempts int       `db:"max_attempts"`
	RunAt       time.Time `db:"run_at"`
	LastError   *string   `db:"last_error"`
	CreatedAt   time.Time `db:"created_at"`
}

func (j job) ToDomain() *domain.Job {
	return &domain.Job{
		ID:          j.ID,
		Kind:        domain.JobKind(j.Kind),
		DedupKey:    j.DedupKey,
		Payload:     json.RawMessage(j.Payload),
		Attempts:    j.Attempts,
		MaxAttempts: j.MaxAttempts,
		RunAt:       j.RunAt,
		LastError:   j.LastError,
		CreatedAt:   j.CreatedAt,
	}
}
