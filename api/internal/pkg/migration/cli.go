package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
)

const migrationTemplate = "BEGIN;\n\nCOMMIT;\n"

func NewCLI(m *Migrator, dir string) *cobra.Command {
	root := &cobra.Command{
		Use:     "migrations",
		Aliases: []string{"m", "migration"},
		Short:   "Database migration commands",
	}

	root.AddCommand(upCommand(m), downCommand(m), createCommand(dir))

	return root
}

func upCommand(m *Migrator) *cobra.Command {
	var steps int

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply pending migrations (all, or --steps N)",
		RunE: func(_ *cobra.Command, _ []string) error {
			var err error
			if steps > 0 {
				err = m.Steps(steps)
			} else {
				err = m.Up()
			}

			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&steps, "steps", 0, "number of migrations to apply (0 = all)")

	return cmd
}

func downCommand(m *Migrator) *cobra.Command {
	var (
		steps int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Roll back migrations (--steps N, or --all)",
		RunE: func(_ *cobra.Command, _ []string) error {
			var err error
			switch {
			case all:
				err = m.Down()
			case steps > 0:
				err = m.Steps(-steps)
			default:
				return errors.New("specify --steps N or --all")
			}

			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&steps, "steps", 1, "number of migrations to roll back")
	cmd.Flags().BoolVar(&all, "all", false, "roll back all migrations")

	return cmd
}

func createCommand(dir string) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new empty migration pair",
		RunE: func(_ *cobra.Command, _ []string) error {
			if name == "" {
				return errors.New("--name is required")
			}

			version := strconv.FormatInt(time.Now().UnixNano(), 10)

			for _, direction := range []string{"up", "down"} {
				file := filepath.Join(dir, fmt.Sprintf("%s_%s.%s.sql", version, name, direction))
				if err := os.WriteFile(file, []byte(migrationTemplate), 0o600); err != nil {
					return err
				}

				fmt.Println("created", file)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "migration name (snake_case)")

	return cmd
}
