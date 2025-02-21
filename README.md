# Git Branch Delete (Go Version)

A command-line tool written in Go for interactively deleting Git branches. This is a Go implementation of the original TypeScript tool.

## Features

- Interactive branch selection with colored output
- Shows detailed branch information:
  - Commit hash
  - Last commit message
  - Merge status (green if merged, red if not)
- Safe deletion with confirmation
- Prevents deletion of current branch
- Vim-style navigation (Ctrl+N, Ctrl+P)

## Installation

To install globally:

```bash
go install github.com/brannonlucas/git-branch-delete-go/cmd/git-branch-delete@latest
```

Or build from source:

```bash
git clone https://github.com/brannonlucas/git-branch-delete-go.git
cd git-branch-delete-go
go build ./cmd/git-branch-delete
```

After building from source, you can:

1. Run it directly: `./git-branch-delete`
2. Move it to your PATH: `mv git-branch-delete /usr/local/bin/`

## Usage

1. Navigate to a Git repository
2. Run `git-branch-delete`
3. Use arrow keys or Ctrl+N/P to navigate branches
4. Press space to select branches for deletion
5. Press enter to confirm selection
6. Type 'y' to confirm deletion

## Key Bindings

- ↑/↓: Navigate through branches
- Ctrl+N: Next branch
- Ctrl+P: Previous branch
- Space: Select/deselect branch
- Enter: Confirm selection
- Ctrl+C: Exit without deleting

## Branch Display Format

Branches are displayed in the following format:

```
branch-name [a1b2c3d] Commit message... (merged)     # in green if merged
other-branch [e4f5g6h] WIP: New feature... (not merged)  # in red if not merged
```

- Branch name
- Commit hash in square brackets
- Truncated commit message (max 30 chars)
- Merge status in parentheses (colored)

## Requirements

- Go 1.21 or higher
- Git installed and accessible in PATH

## Dependencies

- github.com/fatih/color: Terminal color output
- github.com/manifoldco/promptui: Interactive terminal prompts
- golang.org/x/term: Terminal utilities
