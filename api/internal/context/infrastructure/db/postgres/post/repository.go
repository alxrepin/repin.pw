package post

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
	return &Repository{db: client, table: client.Qualify("posts")}
}

func (r *Repository) List(ctx context.Context, page, limit int) ([]domain.Post, int, error) {
	ext := db.Executor(ctx, r.db.Connection())

	var total int
	if err := ext.QueryRowxContext(ctx, fmt.Sprintf(countQuery, r.table)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	query, args, err := sqlx.Named(fmt.Sprintf(listQuery, r.table), map[string]any{
		"limit":  limit,
		"offset": (page - 1) * limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("bind list posts: %w", err)
	}

	rows, err := ext.QueryxContext(ctx, ext.Rebind(query), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query posts: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post

	for rows.Next() {
		var row post
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, fmt.Errorf("scan post: %w", err)
		}

		p, err := row.ToDomain()
		if err != nil {
			return nil, 0, err
		}

		posts = append(posts, *p)
	}

	return posts, total, rows.Err()
}

func (r *Repository) All(ctx context.Context) ([]domain.Post, error) {
	ext := db.Executor(ctx, r.db.Connection())

	rows, err := ext.QueryxContext(ctx, fmt.Sprintf(allQuery, r.table))
	if err != nil {
		return nil, fmt.Errorf("query all posts: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post

	for rows.Next() {
		var row post
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}

		p, err := row.ToDomain()
		if err != nil {
			return nil, err
		}

		posts = append(posts, *p)
	}

	return posts, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(getByIDQuery, r.table), map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("bind get post: %w", err)
	}

	var row post
	if err := ext.QueryRowxContext(ctx, ext.Rebind(query), args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}

		return nil, fmt.Errorf("get post: %w", err)
	}

	return row.ToDomain()
}

// UpdateSEO writes only the search-metadata columns of a post.
func (r *Repository) UpdateSEO(ctx context.Context, id int64, seo domain.PostSEO) error {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(updateSEOQuery, r.table), map[string]any{
		"id":              id,
		"seo_title":       seo.Title,
		"seo_description": seo.Description,
		"seo_keywords":    seo.Keywords,
	})
	if err != nil {
		return fmt.Errorf("bind update post seo: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("update post seo: %w", err)
	}

	return nil
}

func (r *Repository) GetByURL(ctx context.Context, url string) (*domain.Post, error) {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(getByURLQuery, r.table), map[string]any{"url": url})
	if err != nil {
		return nil, fmt.Errorf("bind get post: %w", err)
	}

	var row post
	if err := ext.QueryRowxContext(ctx, ext.Rebind(query), args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}

		return nil, fmt.Errorf("get post: %w", err)
	}

	return row.ToDomain()
}

func (r *Repository) Prev(ctx context.Context, id int64) (*domain.Post, error) {
	return r.adjacent(ctx, prevQuery, id)
}

func (r *Repository) Next(ctx context.Context, id int64) (*domain.Post, error) {
	return r.adjacent(ctx, nextQuery, id)
}

func (r *Repository) adjacent(ctx context.Context, tmpl string, id int64) (*domain.Post, error) {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(tmpl, r.table), map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("bind adjacent post: %w", err)
	}

	var row post
	if err := ext.QueryRowxContext(ctx, ext.Rebind(query), args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}

		return nil, fmt.Errorf("adjacent post: %w", err)
	}

	return row.ToDomain()
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	ext := db.Executor(ctx, r.db.Connection())

	query, args, err := sqlx.Named(fmt.Sprintf(deleteQuery, r.table), map[string]any{"id": id})
	if err != nil {
		return fmt.Errorf("bind delete post: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	return nil
}

func (r *Repository) Upsert(ctx context.Context, p *domain.Post) error {
	ext := db.Executor(ctx, r.db.Connection())

	row, err := fromDomain(p)
	if err != nil {
		return err
	}

	query, args, err := sqlx.Named(fmt.Sprintf(upsertQuery, r.table), row)
	if err != nil {
		return fmt.Errorf("bind upsert post: %w", err)
	}

	if _, err := ext.ExecContext(ctx, ext.Rebind(query), args...); err != nil {
		return fmt.Errorf("upsert post: %w", err)
	}

	return nil
}
