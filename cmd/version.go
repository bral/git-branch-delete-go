package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	CommitSHA = "none"
	BuildTime = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long: `Print the version number of git-branch-delete.
Shows the version, commit hash, build date, and Go version used to build the binary.`,
	Example: `  git-branch-delete version
  git-branch-delete version --short`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Version: %s\nCommit: %s\nBuilt: %s\n", Version, CommitSHA, BuildTime)
	},
}

func runVersion(_ *cobra.Command, _ []string) {
	fmt.Printf("Version: %s\nCommit: %s\nBuilt: %s\n", Version, CommitSHA, BuildTime)
}
