package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/spf13/cobra"
)

var (
	pruneForce bool
)

func init() {
	pruneCmd := newPruneCmd()
	rootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "Force delete branches without confirmation")
}

func newPruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Delete stale branches",
		Long: `Delete branches that have been merged or deleted from remote.
By default, asks for confirmation before deleting.`,
		Example: `  git-branch-delete prune
  git-branch-delete prune --force`,
		RunE: runPrune,
	}
}

func runPrune(cmd *cobra.Command, args []string) error {
	log.Debug("Starting branch pruning")

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

	// Get branches
	branches, err := gitClient.ListBranches()
	if err != nil {
		log.Error("Failed to list branches", "error", err)
		return err
	}

	log.Debug("Retrieved branches", "count", len(branches))

	// Filter stale branches
	var staleBranches []git.GitBranch
	for _, branch := range branches {
		if branch.IsStale && !branch.IsDefault && !branch.IsCurrent {
			staleBranches = append(staleBranches, branch)
		}
	}

	log.Debug("Found stale branches", "count", len(staleBranches))

	if len(staleBranches) == 0 {
		log.Info("No stale branches found")
		return nil
	}

	// If not force mode, confirm deletion
	if !pruneForce {
		var selectedBranches []string
		prompt := &survey.MultiSelect{
			Message: "Select branches to delete:",
			Options: func() []string {
				options := make([]string, len(staleBranches))
				for i, b := range staleBranches {
					options[i] = fmt.Sprintf("%s (%s)", b.Name, b.CommitHash)
				}
				return options
			}(),
		}

		if err := survey.AskOne(prompt, &selectedBranches); err != nil {
			log.Error("Failed to get user input", "error", err)
			return err
		}

		if len(selectedBranches) == 0 {
			log.Info("No branches selected for deletion")
			return nil
		}

		// Map selected options back to branch names
		staleBranches = func() []git.GitBranch {
			selected := make([]git.GitBranch, 0, len(selectedBranches))
			for _, opt := range selectedBranches {
				for _, b := range staleBranches {
					if fmt.Sprintf("%s (%s)", b.Name, b.CommitHash) == opt {
						selected = append(selected, b)
						break
					}
				}
			}
			return selected
		}()
	}

	// Delete selected branches
	for _, branch := range staleBranches {
		log.Info("Deleting branch", "branch", branch.Name)

		if err := gitClient.DeleteBranch(branch.Name, true, false); err != nil {
			log.Error("Failed to delete branch", "branch", branch.Name, "error", err)
			return err
		}

		log.Info("Successfully deleted branch", "branch", branch.Name)
	}

	log.Info("Branch pruning completed", "deleted", len(staleBranches))
	return nil
}
