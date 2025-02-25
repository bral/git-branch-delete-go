# Contributing to git-branch-delete

Thank you for your interest in contributing to git-branch-delete! This guide will help you get started.

## Development Setup

1. Fork and clone the repository:

```bash
git clone https://github.com/YOUR_USERNAME/git-branch-delete-go.git
cd git-branch-delete-go
```

2. Install dependencies:

```bash
go mod download
```

3. Run tests:

```bash
go test ./...
```

## Code Style

- Follow standard Go conventions and [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` to format code
- Add comments for exported functions and types
- Keep functions focused and small
- Write tests for new functionality

## Pull Request Process

1. Create a feature branch:

```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and commit:

```bash
git commit -m "feat: add new feature"
```

3. Run tests and linting:

```bash
go test ./...
golangci-lint run
```

4. Push changes and create PR:

```bash
git push origin feature/your-feature-name
```

5. Create a Pull Request on GitHub

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Adding/updating tests
- `refactor:` Code changes that neither fix bugs nor add features
- `chore:` Changes to build process, etc.

## Testing

- Write unit tests for new functionality
- Include integration tests for CLI commands
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example test:

```go
func TestValidateBranchName(t *testing.T) {
    tests := []struct {
        name    string
        branch  string
        wantErr bool
    }{
        {"valid branch", "feature/123", false},
        {"invalid chars", "feature/123!", true},
        {"protected name", "HEAD", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateBranchName(tt.branch)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateBranchName() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for new types and functions
- Include examples in docs/examples/
- Update API documentation in docs/api/

## Release Process

1. Update version in version.go
2. Update CHANGELOG.md
3. Create a release PR
4. After merge, tag the release:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

## Getting Help

- Open an issue for bugs or feature requests
- Ask questions in discussions
- Join our community chat
