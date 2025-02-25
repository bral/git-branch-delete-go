# Branch Listing Examples

This example demonstrates various ways to list and filter Git branches using git-branch-delete.

## Basic Listing

```bash
# List all local branches
git-branch-delete list

# List all branches including remotes
git-branch-delete list --all

# List only merged branches
git-branch-delete list --merged

# List only stale branches
git-branch-delete list --stale
```

## Filtering and Formatting

```bash
# List branches with custom format
git-branch-delete list --format "{{.Name}} ({{.CommitHash}})"

# List branches with full details
git-branch-delete list --format json

# Filter by pattern
git-branch-delete list | grep 'feature/'

# List branches older than 30 days
git-branch-delete list --stale --days 30
```

## Using the Go API

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"

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

    // Print branches as JSON
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    if err := enc.Encode(branches); err != nil {
        log.Fatal(err)
    }

    // Filter and print specific branches
    for _, branch := range branches {
        if branch.IsMerged && !branch.IsDefault {
            fmt.Printf("Merged branch: %s (%s)\n", branch.Name, branch.CommitHash)
        }
    }
}
```

## Common Use Cases

### Finding Stale Branches

```bash
# List branches not updated in last 90 days
git-branch-delete list --stale --days 90

# List merged branches that haven't been updated
git-branch-delete list --merged --stale

# List remote branches that have been deleted upstream
git-branch-delete list --gone
```

### Scripting Examples

```bash
# Count total branches
git-branch-delete list | wc -l

# Find branches by author
git-branch-delete list --format "{{.Name}} {{.Author}}" | grep "john.doe"

# Export branch list to file
git-branch-delete list --format json > branches.json
```

### Integration with Other Tools

```bash
# Use with fzf for interactive filtering
git-branch-delete list | fzf --multi | xargs git-branch-delete delete

# Use with jq for JSON processing
git-branch-delete list --format json | jq '.[] | select(.merged==true) | .name'

# Use in CI/CD pipelines
git-branch-delete list --merged --format json | jq -r '.[] | .name' > merged-branches.txt
```
