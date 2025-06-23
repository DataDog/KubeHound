package tasks

import "github.com/spf13/cobra"

func permissionRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "permission",
		Aliases: []string{"p"},
		Short:   "permission interpretation helpers",
	}

	// Add subcommands.
	cmd.AddCommand(
		permissionExecPodsCmd(),
		permissionExecGroupsCmd(),
		permissionReachableNamespacesCmd(),
	)

	return cmd
}
