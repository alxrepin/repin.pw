package cli

import (
	"github.com/spf13/cobra"

	"repin/internal/context/application/usecase/regenseo"
)

func NewSEOCommand(uc *regenseo.RegenerateSEO) *cobra.Command {
	root := &cobra.Command{Use: "seo", Short: "SEO metadata administration"}

	var missingOnly bool

	regenerate := &cobra.Command{
		Use:   "regenerate",
		Short: "Queue SEO metadata generation for posts (a running worker picks the jobs up)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			stats, err := uc.Execute(cmd.Context(), missingOnly)
			if err != nil {
				return err
			}

			cmd.Printf("queued %d posts, skipped %d\n", stats.Enqueued, stats.Skipped)

			return nil
		},
	}

	regenerate.Flags().BoolVar(&missingOnly, "missing-only", false,
		"only queue posts that have no SEO metadata yet")

	root.AddCommand(regenerate)

	return root
}
