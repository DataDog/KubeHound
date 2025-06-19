package tasks

import "github.com/spf13/cobra"

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "khaudit",
		Short: "khaudit is a tool to audit KubeHound datasets.",
	}

	// Add subcommands.
	cmd.AddCommand(
		containerEscapeRootCmd(),
		ingestionRootCmd(),
		mcpCmd(),
		pathRootCmd(),
		permissionRootCmd(),
		csvRootCmd(),
		podRootCmd(),
		reportRootCmd(),
		volumeRootCmd(),
	)

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return rootCmd().Execute()
}
