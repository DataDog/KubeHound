package tasks

import "github.com/spf13/cobra"

func pathRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "path",
		Aliases: []string{"p"},
		Short:   "path interpretation helpers",
	}

	// Add subcommands.
	cmd.AddCommand(
		pathJsonifyCmd(),
		pathAttackFlowCmd(),
		pathFilterCmd(),
	)

	return cmd
}
