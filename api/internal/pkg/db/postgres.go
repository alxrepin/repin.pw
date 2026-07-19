package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // register the "pgx" database/sql driver
)

type Config struct {
	URL    string `env:"DATABASE_URL"`
	Schema string `env:"DATABASE_SCHEMA" envDefault:"public"`
}

type Client struct {
	*sqlx.DB

	schema string
}

func Load(ctx context.Context, cfg Config) (*Client, error) {
	pool, err := sqlx.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := pool.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if cfg.Schema != "" {
		stmt := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdent(cfg.Schema))
		if _, err := pool.ExecContext(ctx, stmt); err != nil {
			return nil, fmt.Errorf("ensure schema: %w", err)
		}
	}

	return &Client{DB: pool, schema: cfg.Schema}, nil
}

func MustLoad(ctx context.Context, cfg Config) *Client {
	c, err := Load(ctx, cfg)
	if err != nil {
		panic(err)
	}

	return c
}

func (c *Client) Connection() *sqlx.DB { return c.DB }

func (c *Client) Schema() string { return c.schema }

func (c *Client) Qualify(table string) string {
	if c.schema == "" {
		return quoteIdent(table)
	}

	return quoteIdent(c.schema) + "." + quoteIdent(table)
}

func (c *Client) Close() {
	if err := c.DB.Close(); err != nil {
		panic(fmt.Errorf("close postgres: %w", err))
	}
}

func quoteIdent(name string) string {
	return `"` + name + `"`
}
