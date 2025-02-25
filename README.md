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

# Start with remote branches
git-branch-delete interactive --remote

# Start with specific filter
git-branch-delete interactive --filter="feature/*"
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

#### Batch Delete Feature Branches

```bash
# 1. List all feature branches
git-branch-delete list --filter="feature/*"

# 2. Delete all merged feature branches
git-branch-delete delete --pattern="feature/*" --merged

# 3. Verify deletion
git-branch-delete list --all
```

#### Safe Remote Cleanup

```bash
# 1. Show what would be deleted
git-branch-delete prune --dry-run --remote

# 2. Delete with confirmation
git-branch-delete prune --remote

# 3. Cleanup local references
git-branch-delete prune --local
```

### Advanced Configuration

The configuration file (`~/.config/git-branch-delete.yaml`) supports advanced options:

```yaml
# Default configuration
default_branch: main
default_remote: origin

# Branch protection
protected_branches:
  - main
  - master
  - develop
  - release/*
  - hotfix/*

# Remote settings
remotes:
  - origin
  - upstream

# Deletion settings
deletion:
  auto_confirm: false
  dry_run: false
  force: false

# Pattern settings
patterns:
  include:
    - feature/*
    - bugfix/*
  exclude:
    - release/*
    - hotfix/*

# UI settings
ui:
  color: true
  interactive: true
  progress: true
```

### Environment Variables

Full list of supported environment variables:

```bash
# Core settings
export GBD_DEFAULT_BRANCH=main
export GBD_DEFAULT_REMOTE=origin
export GBD_AUTO_CONFIRM=false
export GBD_DRY_RUN=false

# Authentication
export GBD_GIT_USERNAME=your-username
export GBD_CREDENTIAL_HELPER=osxkeychain

# UI settings
export GBD_COLOR=true
export GBD_PROGRESS=true
export GBD_INTERACTIVE=true

# Logging
export GBD_LOG_LEVEL=info
export GBD_LOG_FILE=/path/to/log
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
