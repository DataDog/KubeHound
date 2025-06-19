package tasks

import "github.com/spf13/cobra"

func containerEscapeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "container-escape",
		Aliases: []string{"ce"},
		Short:   "container escape enumeration",
	}

	// Add subcommands.
	cmd.AddCommand(
		containerEscapeProfilesCmd(),
		containerEscapeContainersCmd(),
		containerEscapePathsCmd(),
	)

	return cmd
}
