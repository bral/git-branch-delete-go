package git

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Git branch naming rules:
	// - Cannot start with '.'
	// - Cannot have double dots '..'
	// - Cannot have ASCII control characters
	// - Cannot contain: space, ~, ^, :, ?, *, [, \
	// - Cannot end with '/'
	// - Cannot end with '.lock'
	// Using multiple regexes instead of negative lookahead
	branchStartDotRegex = regexp.MustCompile(`^\.`)
	doubleDotRegex     = regexp.MustCompile(`\.\.`)
	endSlashRegex      = regexp.MustCompile(`/$`)
	endLockRegex       = regexp.MustCompile(`\.lock$`)
	// More restrictive valid chars regex
	validCharsRegex    = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9/_]+$`)

	// Consolidated git command validation
	allowedGitCommands = map[string]bool{
		// Core commands we use
		"branch":        true,
		"push":         true,
		"rev-parse":    true,
		"show-ref":     true,
		"ls-remote":    true,
		"for-each-ref": true,
		"checkout":     true,  // For branch creation and switching
		"commit":       true,  // For creating test commits
	}

	// Allowed git flags with descriptions for security audit
	allowedGitFlags = map[string]bool{
		// Branch operations
		"-d":            true, // Delete branch
		"-D":            true, // Force delete branch
		"-b":            true, // Create and checkout branch
		"--delete":      true, // Delete branch (long form)
		"--force":       true, // Force operation
		"--allow-empty": true, // Allow empty commits

		// Branch listing and info
		"-r":            true, // Remote branches
		"--remotes":     true, // Remote branches (long form)
		"--merged":      true, // List merged branches
		"--no-merged":   true, // List unmerged branches
		"--format":      true, // Custom format
		"--abbrev-ref": true,  // Short ref names
		"--verify":     true,  // Verify ref exists
		"--quiet":      true,  // Suppress output
		"--porcelain":  true,  // Machine-readable output
		"-v":           true,  // Verbose
		"-vv":          true,  // Very verbose
		"--short":      true,  // Short SHA

		// Remote operations
		"origin":       true,  // Default remote name
		"--progress":   true,  // Show progress
		"--all":        true,  // All refs

		// Special refs
		"HEAD":         true,  // Current HEAD
		"refs/heads":   true,  // Local branches
		"refs/remotes": true,  // Remote branches

		// Git config
		"-c":           true,  // Set config
	}

	// Dangerous patterns that could be used for command injection
	dangerousPatterns = []string{
		";", "&", "|", "`", "$", "(", ")", "<", ">", "\\",
		"\n", "\r", "\t", "\v", "\f",
		"../", ".../", "~", "%", "@{",
		":", "?", "*", "[", "]", "{", "}", "'", "\"",
	}

	// More comprehensive invalid sequences
	invalidSequences = []string{
		"..", "//", "@{", ".lock/", "/.git/", ".git/",
		"../", "..\\", ".\\", "\\", "./../", "/..",
	}

	// More restrictive branch name pattern
	// - Must start with alphanumeric
	// - Can contain alphanumeric, dash, underscore, forward slash
	// - Cannot end with slash or dot
	// - Maximum length enforced separately
	branchNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9/_]*[a-zA-Z0-9]$`)
)

// ValidateGitArg validates a git command argument for security.
//
// Parameters:
//   - arg: The argument to validate
//
// Returns an error if:
//   - Argument contains shell metacharacters
//   - Argument contains control characters
//   - Argument contains path traversal sequences
//
// Example:
//
//	err := ValidateGitArg("feature/123")
//	if err != nil {
//	    log.Fatal("Invalid branch name:", err)
//	}
func ValidateGitArg(arg string) error {
	// Allow empty arguments
	if arg == "" {
		return nil
	}

	// Check if it's an allowed command
	if allowedGitCommands[arg] {
		return nil
	}

	// Check if it's an allowed flag
	if allowedGitFlags[arg] {
		return nil
	}

	// Check if it's a format specifier
	if strings.HasPrefix(arg, "%(") && strings.HasSuffix(arg, ")") {
		return nil
	}

	// Check if it's a ref path
	if strings.HasPrefix(arg, "refs/") {
		return ValidateBranchName(strings.TrimPrefix(arg, "refs/"))
	}

	// Check if it's a branch name
	if branchNamePattern.MatchString(arg) {
		return nil
	}

	return fmt.Errorf("unsupported git argument: %s", arg)
}

// ValidateBranchName validates a git branch name.
//
// Parameters:
//   - name: The branch name to validate
//
// Returns an error if:
//   - Name contains invalid characters
//   - Name starts or ends with '/'
//   - Name contains '..' sequence
//   - Name matches Git's reserved names
//
// Example:
//
//	err := ValidateBranchName("feature/new-branch")
//	if err != nil {
//	    log.Fatal("Invalid branch name:", err)
//	}
func ValidateBranchName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if !branchNamePattern.MatchString(name) {
		return fmt.Errorf("invalid branch name format")
	}

	return nil
}

// SanitizeBranchName sanitizes a branch name for safe use.
//
// Parameters:
//   - name: The branch name to sanitize
//
// Returns:
//   - string: Sanitized branch name safe for git operations
//
// Example:
//
//	safeName := SanitizeBranchName("feature/my-branch")
func SanitizeBranchName(name string) string {
	// Remove any characters that could be used for command injection
	for _, char := range dangerousPatterns {
		name = strings.ReplaceAll(name, char, "")
	}

	// Remove any control characters and spaces
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return -1
		}
		return r
	}, name)

	// Remove any invalid sequences
	for _, seq := range invalidSequences {
		name = strings.ReplaceAll(name, seq, "")
	}

	// Remove leading dots and dashes
	name = strings.TrimLeft(name, ".-")

	// Remove trailing dots and slashes
	name = strings.TrimRight(name, "./")

	// Replace any remaining invalid characters with dashes
	name = validCharsRegex.ReplaceAllString(name, "-")

	return strings.TrimSpace(name)
}
