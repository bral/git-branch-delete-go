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
	validCharsRegex    = regexp.MustCompile(`^[a-zA-Z0-9-/_]+$`)

	// Allowed git command arguments
	allowedGitArgs = map[string]bool{
		// Commands
		"branch":        true,
		"push":         true,
		"pull":         true,
		"fetch":        true,
		"checkout":     true,
		"rev-parse":    true,
		"show-ref":     true,
		"ls-remote":    true,
		"for-each-ref": true,
		"gc":           true,
		"prune":        true,
		"pack-refs":    true,

		// Common flags
		"-c":            true,
		"-d":            true,
		"-D":            true,
		"--delete":      true,
		"--merged":      true,
		"--format":      true,
		"--verify":      true,
		"--quiet":       true,
		"--all":         true,
		"--prune":       true,
		"--auto":        true,
		"-r":            true,
		"--heads":       true,
		"--abbrev-ref":  true,
		"--no-contains": true,
		"--contains":    true,
		"--sort":        true,
		"--points-at":   true,
		"--porcelain":   true,
		"--progress":    true,
		"HEAD":          true,
		"origin":        true,
	}

	// Dangerous patterns that could be used for command injection
	dangerousPatterns = []string{
		";", "&", "|", "`", "$", "(", ")", "<", ">", "\\",
		"\n", "\r", "\t", "\v", "\f",
		"../", ".../", "~",
	}

	// disallowedChars are characters that could be used for command injection
	disallowedChars = []string{";", "&&", "||", "`", "$", "|", ">", "<", "(", ")", "{", "}", "[", "]", "\"", "'", "\n", "\r"}

	// invalidEndings are invalid branch name endings
	invalidEndings = []string{"/", ".", ".lock"}

	// invalidSequences are invalid sequences in branch names
	invalidSequences = []string{"..", "//", "@{", ".lock/"}

	// allowedGitCommands are the only git commands we permit
	allowedGitCommands = map[string]bool{
		"branch":        true,
		"show-ref":     true,
		"push":         true,
		"ls-remote":    true,
		"for-each-ref": true,
	}

	// allowedGitFlags are the only git flags we permit
	allowedGitFlags = map[string]bool{
		"--format":      true,
		"--verify":      true,
		"--quiet":       true,
		"--delete":      true,
		"-d":           true,
		"-D":           true,
		"-r":           true,
		"--heads":      true,
		"--merged":     true,
		"refs/heads":   true,
		"refs/remotes": true,
		"origin":       true,
	}
)

// ValidateGitArg validates a git command argument
func ValidateGitArg(arg string) error {
	// Skip validation for specific patterns
	if strings.HasPrefix(arg, "refs/") ||
	   strings.HasPrefix(arg, "%(") ||
	   strings.HasPrefix(arg, "credential.") {
		return nil
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("argument contains dangerous pattern: %s", pattern)
		}
	}

	// Check if it's an allowed argument
	if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
		if !allowedGitArgs[arg] {
			return fmt.Errorf("unsupported git flag: %s", arg)
		}
		return nil
	}

	// If it's a command, check if it's allowed
	if allowedGitArgs[arg] {
		return nil
	}

	// For branch names and other arguments, validate characters
	return ValidateBranchName(arg)
}

// ValidateBranchName validates a git branch name
func ValidateBranchName(name string) error {
	// Check length
	if len(name) == 0 {
		return fmt.Errorf("branch name cannot be empty")
	}
	if len(name) > 255 {
		return fmt.Errorf("branch name too long (max 255 characters)")
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if strings.Contains(name, pattern) {
			return fmt.Errorf("branch name contains dangerous pattern: %s", pattern)
		}
	}

	// Check branch naming rules
	if branchStartDotRegex.MatchString(name) {
		return fmt.Errorf("branch name cannot start with '.'")
	}
	if doubleDotRegex.MatchString(name) {
		return fmt.Errorf("branch name cannot contain '..'")
	}
	if endSlashRegex.MatchString(name) {
		return fmt.Errorf("branch name cannot end with '/'")
	}
	if endLockRegex.MatchString(name) {
		return fmt.Errorf("branch name cannot end with '.lock'")
	}
	if !validCharsRegex.MatchString(name) {
		return fmt.Errorf("branch name can only contain letters, numbers, dashes, underscores, and forward slashes")
	}

	return nil
}

// SanitizeBranchName removes any potentially dangerous characters from a branch name
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
