package cli

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"repin/internal/bootstrap"
	"repin/internal/context/application/usecase/regenseo"
	"repin/internal/context/application/usecase/rerender"
	"repin/internal/context/infrastructure/db/postgres/channel"
	"repin/internal/context/infrastructure/db/postgres/job"
	"repin/internal/context/infrastructure/db/postgres/post"
	clipres "repin/internal/context/presentation/cli"
	"repin/internal/pkg/config"
	"repin/internal/pkg/db"
	"repin/internal/pkg/logger"
	"repin/internal/pkg/migration"
	"repin/internal/pkg/validator"
)

type registry struct {
	cfg  *bootstrap.CLIConfig
	log  *zerolog.Logger
	db   *db.Client
	root *cobra.Command
}

func newRegistry(ctx context.Context) *registry {
	r := new(registry)
	if err := r.load(ctx); err != nil {
		r.cleanup()
		panic(err)
	}

	return r
}

func (r *registry) load(ctx context.Context) error {
	r.cfg = config.MustLoad(bootstrap.CLIConfig{})
	r.log = logger.MustLoad(r.cfg.Logger.Config())
	r.db = db.MustLoad(ctx, r.cfg.Database.Config())

	migrator := migration.MustLoad(r.db, r.cfg.Migration.Config())

	r.root = &cobra.Command{Use: "repin", Short: "repin.pw administration CLI"}
	r.root.AddCommand(migration.NewCLI(migrator, r.cfg.Migration.Dir))
	r.root.AddCommand(clipres.NewRerenderCommand(
		rerender.NewRerenderPosts(post.NewRepository(r.db), channel.NewRepository(r.db), db.NewTxRunner(r.db)),
	))
	r.root.AddCommand(clipres.NewSEOCommand(
		regenseo.NewRegenerateSEO(post.NewRepository(r.db), job.NewRepository(r.db)),
	))

	return validator.ValidateStructDependencies(r)
}

func (r *registry) cleanup() {
	if r.db != nil {
		r.db.Close()
	}
}
