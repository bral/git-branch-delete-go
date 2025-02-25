# git-branch-delete

A powerful CLI tool for managing Git branches with features for safe deletion, interactive selection, and branch cleanup.

## Features

- 🔍 List local and remote branches with detailed status (commit hash, merge status)
- 🗑️ Safely delete branches with protection for default branches (main, master, develop)
- 🤝 Interactive mode with multi-select and keyboard navigation
- 🧹 Prune stale and merged branches
- 🎨 Color-coded output for better visibility
- 🔒 Protected branches configuration
- 🔄 Remote branch handling with authentication support
- 🚦 Dry-run mode for safety

## Installation

### Using Go

```bash
go install github.com/bral/git-branch-delete-go@latest
```

### From Release

Download the latest release from the [releases page](https://github.com/bral/git-branch-delete-go/releases).

## Usage

### Basic Commands

```bash
# List all branches with status
git-branch-delete list --all

# Delete a single branch
git-branch-delete delete feature/old-branch

# Delete a remote branch
git-branch-delete delete feature/old-branch --remote

# Force delete an unmerged branch
git-branch-delete delete feature/risky-branch --force
```

### Interactive Mode

Interactive mode provides a user-friendly interface for managing multiple branches:

```bash
# Start interactive mode
git-branch-delete interactive

# Include remote branches
git-branch-delete interactive --all

# Force delete without merge checks
git-branch-delete interactive --force
```

Navigation:

- Use ↑/↓ to move between branches
- Space to select/deselect branches
- Enter to confirm selection
- q to quit without changes

### Common Workflows

#### Cleanup After Release

```bash
# 1. List stale branches
git-branch-delete list --stale

# 2. Prune remote branches
git-branch-delete prune --remote

# 3. Delete merged feature branches
git-branch-delete interactive --merged
```

#### Safe Remote Cleanup

```bash
# 1. Show what would be deleted
git-branch-delete prune --dry-run

# 2. Delete with confirmation
git-branch-delete prune

# 3. Force delete if needed
git-branch-delete prune --force
```

### Configuration

Create `~/.config/git-branch-delete.yaml`:

```yaml
# Default configuration
default_branch: main

# Protected branches
protected_branches:
  - main
  - master
  - develop

# Default remote
default_remote: origin

# Operation settings
auto_confirm: false
dry_run: false
```

### Environment Variables

Core configuration:

```bash
# Core settings
export GBD_DEFAULT_BRANCH=main
export GBD_DEFAULT_REMOTE=origin
export GBD_AUTO_CONFIRM=false
export GBD_DRY_RUN=false
```

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
