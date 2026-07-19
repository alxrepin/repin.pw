package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"repin/internal/context/domain"
	"repin/internal/pkg/db"
)

type Repository struct {
	db    *db.Client
	table string
}

func NewRepository(client *db.Client) *Repository {
	return &Repository{db: client, table: client.Qualify("media")}
}

func (r *Repository) GetByMessageID(ctx context.Context, id int64) (*domain.PostMedia, error) {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(getByMessageIDQuery, r.table), map[string]any{"tg_message_id": id})
	if err != nil {
		return nil, fmt.Errorf("bind get media: %w", err)
	}

	var row media
	if err := ext.QueryRowxContext(ctx, ext.Rebind(query), args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMediaNotFound
		}

		return nil, fmt.Errorf("get media: %w", err)
	}

	return row.ToDomain(), nil
}

func (r *Repository) ListByPostIDs(ctx context.Context, ids []int64) (map[int64][]domain.PostMedia, error) {
	result := make(map[int64][]domain.PostMedia)
	if len(ids) == 0 {
		return result, nil
	}

	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.In(fmt.Sprintf(listByPostIDsQuery, r.table), ids)
	if err != nil {
		return nil, fmt.Errorf("bind list media: %w", err)
	}

	rows, err := ext.QueryxContext(ctx, ext.Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("query media: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row media
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("scan media: %w", err)
		}

		result[row.PostID] = append(result[row.PostID], *row.ToDomain())
	}

	return result, rows.Err()
}

func (r *Repository) Upsert(ctx context.Context, m *domain.PostMedia) error {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(upsertQuery, r.table), fromDomain(m))
	if err != nil {
		return fmt.Errorf("bind upsert media: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("upsert media: %w", err)
	}

	return nil
}

func (r *Repository) DeleteByMessageID(ctx context.Context, id int64) error {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(deleteByMessageIDQuery, r.table), map[string]any{"tg_message_id": id})
	if err != nil {
		return fmt.Errorf("bind delete media: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}

	return nil
}

func (r *Repository) DeleteStale(ctx context.Context, postID int64, keepIDs []int64) error {
	ext := db.Executor(ctx, r.db.Connection())

	var (
		query string
		args  []any
		err   error
	)

	if len(keepIDs) == 0 {
		query, args, err = sqlx.Named(fmt.Sprintf(deleteByPostIDQuery, r.table), map[string]any{"post_id": postID})
	} else {
		query, args, err = sqlx.In(fmt.Sprintf(deleteStaleQuery, r.table), postID, keepIDs)
	}

	if err != nil {
		return fmt.Errorf("bind delete stale media: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("delete stale media: %w", err)
	}

	return nil
}
