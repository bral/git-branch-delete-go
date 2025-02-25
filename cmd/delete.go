package cmd

import (
	"fmt"
	"os"

	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/bral/git-branch-delete-go/pkg/git"
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
		return fmt.Errorf("branch name required")
	}

	branchName := args[0]

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current directory: %v", err)
		return err
	}

	// Initialize git client
	gitClient := git.New(dir)

	// Check if branch is protected
	for _, protected := range cfg.ProtectedBranches {
		if branchName == protected {
			return fmt.Errorf("cannot delete protected branch: %s", branchName)
		}
	}

	// Delete the branch
	if err := gitClient.DeleteBranch(branchName, force, remote); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	log.Info("Successfully deleted branch: %s", branchName)

	// If --all flag is set, also delete remote branch
	if all && !remote {
		log.Info("Deleting remote branch: %s", branchName)
		if err := gitClient.DeleteBranch(branchName, force, true); err != nil {
			return fmt.Errorf("failed to delete remote branch: %w", err)
		}
		log.Info("Successfully deleted remote branch: %s", branchName)
	}

	return nil
}
