package cli

import (
	"github.com/spf13/cobra"

	"repin/internal/context/application/usecase/rerender"
)

func NewRerenderCommand(uc *rerender.RerenderPosts) *cobra.Command {
	return &cobra.Command{
		Use:   "rerender",
		Short: "Rebuild post title/text from the stored Telegram captions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			stats, err := uc.Execute(cmd.Context())
			if err != nil {
				return err
			}

			cmd.Printf("re-rendered %d posts, skipped %d without a raw caption\n", stats.Rendered, stats.Skipped)

			return nil
		},
	}
}
