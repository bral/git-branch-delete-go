package cmd

import (
	"fmt"
	"os"

	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/spf13/cobra"
)

var (
	force  bool
	remote bool
	all    bool
)

func init() {
	deleteCmd := newDeleteCmd()
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete branches even if not merged")
	deleteCmd.Flags().BoolVarP(&remote, "remote", "r", false, "Delete remote branches")
	deleteCmd.Flags().BoolVarP(&all, "all", "a", false, "Delete both local and remote branches")
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [branches...]",
		Short: "Delete git branches",
		Long: `Delete one or more git branches locally and/or remotely.
Safely handles branch deletion with checks for unmerged changes.`,
		Example: `  git-branch-delete delete feature/123
  git-branch-delete delete -f old-branch
  git-branch-delete delete -r origin/feature/123
  git-branch-delete delete -a feature/123`,
		RunE: runDelete,
	}
}

func runDelete(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no branches specified")
	}

	log.Debug("Starting branch deletion", "branches", args)

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current directory", "error", err)
		return err
	}

	// Initialize git client
	gitClient, err := git.New(dir)
	if err != nil {
		log.Error("Failed to initialize git client", "error", err)
		return err
	}

	// Process each branch
	for _, branch := range args {
		log.Info("Processing branch", "branch", branch)

		if cfg.DryRun {
			log.Info("Would delete branch (dry run)", "branch", branch, "force", force, "remote", remote)
			continue
		}

		// Delete branch
		if err := gitClient.DeleteBranch(branch, force, remote); err != nil {
			log.Error("Failed to delete branch", "branch", branch, "error", err)
			return err
		}

		log.Info("Successfully deleted branch", "branch", branch)

		// If --all flag is set, also delete remote branch
		if all && !remote {
			log.Info("Deleting remote branch", "branch", branch)
			if err := gitClient.DeleteBranch(branch, force, true); err != nil {
				log.Error("Failed to delete remote branch", "branch", branch, "error", err)
				return err
			}
			log.Info("Successfully deleted remote branch", "branch", branch)
		}
	}

	return nil
}
