# Git Package API Documentation

## Overview

The `git` package provides a high-level interface for managing Git branches in a repository. It focuses on safety and usability, with features for listing, deleting, and managing both local and remote branches.

## Core Types

### Git

The main type for interacting with a Git repository.

```go
type Git struct {
    // Contains unexported fields
}

// New creates a new Git instance for the given working directory
func New(workDir string) (*Git, error)
```

### GitBranch

Represents a Git branch and its metadata.

```go
type GitBranch struct {
    Name           string // Branch name
    CommitHash     string // Latest commit hash
    Reference      string // Full reference (e.g., refs/heads/main)
    IsCurrent      bool   // Is currently checked out
    IsRemote       bool   // Is a remote branch
    IsDefault      bool   // Is a default branch (main/master)
    IsMerged       bool   // Is merged into current branch
    IsStale        bool   // Is deleted from remote
    IsBehind       bool   // Is behind remote
    Message        string // Latest commit message
    TrackingBranch string // Remote tracking branch
}
```

## Core Functions

### Branch Management

```go
// List all branches in the repository
func (g *Git) ListBranches() ([]GitBranch, error)

// Delete a branch locally or remotely
func (g *Git) DeleteBranch(name string, force bool, remote bool) error

// Create a new branch
func (g *Git) CreateBranch(name string, createCommit bool) error

// Push a branch to remote
func (g *Git) PushBranch(name string) error

// Checkout a branch
func (g *Git) CheckoutBranch(name string) error
```

### Batch Operations

```go
// BatchProcessor handles concurrent branch operations
type BatchProcessor struct {
    // Contains unexported fields
}

// Process multiple branches concurrently
func (bp *BatchProcessor) ProcessBranches(ctx context.Context, branches []GitBranch, fn func(GitBranch) error) error
```

### Validation

```go
// Validate a git command argument
func ValidateGitArg(arg string) error

// Validate a branch name
func ValidateBranchName(name string) error

// Sanitize a branch name
func SanitizeBranchName(name string) string
```

## Error Types

The package provides custom error types for specific failure cases:

```go
type ErrInvalidBranch struct {
    Branch string
    Reason string
}

type ErrProtectedBranch struct {
    Branch string
}

type ErrUnmergedBranch struct {
    Branch string
}

type ErrGitCommand struct {
    Command string
    Output  string
    Err     error
}

type ErrTimeout struct {
    Command string
    Timeout string
}
```

## Usage Examples

### Basic Branch Management

```go
// Create a new Git instance
git, err := git.New(".")
if err != nil {
    log.Fatal(err)
}

// List all branches
branches, err := git.ListBranches()
if err != nil {
    log.Fatal(err)
}

// Delete a branch
err = git.DeleteBranch("feature/123", false, false)
if err != nil {
    log.Fatal(err)
}
```

### Batch Processing

```go
// Create a batch processor
processor := NewBatchProcessor(git)

// Process multiple branches
err := processor.ProcessBranches(ctx, branches, func(b GitBranch) error {
    if b.IsMerged {
        return git.DeleteBranch(b.Name, false, false)
    }
    return nil
})
```

### Error Handling

```go
err := git.DeleteBranch("main", false, false)
if err != nil {
    switch e := err.(type) {
    case *git.ErrProtectedBranch:
        fmt.Printf("Cannot delete protected branch: %s\n", e.Branch)
    case *git.ErrUnmergedBranch:
        fmt.Printf("Branch not fully merged: %s\n", e.Branch)
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Best Practices

1. Always validate branch names before operations:

   ```go
   if err := ValidateBranchName(name); err != nil {
       return err
   }
   ```

2. Use context for batch operations:

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

3. Handle authentication errors appropriately:

   ```go
   if err != nil && strings.Contains(err.Error(), "Authentication failed") {
       // Handle auth error
   }
   ```

4. Check for protected branches:
   ```go
   if isProtectedBranch(name) {
       return git.ErrProtectedBranch{Branch: name}
   }
   ```
