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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nCommit: %s\nBuilt: %s\n", Version, CommitSHA, BuildTime)
	},
}
