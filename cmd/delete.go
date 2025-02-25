package cmd

import (
	"fmt"
	"os"

	"github.com/bral/git-branch-delete-go/internal/git"
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
		return fmt.Errorf("at least one branch name is required")
	}

	if all {
		remote = true
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	g, err := git.New(wd)
	if err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	branches, err := g.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	for _, branchToDelete := range args {
		var found bool
		for _, branch := range branches {
			if branch.Name == branchToDelete ||
			   (remote && "origin/"+branchToDelete == branch.Name) {
				found = true

				if branch.IsDefault {
					return fmt.Errorf("cannot delete default branch: %s", branch.Name)
				}

				if branch.IsCurrent {
					return fmt.Errorf("cannot delete current branch: %s", branch.Name)
				}

				if err := g.DeleteBranch(branch.Name, force, remote); err != nil {
					return fmt.Errorf("failed to delete branch %s: %w", branch.Name, err)
				}

				fmt.Printf("Successfully deleted branch: %s\n", branch.Name)
			}
		}

		if !found {
			return fmt.Errorf("branch not found: %s", branchToDelete)
		}
	}

	return nil
}
