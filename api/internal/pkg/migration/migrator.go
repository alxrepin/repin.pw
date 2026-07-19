package migration

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // file:// migration source

	"repin/internal/pkg/db"
)

type Config struct {
	Dir   string `env:"MIGRATIONS_DIR" envDefault:"migrations"`
	Table string `env:"MIGRATIONS_TABLE" envDefault:"schema_migrations"`
}

type Migrator struct {
	*migrate.Migrate
}

func Load(client *db.Client, cfg Config) (*Migrator, error) {
	driver, err := migratepg.WithInstance(client.DB.DB, &migratepg.Config{
		SchemaName:      client.Schema(),
		MigrationsTable: cfg.Table,
	})
	if err != nil {
		return nil, fmt.Errorf("migration driver: %w", err)
	}

	path, err := filepath.Abs(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("migration dir: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("migration instance: %w", err)
	}

	return &Migrator{Migrate: m}, nil
}

func MustLoad(client *db.Client, cfg Config) *Migrator {
	m, err := Load(client, cfg)
	if err != nil {
		panic(err)
	}

	return m
}
