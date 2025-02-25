# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.1] - 2024-02-25

### Added

- Simple UI implementation for branch selection and deletion confirmation
- Interactive command with force and all flags
- Comprehensive test suite for all components
- Error handling for Git operations
- Support for local branch operations

### Changed

- Moved Git package from internal to pkg directory
- Improved error messages and user feedback
- Enhanced test coverage and organization

### Fixed

- Command test initialization
- Git client error handling
- Branch type consistency

## [1.0.0] - YYYY-MM-DD

### Added

- Initial release
- Basic branch deletion functionality
- Command-line interface
- Git operations wrapper
- Error handling
- Test suite

### Changed

### Deprecated

### Removed

### Fixed

### Security

- Secure Git command execution with argument validation
- Protected environment variable handling
- Authentication error handling with helpful messages

[Unreleased]: https://github.com/bral/git-branch-delete-go/compare/v2.0.1...HEAD
[2.0.1]: https://github.com/bral/git-branch-delete-go/compare/v1.0.0...v2.0.1
