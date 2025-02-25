package cmd

import (
	"github.com/bral/git-branch-delete-go/internal/config"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config

	// Global flags
	quietFlag  bool
	debugFlag  bool
	dryRunFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "git-branch-delete",
	Short: "A tool to manage git branch deletion",
	Long: `A CLI tool that helps manage and delete Git branches
safely and efficiently across your repositories.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set up logging first
		log.SetQuiet(quietFlag)
		log.SetDebug(debugFlag)

		// Load config
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		// Override config with flags
		if dryRunFlag {
			cfg.DryRun = true
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "suppress info messages")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "show what would be done without doing it")
}
