package main

import (
	"os"
	"path/filepath"

	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/ui"

	"github.com/fatih/color"
)

func findGitRoot() bool {
	dir, err := os.Getwd()
	if err != nil {
		return false
	}

	for dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
			return true
		}
		dir = filepath.Dir(dir)
	}
	return false
}

func main() {
	if !findGitRoot() {
		color.Blue("Not a Git repository. Please navigate to a directory with a .git folder.")
		os.Exit(1)
	}

	branches, err := git.GetBranches()
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
	results := git.DeleteBranches(selectedBranches)

	successCount := 0
	for _, result := range results {
		if result.Success {
			color.Green("✓ Deleted branch %s", result.Name)
			successCount++
		} else {
			color.Red("✗ Failed to delete %s: %s", result.Name, result.Error)
		}
	}

	if successCount == len(results) {
		color.Green("\nSuccessfully deleted all %d branch(es).", len(results))
	} else {
		color.Yellow("\nDeleted %d out of %d branch(es).", successCount, len(results))
	}
}
