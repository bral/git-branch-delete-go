package git

import (
	"fmt"
	"strings"
	"unicode"
)

var (
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
	// Allow empty arguments
	if arg == "" {
		return nil
	}

	// Allow certain safe git flags and options
	safeFlags := map[string]bool{
		"--abbrev-ref": true,
		"--format":     true,
		"--merged":     true,
		"--verify":     true,
		"--quiet":      true,
		"--heads":      true,
		"-r":          true,
		"-d":          true,
		"-D":          true,
		"--delete":    true,
	}

	// Check if it's a safe flag
	if safeFlags[arg] {
		return nil
	}

	// Check if it's a format string
	if strings.HasPrefix(arg, "%(") && strings.HasSuffix(arg, ")") {
		return nil
	}

	// Check if it's a ref path
	if strings.HasPrefix(arg, "refs/") {
		return nil
	}

	// Check if it's HEAD
	if arg == "HEAD" {
		return nil
	}

	// For other arguments, validate as branch name
	if strings.HasPrefix(arg, "-") {
		return fmt.Errorf("disallowed git argument: %s", arg)
	}

	return ValidateBranchName(arg)
}

// ValidateBranchName checks if a branch name is valid and safe
func ValidateBranchName(name string) error {
	// Check for empty name
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check for disallowed characters that could be used for command injection
	for _, char := range disallowedChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("branch name contains invalid character: %s", char)
		}
	}

	// Check for invalid sequences
	for _, seq := range invalidSequences {
		if strings.Contains(name, seq) {
			return fmt.Errorf("branch name cannot contain '%s'", seq)
		}
	}

	// Check for invalid endings
	for _, ending := range invalidEndings {
		if strings.HasSuffix(name, ending) {
			return fmt.Errorf("branch name cannot end with '%s'", ending)
		}
	}

	// Check for invalid beginnings
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("branch name cannot start with '-'")
	}

	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("branch name cannot start with '.'")
	}

	// Check for control characters and spaces
	for i, r := range name {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return fmt.Errorf("branch name contains invalid character at position %d: %q", i+1, r)
		}
	}

	// Check for ASCII control characters explicitly
	for i, r := range name {
		if r <= 0x20 || r == 0x7F {
			return fmt.Errorf("branch name contains control character at position %d: %q", i+1, r)
		}
	}

	// Check for consecutive dots
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain consecutive dots")
	}

	// Check for consecutive slashes
	if strings.Contains(name, "//") {
		return fmt.Errorf("branch name cannot contain consecutive slashes")
	}

	return nil
}

// SanitizeBranchName removes any potentially dangerous characters from a branch name
func SanitizeBranchName(name string) string {
	// Remove any characters that could be used for command injection
	for _, char := range disallowedChars {
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

	return strings.TrimSpace(name)
}
