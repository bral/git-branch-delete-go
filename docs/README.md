# git-branch-delete Documentation

Welcome to the git-branch-delete documentation! This tool helps you manage Git branches safely and efficiently.

## Quick Links

- [API Documentation](api/README.md) - Detailed documentation of the Go API
- [Examples](examples/) - Code examples and common use cases
  - [Batch Branch Deletion](examples/batch-delete.md)
  - [Listing Branches](examples/list-branches.md)
- [Contributing Guide](contributing/README.md) - How to contribute to the project

## Getting Started

1. Installation:

```bash
# Install the latest version
go install github.com/bral/git-branch-delete-go@latest

# Install specific version
go install github.com/bral/git-branch-delete-go@v2.0.1
```

2. Basic Usage:

```bash
# List branches
git-branch-delete list

# Delete a branch
git-branch-delete delete feature/123

# Interactive mode
git-branch-delete interactive
```

## Command Reference

- `list` - List git branches
- `delete` - Delete git branches
- `interactive` - Interactively select branches to delete
- `prune` - Delete stale branches
- `test` - Create random test branches
- `version` - Print version information
- `completion` - Generate shell completion scripts

## Configuration

Configuration is stored in `.gitconfig`:

```ini
[git-branch-delete]
    protected = main,master,develop
    remote = origin
    days = 30
```

## Support

- [GitHub Issues](https://github.com/bral/git-branch-delete-go/issues)
- [GitHub Discussions](https://github.com/bral/git-branch-delete-go/discussions)

## License

MIT License - see [LICENSE](../LICENSE) for details
