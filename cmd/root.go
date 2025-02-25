package cmd

import (
	"github.com/bral/git-branch-delete-go/internal/config"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	cfg       *config.Config
	quietFlag bool
	debugFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "git-branch-delete",
	Short: "A tool for managing Git branches",
	Long: `git-branch-delete is a CLI tool for managing Git branches.
It provides features for listing, deleting, and pruning branches.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if quietFlag {
			log.SetQuiet(true)
		} else if debugFlag {
			log.SetDebug(true)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/git-branch-delete.yaml)")
	rootCmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "suppress all output except errors")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug output")
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatal("Error loading config: %v", err)
	}
}
