package git

import "fmt"

// ErrBranchNotFound indicates the branch doesn't exist
type ErrBranchNotFound struct {
	Branch string
}

func (e *ErrBranchNotFound) Error() string {
	return fmt.Sprintf("branch not found: %s", e.Branch)
}

// ErrProtectedBranch indicates attempt to delete a protected branch
type ErrProtectedBranch struct {
	Branch string
}

func (e *ErrProtectedBranch) Error() string {
	return fmt.Sprintf("cannot delete protected branch: %s", e.Branch)
}

// ErrCurrentBranch indicates attempt to delete the current branch
type ErrCurrentBranch struct {
	Branch string
}

func (e *ErrCurrentBranch) Error() string {
	return fmt.Sprintf("cannot delete current branch: %s", e.Branch)
}

// ErrUnmergedBranch indicates attempt to delete an unmerged branch without force
type ErrUnmergedBranch struct {
	Branch string
}

func (e *ErrUnmergedBranch) Error() string {
	return fmt.Sprintf("branch has unmerged changes: %s", e.Branch)
}

// ErrNotGitRepo indicates the directory is not a git repository
type ErrNotGitRepo struct {
	Dir string
}

func (e *ErrNotGitRepo) Error() string {
	return fmt.Sprintf("directory %s is not a git repository", e.Dir)
}

// ErrInvalidBranch indicates that the branch name is invalid
type ErrInvalidBranch struct {
	Branch string
}

func (e *ErrInvalidBranch) Error() string {
	return fmt.Sprintf("invalid branch name: %s", e.Branch)
}

// ErrGitCommand indicates that a git command failed
type ErrGitCommand struct {
	Cmd    string
	Output string
}

func (e *ErrGitCommand) Error() string {
	return fmt.Sprintf("git command failed: %s\nOutput: %s", e.Cmd, e.Output)
}

// ErrTimeout indicates that a git operation timed out
type ErrTimeout struct {
	Operation string
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("operation timed out: %s", e.Operation)
}
