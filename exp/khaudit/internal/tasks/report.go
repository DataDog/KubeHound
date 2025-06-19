package tasks

import "github.com/spf13/cobra"

func reportRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "report",
		Aliases: []string{"r"},
		Short:   "report generation helpers",
	}

	// Add subcommands.
	cmd.AddCommand(
		reportMarkdownRootCmd(),
	)

	return cmd
}
