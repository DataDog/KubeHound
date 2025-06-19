package tasks

import "github.com/spf13/cobra"

func ingestionRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ingestion",
		Aliases: []string{"i"},
		Short:   "ingestion management",
	}

	// Add subcommands.
	cmd.AddCommand(
		ingestionListCmd(),
		ingestionStatsCmd(),
	)

	return cmd
}
