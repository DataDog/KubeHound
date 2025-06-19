package tasks

import "github.com/spf13/cobra"

func volumeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume",
		Aliases: []string{"v"},
		Short:   "volume helpers",
	}

	// Add subcommands.
	cmd.AddCommand(
		volumeListCmd(),
		volumeHostpathCmd(),
	)

	return cmd
}
