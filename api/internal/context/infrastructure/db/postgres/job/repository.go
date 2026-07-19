package job

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"repin/internal/context/domain"
	"repin/internal/pkg/db"
)

const defaultMaxAttempts = 5

type Repository struct {
	db    *db.Client
	table string
}

func NewRepository(client *db.Client) *Repository {
	return &Repository{db: client, table: client.Qualify("jobs")}
}

func (r *Repository) Enqueue(ctx context.Context, kind domain.JobKind, dedupKey string, payload any) error {
	ext := db.Executor(ctx, r.db.Connection())

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode payload of %s job: %w", kind, err)
	}

	query, args, err := sqlx.Named(fmt.Sprintf(enqueueQuery, r.table), map[string]any{
		"kind":         string(kind),
		"dedup_key":    dedupKey,
		"payload":      string(encoded),
		"max_attempts": defaultMaxAttempts,
	})
	if err != nil {
		return fmt.Errorf("bind enqueue job: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("enqueue %s job: %w", kind, err)
	}

	return nil
}

func (r *Repository) Claim(ctx context.Context, kinds []domain.JobKind) (*domain.Job, error) {
	ext := db.Executor(ctx, r.db.Connection())

	names := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		names = append(names, string(kind))
	}

	query, args, err := sqlx.Named(
		fmt.Sprintf(claimQuery, r.table, r.table),
		map[string]any{"kinds": names},
	)
	if err != nil {
		return nil, fmt.Errorf("bind claim job: %w", err)
	}

	var row job
	if err := ext.QueryRowxContext(ctx, ext.Rebind(query), args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNoJobs
		}

		return nil, fmt.Errorf("claim job: %w", err)
	}

	return row.ToDomain(), nil
}

func (r *Repository) Complete(ctx context.Context, id int64) error {
	return r.exec(ctx, fmt.Sprintf(completeQuery, r.table), map[string]any{"id": id}, "complete job")
}

func (r *Repository) Retry(ctx context.Context, id int64, runAt time.Time, cause string) error {
	err := r.exec(ctx, fmt.Sprintf(retryQuery, r.table, r.table), map[string]any{
		"id":         id,
		"run_at":     runAt,
		"last_error": cause,
	}, "retry job")
	if err != nil {
		return err
	}

	return r.exec(ctx, fmt.Sprintf(dropSupersededQuery, r.table), map[string]any{"id": id}, "drop superseded job")
}

func (r *Repository) Bury(ctx context.Context, id int64, cause string) error {
	return r.exec(ctx, fmt.Sprintf(buryQuery, r.table), map[string]any{"id": id, "last_error": cause}, "bury job")
}

func (r *Repository) RequeueStale(ctx context.Context, lease time.Duration) (int64, error) {
	ext := db.Executor(ctx, r.db.Connection())
	params := map[string]any{"threshold": time.Now().Add(-lease)}

	query, args, err := sqlx.Named(fmt.Sprintf(requeueStaleQuery, r.table, r.table, r.table), params)
	if err != nil {
		return 0, fmt.Errorf("bind requeue stale jobs: %w", err)
	}

	res, err := ext.ExecContext(ctx, ext.Rebind(query), args...)
	if err != nil {
		return 0, fmt.Errorf("requeue stale jobs: %w", err)
	}

	reclaimed, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if err := r.exec(ctx, fmt.Sprintf(dropStaleSupersededQuery, r.table), params, "drop superseded stale jobs"); err != nil {
		return reclaimed, err
	}

	return reclaimed, nil
}

func (r *Repository) exec(ctx context.Context, stmt string, params map[string]any, action string) error {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(stmt, params)
	if err != nil {
		return fmt.Errorf("bind %s: %w", action, err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}

	return nil
}
