package tasks

import "github.com/spf13/cobra"

func csvRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "CSV utilities",
	}

	// Add subcommands.
	cmd.AddCommand(
		csvMarkdownCmd(),
	)

	return cmd
}
