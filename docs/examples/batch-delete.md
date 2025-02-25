# Batch Branch Deletion Example

This example demonstrates how to safely delete multiple branches in batch using the git-branch-delete tool.

## Using the Interactive Mode

The simplest way to delete multiple branches is using interactive mode:

```bash
# Start interactive mode
git-branch-delete interactive

# Include remote branches
git-branch-delete interactive --all

# Only show merged branches
git-branch-delete interactive --merged
```

## Using the Command Line

For scripting or automation, you can use direct commands:

```bash
# List and delete all merged feature branches
git-branch-delete list --merged | grep 'feature/' | xargs git-branch-delete delete

# Delete all stale remote branches
git-branch-delete list --stale | xargs git-branch-delete delete --remote

# Force delete all branches matching a pattern
git-branch-delete list | grep 'feature/old-' | xargs git-branch-delete delete --force
```

## Using the Go API

For programmatic usage, you can use the Go API:

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/bral/git-branch-delete-go/internal/git"
)

func main() {
    // Create a git instance
    g, err := git.New(".")
    if err != nil {
        log.Fatal(err)
    }

    // List all branches
    branches, err := g.ListBranches()
    if err != nil {
        log.Fatal(err)
    }

    // Create a batch processor
    processor := git.NewBatchProcessor(g)

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Process branches in parallel
    err = processor.ProcessBranches(ctx, branches, func(b git.GitBranch) error {
        // Skip protected branches
        if b.IsDefault {
            return nil
        }

        // Skip unmerged branches
        if !b.IsMerged {
            return nil
        }

        // Delete the branch
        return g.DeleteBranch(b.Name, false, false)
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Safety Considerations

1. Always verify branches are merged before deletion
2. Use `--dry-run` for testing deletions
3. Protect default branches in configuration
4. Handle authentication for remote operations
5. Consider using timeouts for batch operations

## Error Handling

```bash
# Use --force with caution
git-branch-delete delete feature/123 --force

# Check exit codes in scripts
if git-branch-delete delete feature/123; then
    echo "Branch deleted successfully"
else
    echo "Failed to delete branch"
fi

# Use --quiet for scripting
git-branch-delete --quiet delete feature/123 || exit 1
```
