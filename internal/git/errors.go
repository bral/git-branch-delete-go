package git

import (
	"fmt"
)

// Custom error types for better error handling
type (
	// ErrInvalidBranch indicates an invalid branch name or operation
	ErrInvalidBranch struct {
		Name   string
		Reason string
	}

	// ErrProtectedBranch indicates an operation on a protected branch
	ErrProtectedBranch struct {
		Name string
	}

	// ErrUnmergedBranch indicates an operation on an unmerged branch
	ErrUnmergedBranch struct {
		Name string
	}

	// ErrGitCommand indicates a git command failure
	ErrGitCommand struct {
		Command string
		Output  string
		Err     error
	}

	// ErrTimeout indicates a git command timeout
	ErrTimeout struct {
		Command string
		Timeout string
	}
)

// Error implementations
func (e *ErrInvalidBranch) Error() string {
	return fmt.Sprintf("invalid branch '%s': %s", e.Name, e.Reason)
}

func (e *ErrProtectedBranch) Error() string {
	return fmt.Sprintf("cannot modify protected branch '%s'", e.Name)
}

func (e *ErrUnmergedBranch) Error() string {
	return fmt.Sprintf("branch '%s' is not fully merged", e.Name)
}

func (e *ErrGitCommand) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("git command '%s' failed: %s\nOutput: %s", e.Command, e.Err, e.Output)
	}
	return fmt.Sprintf("git command '%s' failed: %s", e.Command, e.Err)
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("git command '%s' timed out after %s", e.Command, e.Timeout)
}

// Helper functions to create errors
func newInvalidBranchError(name, reason string) error {
	return &ErrInvalidBranch{Name: name, Reason: reason}
}

func newProtectedBranchError(name string) error {
	return &ErrProtectedBranch{Name: name}
}

func newUnmergedBranchError(name string) error {
	return &ErrUnmergedBranch{Name: name}
}

func newGitCommandError(cmd, output string, err error) error {
	return fmt.Errorf("git command '%s' failed: %s: %w", cmd, output, err)
}

func newTimeoutError(cmd, timeout string) error {
	return fmt.Errorf("git command '%s' timed out after %s", cmd, timeout)
}
