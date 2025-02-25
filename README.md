# git-branch-delete

A powerful CLI tool for managing Git branches with features for safe deletion, interactive selection, and branch cleanup.

## Features

- 🔍 List local and remote branches with detailed status
- 🗑️ Safely delete branches with protection for default branches
- 🤝 Interactive mode for selecting multiple branches
- 🧹 Prune stale and merged branches
- 🎨 Color-coded output for better visibility
- 🔒 Protected branches configuration
- 🔄 Remote branch handling
- 🚦 Dry-run mode for safety

## Installation

### Using Go

```bash
go install github.com/bral/git-branch-delete-go@latest
```

### Using Homebrew

```bash
brew install brannonlucas/tap/git-branch-delete
```

### From Release

Download the latest release from the [releases page](https://github.com/bral/git-branch-delete-go/releases).

## Usage

### List Branches

```bash
# List local branches
git-branch-delete list

# List remote branches
git-branch-delete list --remote

# List all branches
git-branch-delete list --all
```

### Interactive Mode

```bash
# Select branches to delete interactively
git-branch-delete interactive
# or use the shorthand
git-branch-delete i
```

### Prune Stale Branches

```bash
# Show stale branches
git-branch-delete prune --dry-run

# Delete stale branches (with confirmation)
git-branch-delete prune

# Force delete stale branches
git-branch-delete prune --force
```

### Configuration

Create `~/.config/git-branch-delete.yaml`:

```yaml
# Override default branch detection
default_branch: main

# Protect specific branches from deletion
protected_branches:
  - main
  - master
  - develop

# Default remote (default: origin)
default_remote: origin

# Skip confirmation prompts
auto_confirm: false

# Show what would be deleted without actually deleting
dry_run: false
```

Environment variables are also supported:

- `GBD_DEFAULT_BRANCH`
- `GBD_DEFAULT_REMOTE`
- `GBD_AUTO_CONFIRM`
- `GBD_DRY_RUN`

### Shell Completion

```bash
# Bash
source <(git-branch-delete completion bash)

# Zsh
source <(git-branch-delete completion zsh)

# Fish
git-branch-delete completion fish | source

# PowerShell
git-branch-delete completion powershell | Out-String | Invoke-Expression
```

## Development

### Requirements

- Go 1.22 or later
- Make

### Setup

```bash
# Clone the repository
git clone https://github.com/bral/git-branch-delete-go.git
cd git-branch-delete-go

# Install dependencies
make deps

# Build
make build

# Run tests
make test
```

### Commands

- `make build` - Build the binary
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make install` - Install to $GOPATH/bin
- `make deps` - Install dependencies

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file
