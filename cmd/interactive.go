package cmd

import (
	"os"

	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/bral/git-branch-delete-go/internal/ui"
	"github.com/bral/git-branch-delete-go/pkg/git"
	"github.com/spf13/cobra"
)

var (
	interactiveForce bool
	interactiveAll  bool
)

func init() {
	interactiveCmd := newInteractiveCmd()
	rootCmd.AddCommand(interactiveCmd)

	interactiveCmd.Flags().BoolVarP(&interactiveForce, "force", "f", false, "Force delete branches without checking merge status")
	interactiveCmd.Flags().BoolVarP(&interactiveAll, "all", "a", false, "Delete both local and remote branches")
}

func newInteractiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive",
		Short: "Interactively select branches to delete",
		Long: `Interactively select branches to delete.
Shows a list of branches with their status and allows selecting multiple branches for deletion.`,
		Example: `  git-branch-delete interactive
  git-branch-delete interactive --force
  git-branch-delete interactive --all`,
		RunE: runInteractive,
	}
}

func runInteractive(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current directory: %v", err)
		return err
	}

	gitClient := git.New(dir)
	branches, err := gitClient.ListBranches()
	if err != nil {
		log.Error("Failed to list branches: %v", err)
		return err
	}

	selectedBranches, err := ui.SelectBranches(branches)
	if err != nil {
		log.Error("Failed to select branches: %v", err)
		return err
	}

	if len(selectedBranches) == 0 {
		log.Info("No branches selected for deletion")
		return nil
	}

	confirmed, err := ui.ConfirmDeletion(selectedBranches)
	if err != nil {
		log.Error("Failed to confirm deletion: %v", err)
		return err
	}

	if !confirmed {
		log.Info("Operation cancelled")
		return nil
	}

	successCount := 0
	totalOperations := len(selectedBranches)
	if interactiveAll {
		totalOperations *= 2
	}

	for _, branchName := range selectedBranches {
		err := gitClient.DeleteBranch(branchName, interactiveForce, false)
		if err != nil {
			log.Error("Failed to delete branch %s: %v", branchName, err)
		} else {
			log.Info("Successfully deleted branch: %s", branchName)
			successCount++
		}

		// If --all flag is set, also delete remote branch
		if interactiveAll {
			log.Info("Deleting remote branch: %s", branchName)
			if err := gitClient.DeleteBranch(branchName, interactiveForce, true); err != nil {
				log.Error("Failed to delete remote branch %s: %v", branchName, err)
			} else {
				log.Info("Successfully deleted remote branch: %s", branchName)
				successCount++
			}
		}
	}

	log.Info("Branch deletion completed: %d of %d branches deleted", successCount, totalOperations)
	return nil
}
