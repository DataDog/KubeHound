package tasks

import "github.com/spf13/cobra"

func podRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pod",
		Aliases: []string{"p"},
		Short:   "pod interpretation helpers",
	}

	// Add subcommands.
	cmd.AddCommand(
		podExecRootCmd(),
	)

	return cmd
}
