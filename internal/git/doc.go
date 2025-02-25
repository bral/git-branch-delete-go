/*
Package git provides functionality for managing Git branches in a repository.

This package offers a high-level interface for common Git branch operations,
with a focus on safety and usability. It includes features for listing,
deleting, and managing both local and remote branches.

Basic usage:

	// Create a new Git instance for the current directory
	g := git.New(".")

	// List all branches
	branches, err := g.ListBranches()
	if err != nil {
		log.Fatal(err)
	}

	// Delete a branch
	err = g.DeleteBranch("feature/old-branch", true, false)
	if err != nil {
		log.Fatal(err)
	}

Branch Information:

The Branch struct provides detailed information about each branch:

	type Branch struct {
		Name       string // Branch name
		CommitHash string // Latest commit hash
		Message    string // Latest commit message
		IsLocal    bool   // Is a local branch
		IsRemote   bool   // Is a remote branch
		IsDefault  bool   // Is the default branch (main/master)
		IsCurrent  bool   // Is the currently checked out branch
		IsStale    bool   // Has been deleted from remote
		IsMerged   bool   // Has been merged to default branch
	}

Safety Features:

The package implements several safety measures:
  - Protection for default branches
  - Current branch deletion prevention
  - Stale branch detection
  - Merged branch tracking
  - Remote branch handling

Error Handling:

Custom error types are provided for common scenarios:
  - ErrBranchNotFound
  - ErrProtectedBranch
  - ErrCurrentBranch
  - ErrUnmergedBranch
  - ErrNotGitRepo

These can be used for specific error handling:

	err := g.DeleteBranch("main", false, false)
	if err != nil {
		switch e := err.(type) {
		case *git.ErrProtectedBranch:
			fmt.Printf("Cannot delete protected branch: %s\n", e.Branch)
		case *git.ErrCurrentBranch:
			fmt.Printf("Cannot delete current branch: %s\n", e.Branch)
		default:
			fmt.Printf("Error: %v\n", err)
		}
	}
*/
package git
