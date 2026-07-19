package channel

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
	return &Repository{db: client, table: client.Qualify("channels")}
}

func (r *Repository) Get(ctx context.Context) (*domain.Channel, error) {
	ext := db.Executor(ctx, r.db.Connection())

	var row channel
	if err := ext.QueryRowxContext(ctx, fmt.Sprintf(getQuery, r.table)).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrChannelNotFound
		}

		return nil, fmt.Errorf("get channel: %w", err)
	}

	return row.ToDomain(), nil
}

func (r *Repository) Upsert(ctx context.Context, c *domain.Channel) error {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(upsertQuery, r.table, r.table), fromDomain(c))
	if err != nil {
		return fmt.Errorf("bind upsert channel: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("upsert channel: %w", err)
	}

	return nil
}
