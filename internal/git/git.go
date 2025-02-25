package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the default timeout for git commands
	DefaultTimeout = 30 * time.Second
)

// Git represents a git repository
type Git struct {
	workDir string
	gitPath string
	timeout time.Duration
}

// New creates a new Git instance
func New(workDir string) (*Git, error) {
	// Find git executable path explicitly
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git executable not found: %w", err)
	}

	// Verify workDir exists and is absolute
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("invalid working directory: %w", err)
	}

	// Verify workDir is a git repository
	gitDir := filepath.Join(workDir, ".git")
	if fi, err := os.Stat(gitDir); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("not a git repository: %s", workDir)
	}

	return &Git{
		workDir: workDir,
		gitPath: gitPath,
		timeout: DefaultTimeout,
	}, nil
}

// SetTimeout sets the timeout duration for git operations.
// If timeout is less than or equal to zero, the timeout will not be changed.
// The default timeout is 30 seconds.
func (g *Git) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		g.timeout = timeout
	}
}

// execGit executes a git command securely with timeout
func (g *Git) execGit(args ...string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Validate all arguments
	for _, arg := range args {
		// Skip format strings and ref paths
		if strings.HasPrefix(arg, "%(") || strings.HasPrefix(arg, "refs/") {
			continue
		}
		if err := ValidateGitArg(arg); err != nil {
			return "", newInvalidBranchError(arg, err.Error())
		}
	}

	// Use absolute path to git executable
	cmd := exec.CommandContext(ctx, g.gitPath, args...)
	cmd.Dir = g.workDir

	// Use separate buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Always set stdin to prevent hanging
	cmd.Stdin = os.Stdin

	// Get existing environment
	env := os.Environ()

	// Explicitly allowed environment variables
	allowedEnvPrefixes := map[string]bool{
		"HOME=":            true, // Required for git config
		"USER=":            true, // Required for git config
		"PATH=":            true, // Required for git executable
		"SSH_AUTH_SOCK=":   true, // Required for SSH auth
		"SSH_AGENT_PID=":   true, // Required for SSH auth
		"DISPLAY=":         true, // Required for SSH askpass
		"TERM=":            true, // Required for terminal output
		"LANG=":            true, // Required for locale
		"LC_ALL=":          true, // Required for locale
		"XDG_CONFIG_HOME=": true, // Required for git config
		"XDG_CACHE_HOME=":  true, // Required for git credential
	}

	// Explicitly allowed GIT_ variables
	allowedGitVars := map[string]bool{
		"GIT_TERMINAL_PROMPT":   true,
		"GIT_ASKPASS":           true,
		"GIT_SSH":               true,
		"GIT_SSH_COMMAND":       true,
		"GIT_CONFIG_NOSYSTEM":   true,
		"GIT_AUTHOR_NAME":       true,
		"GIT_AUTHOR_EMAIL":      true,
		"GIT_COMMITTER_NAME":    true,
		"GIT_COMMITTER_EMAIL":   true,
		"GIT_CREDENTIAL_HELPER": true,
	}

	// Filter environment variables
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		// Check if it's an explicitly allowed env var
		allowed := false
		for prefix := range allowedEnvPrefixes {
			if strings.HasPrefix(e, prefix) {
				allowed = true
				break
			}
		}

		// Check if it's an allowed GIT_ variable
		if strings.HasPrefix(e, "GIT_") {
			varName := strings.SplitN(e, "=", 2)[0]
			if allowedGitVars[varName] {
				allowed = true
			}
		}

		if allowed {
			filteredEnv = append(filteredEnv, e)
		}
	}

	// Append our git-specific environment variables
	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=1",  // Always enable terminal prompts
		"GIT_PROTOCOL=version=2", // Use Git protocol v2
		"LC_ALL=C",               // Use consistent locale
	}

	filteredEnv = append(filteredEnv, gitEnv...)
	cmd.Env = filteredEnv

	// Execute command with timeout
	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", newTimeoutError(strings.Join(args, " "), g.timeout.String())
		}
		return "", newGitCommandError(strings.Join(args, " "), stderr.String(), err)
	}

	// Validate output for potential command injection
	output := stdout.String()
	for _, r := range []rune{'\x00', '\x07', '\x1B', '\x9B'} {
		if strings.ContainsRune(output, r) {
			return "", newGitCommandError(strings.Join(args, " "), output, fmt.Errorf("output contains invalid characters"))
		}
	}

	return strings.TrimSpace(output), nil
}

// execGitWithStdout executes a git command and returns its stdout pipe
func (g *Git) execGitWithStdout(args ...string) (*exec.Cmd, io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, g.gitPath, args...)
	cmd.Dir = g.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Prevent hanging

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	return cmd, stdout, nil
}

// ParseBranchLine parses a line of branch information from git for-each-ref
func (g *Git) ParseBranchLine(line string) (Branch, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return Branch{}, fmt.Errorf("invalid branch line format: %s", line)
	}

	refName := parts[0]
	commitHash := parts[1]

	var trackingInfo string
	if len(parts) > 2 {
		trackingInfo = strings.Join(parts[2:], " ")
	}

	branch := Branch{
		Name:       strings.TrimPrefix(strings.TrimPrefix(refName, "refs/heads/"), "refs/remotes/"),
		CommitHash: commitHash,
		IsRemote:   strings.HasPrefix(refName, "refs/remotes/"),
		IsDefault:  g.isDefaultBranch(refName),
	}

	// Parse tracking info
	if strings.Contains(trackingInfo, "behind") {
		branch.IsMerged = true
	}
	if strings.Contains(trackingInfo, "gone") {
		branch.IsStale = true
	}

	return branch, nil
}

// isDefaultBranch checks if the given ref is a default branch (main/master)
func (g *Git) isDefaultBranch(ref string) bool {
	defaultBranches := []string{"refs/heads/main", "refs/heads/master"}
	for _, defaultBranch := range defaultBranches {
		if ref == defaultBranch {
			return true
		}
	}
	return false
}

// branchExists checks if a branch exists locally or remotely
func (g *Git) branchExists(name string, remote bool) (bool, error) {
	var args []string
	if remote {
		args = []string{"ls-remote", "origin", "refs/heads/" + name}
	} else {
		args = []string{"show-ref", "--verify", "--quiet", "refs/heads/" + name}
	}

	_, err := g.execGit(args...)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unknown revision") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// handleAuthError provides interactive help for authentication errors
func (g *Git) handleAuthError(_ string) error {
	// Try to verify remote access
	if err := g.verifyRemoteAccess(); err != nil {
		return fmt.Errorf("failed to verify remote access: %w", err)
	}
	return nil
}

// DeleteBranch deletes a git branch either locally or remotely.
// Parameters:
//   - name: The name of the branch to delete
//   - force: If true, force delete even if branch is not fully merged
//   - remote: If true, delete the remote branch instead of local
//
// Returns an error if:
//   - The branch doesn't exist
//   - Authentication fails for remote operations
//   - The branch is protected
//   - The branch is not fully merged (without force)
func (g *Git) DeleteBranch(name string, force, remote bool) error {
	// Check if branch exists
	exists, err := g.branchExists(name, remote)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("branch '%s' does not exist", name)
	}

	// For remote operations, verify access first
	if remote {
		if err := g.verifyRemoteAccess(); err != nil {
			if strings.Contains(err.Error(), "Authentication failed") ||
				strings.Contains(err.Error(), "could not read Username") ||
				strings.Contains(err.Error(), "Permission denied") {
				return g.handleAuthError(err.Error())
			}
			return err
		}
	}

	// Delete branch
	var args []string
	if remote {
		args = []string{"push", "origin", "--delete", name}
	} else {
		if force {
			args = []string{"branch", "-D", name}
		} else {
			args = []string{"branch", "-d", name}
		}
	}

	_, err = g.execGit(args...)
	if err != nil {
		// Handle authentication and permission errors
		errStr := err.Error()
		if strings.Contains(errStr, "Authentication failed") ||
			strings.Contains(errStr, "could not read Username") ||
			strings.Contains(errStr, "Permission denied") {
			return g.handleAuthError(errStr)
		}
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// verifyRemoteAccess checks if we can access the remote repository
func (g *Git) verifyRemoteAccess() error {
	// Try to list remote refs
	_, err := g.execGit("ls-remote", "--quiet", "origin")
	if err != nil {
		if strings.Contains(err.Error(), "could not read Username") ||
			strings.Contains(err.Error(), "Authentication failed") {
			return fmt.Errorf("authentication failed. For HTTPS, run: git config --global credential.helper store\nFor SSH, ensure your SSH key is added to GitHub")
		}
		if strings.Contains(err.Error(), "Permission denied") {
			return fmt.Errorf("permission denied. Please check your credentials and repository permissions")
		}
		return fmt.Errorf("failed to access remote repository: %w", err)
	}
	return nil
}

// getTrackingBranch returns the current tracking branch
func (g *Git) getTrackingBranch() string {
	currentTrackingBranch, err := g.execGit("rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		return "" // Don't fail if branch has no upstream
	}
	return currentTrackingBranch
}

// getMergedBranches returns a map of merged branches
func (g *Git) getMergedBranches() (map[string]bool, error) {
	mergedBranches := make(map[string]bool)

	// Get local merged branches
	mergedOut, err := g.execGit("branch", "--merged")
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	for _, line := range strings.Split(mergedOut, "\n") {
		branch := strings.TrimLeft(strings.TrimSpace(line), "* ")
		if branch != "" {
			mergedBranches[branch] = true
		}
	}

	// Get remote merged branches
	remoteMergedOut, err := g.execGit("branch", "--merged")
	if err == nil { // Don't fail if remote check fails
		for _, line := range strings.Split(remoteMergedOut, "\n") {
			branch := strings.TrimSpace(line)
			if branch != "" && !strings.HasSuffix(branch, "/HEAD") {
				mergedBranches[branch] = true
			}
		}
	}

	return mergedBranches, nil
}

// processLocalBranch processes a local branch line
func (g *Git) processLocalBranch(line string, mergedBranches map[string]bool, _ string) (Branch, error) {
	if line == "" {
		return Branch{}, fmt.Errorf("empty branch line")
	}

	line = strings.TrimSpace(line)
	isCurrent := strings.HasPrefix(line, "*")
	if isCurrent {
		line = strings.TrimPrefix(line, "*")
	}
	name := strings.TrimSpace(line)

	// Get commit hash for branch
	hash, err := g.execGit("rev-parse", "--short", name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get commit hash: %w", err)
	}

	branch := Branch{
		Name:       name,
		CommitHash: strings.TrimSpace(hash),
		IsLocal:    true,
		IsCurrent:  isCurrent,
		IsRemote:   false,
		IsDefault:  isProtectedBranch(name),
		IsMerged:   mergedBranches[name],
	}

	return branch, nil
}

// ListBranches returns a list of all branches in the repository.
// The returned list includes both local and remote branches with their current status.
// Each branch includes information about:
//   - Whether it's the current branch
//   - Whether it's a remote branch
//   - Whether it's the default branch
//   - Whether it's merged
//   - Whether it's stale (deleted from remote)
//   - The latest commit hash and message
//
// Returns an error if:
//   - Not in a git repository
//   - Git command fails
//   - Unable to parse branch information
func (g *Git) ListBranches() ([]Branch, error) {
	currentTrackingBranch := g.getTrackingBranch()

	mergedBranches, err := g.getMergedBranches()
	if err != nil {
		return nil, err
	}

	// Get all local branches
	localOut, err := g.execGit("branch")
	if err != nil {
		return nil, err
	}

	// Count number of lines to pre-allocate slice
	lineCount := len(strings.Split(localOut, "\n"))
	branches := make([]Branch, 0, lineCount)

	for _, line := range strings.Split(localOut, "\n") {
		if line == "" {
			continue
		}

		branch, err := g.processLocalBranch(line, mergedBranches, currentTrackingBranch)
		if err != nil {
			continue // Skip branches we can't process
		}
		branches = append(branches, branch)
	}

	return branches, nil
}

// isProtectedBranch checks if a branch is protected
func isProtectedBranch(name string) bool {
	protected := []string{"main", "master", "develop", "release"}
	name = strings.TrimSpace(strings.ToLower(name))
	for _, p := range protected {
		if name == p {
			return true
		}
	}
	return false
}

// CreateBranch creates a new git branch.
// Parameters:
//   - name: The name of the new branch
//   - createCommit: If true, creates an empty commit on the new branch
//
// Returns an error if:
//   - Branch name is invalid
//   - Branch already exists
//   - Not in a git repository
//   - Git command fails
func (g *Git) CreateBranch(name string, createCommit bool) error {
	// Create and checkout branch
	_, err := g.execGit("checkout", "-b", name)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	if createCommit {
		_, err = g.execGit("commit", "--allow-empty", "-m", fmt.Sprintf("Test commit for %s", name))
		if err != nil {
			return fmt.Errorf("failed to create test commit: %w", err)
		}
	}

	return nil
}

// PushBranch pushes a local branch to the remote repository.
// Parameters:
//   - name: The name of the branch to push
//
// Returns an error if:
//   - Branch doesn't exist locally
//   - Authentication fails
//   - Remote access fails
//   - Git command fails
func (g *Git) PushBranch(name string) error {
	_, err := g.execGit("push", "-u", "origin", name)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}
	return nil
}

// CheckoutBranch checks out the specified git branch.
// Parameters:
//   - name: The name of the branch to checkout
//
// Returns an error if:
//   - Branch doesn't exist
//   - Working directory is not clean
//   - Git command fails
func (g *Git) CheckoutBranch(name string) error {
	_, err := g.execGit("checkout", name)
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}
	return nil
}
