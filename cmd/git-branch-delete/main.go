package main

import (
	"os"
	"path/filepath"

	"github.com/bral/git-branch-delete-go/internal/ui"
	"github.com/bral/git-branch-delete-go/pkg/git"

	"github.com/fatih/color"
)

func findGitRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
			return dir, true
		}
		dir = filepath.Dir(dir)
	}
	return "", false
}

func main() {
	gitDir, isGitRepo := findGitRoot()
	if !isGitRepo {
		color.Blue("Not a Git repository. Please navigate to a directory with a .git folder.")
		os.Exit(1)
	}

	gitClient := git.New(gitDir)
	branches, err := gitClient.ListBranches()
	if err != nil {
		color.Red("Error getting branches: %v", err)
		os.Exit(1)
	}

	if len(branches) == 0 {
		color.Blue("No branches found.")
		os.Exit(0)
	}

	// Check if there's only the current branch
	if len(branches) == 1 && branches[0].IsCurrent {
		color.Blue("Only one branch (current) exists. Nothing to do.")
		os.Exit(0)
	}

	// Select branches to delete
	selectedBranches, err := ui.SelectBranches(branches)
	if err != nil {
		color.Red("Error selecting branches: %v", err)
		os.Exit(1)
	}

	if len(selectedBranches) == 0 {
		color.Blue("No branches selected for deletion.")
		os.Exit(0)
	}

	// Confirm deletion
	confirmed, err := ui.ConfirmDeletion(selectedBranches)
	if err != nil {
		color.Red("Error during confirmation: %v", err)
		os.Exit(1)
	}

	if !confirmed {
		color.Blue("Operation cancelled. No branches deleted.")
		os.Exit(0)
	}

	// Delete branches and show results
	successCount := 0
	for _, branchName := range selectedBranches {
		err := gitClient.DeleteBranch(branchName, true, false)
		if err != nil {
			color.Red("✗ Failed to delete %s: %s", branchName, err)
		} else {
			color.Green("✓ Deleted branch %s", branchName)
			successCount++
		}
	}

	if successCount == len(selectedBranches) {
		color.Green("\nSuccessfully deleted all %d branch(es).", len(selectedBranches))
	} else {
		color.Yellow("\nDeleted %d out of %d branch(es).", successCount, len(selectedBranches))
	}
}
