package cmd

import (
	"github.com/spf13/cobra"
)

// initTestRoot initializes a new root command for testing
func initTestRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-branch-delete",
		Short: "A tool for managing git branches",
		Long: `A command line tool for managing git branches.
Provides functionality to list, delete, and prune branches.`,
	}

	// Add global flags
	cmd.PersistentFlags().String("config", "", "config file (default is $HOME/.config/git-branch-delete.yaml)")
	cmd.PersistentFlags().Bool("debug", false, "enable debug output")
	cmd.PersistentFlags().Bool("quiet", false, "suppress all output except errors")

	return cmd
}
