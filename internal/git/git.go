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
	workDir   string
	gitPath   string
	timeout   time.Duration
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

// SetTimeout sets the timeout for git commands
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

	// Filter and append our specific git environment variables
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		// Keep important environment variables
		if strings.HasPrefix(e, "HOME=") ||
			strings.HasPrefix(e, "PATH=") ||
			strings.HasPrefix(e, "SSH_") ||
			strings.HasPrefix(e, "GIT_") || // Keep all existing GIT_* variables
			strings.HasPrefix(e, "XDG_CONFIG_HOME=") ||
			strings.HasPrefix(e, "TERM=") {
			filteredEnv = append(filteredEnv, e)
		}
	}

	// Append our git-specific environment variables
	gitEnv := []string{
		"GIT_TERMINAL_PROMPT=1",     // Always enable terminal prompts
		"GIT_CONFIG_NOSYSTEM=1",     // Ignore system config
		"GIT_FLUSH=1",               // Disable output buffering
		"GIT_PROTOCOL=version=2",    // Use Git protocol v2
		"GIT_TRACE_PACK_ACCESS=",    // Disable pack tracing
		"GIT_TRACE_PACKET=",         // Disable packet tracing
		"GIT_TRACE=",                // Disable general tracing
		"LC_ALL=C",                  // Use consistent locale
	}

	cmd.Env = append(filteredEnv, gitEnv...)

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
	if strings.ContainsAny(output, "\x00\x07\x1B\x9B") {
		return "", newGitCommandError(strings.Join(args, " "), output, fmt.Errorf("output contains invalid characters"))
	}

	return strings.TrimSpace(output), nil
}

// execGitQuiet executes a git command without validation for internal use
func (g *Git) execGitQuiet(args ...string) (string, error) {
	cmd := exec.Command(g.gitPath, args...)
	cmd.Dir = g.workDir
	cmd.Stdin = os.Stdin  // Prevent hanging
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GitBranch represents a git branch and its metadata
type GitBranch struct {
	Name       string
	CommitHash string
	Reference  string
	IsCurrent  bool
	IsRemote   bool
	IsDefault  bool
	IsMerged   bool
	IsStale    bool
	IsBehind   bool
	Message    string
}

// execGitWithStdout executes a git command and returns its stdout pipe
func (g *Git) execGitWithStdout(args ...string) (*exec.Cmd, io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, g.gitPath, args...)
	cmd.Dir = g.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin  // Prevent hanging

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
func (g *Git) ParseBranchLine(line string) (GitBranch, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return GitBranch{}, fmt.Errorf("invalid branch line format: %s", line)
	}

	refName := parts[0]
	commitHash := parts[1]

	var trackingInfo string
	if len(parts) > 2 {
		trackingInfo = strings.Join(parts[2:], " ")
	}

	branch := GitBranch{
		Name:       strings.TrimPrefix(strings.TrimPrefix(refName, "refs/heads/"), "refs/remotes/"),
		CommitHash: commitHash,
		Reference:  refName,
		IsRemote:   strings.HasPrefix(refName, "refs/remotes/"),
		IsDefault:  g.isDefaultBranch(refName),
	}

	// Parse tracking info
	if strings.Contains(trackingInfo, "behind") {
		branch.IsBehind = true
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

// DeleteBranch deletes a git branch
func (g *Git) DeleteBranch(name string, force bool, remote bool) error {
	// Validate branch name
	if err := ValidateBranchName(name); err != nil {
		return newInvalidBranchError(name, err.Error())
	}

	// Don't allow deletion of protected branches
	if isProtectedBranch(name) {
		return newProtectedBranchError(name)
	}

	// Verify branch exists before attempting deletion
	exists, err := g.branchExists(name, remote)
	if err != nil {
		return newGitCommandError("branch exists check", "", err)
	}
	if !exists {
		return newInvalidBranchError(name, "branch does not exist")
	}

	// Check if branch is fully merged if not force deleting
	if !force && !remote {
		merged, err := g.isBranchMerged(name)
		if err != nil {
			return newGitCommandError("merge check", "", err)
		}
		if !merged {
			return newUnmergedBranchError(name)
		}
	}

	var args []string
	if remote {
		// For remote branches, use push --delete
		args = []string{"push", "origin", "--delete", name}
	} else {
		if force {
			args = []string{"branch", "-D", name}
		} else {
			args = []string{"branch", "-d", name}
		}
	}

	if remote {
		// For remote operations, use standard git command with proper error handling
		cmd := exec.Command(g.gitPath, args...)
		cmd.Dir = g.workDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				errStr := string(exitErr.Stderr)
				if strings.Contains(errStr, "could not read Username") ||
				   strings.Contains(errStr, "Authentication failed") {
					return fmt.Errorf("authentication failed. For HTTPS, run: git config --global credential.helper store\nFor SSH, ensure your SSH key is added to GitHub")
				}
				if strings.Contains(errStr, "Permission denied") {
					return fmt.Errorf("permission denied. Please check your credentials and repository permissions")
				}
				if strings.Contains(errStr, "remote rejected") {
					return fmt.Errorf("remote rejected deletion of branch '%s'. Check if you have write access to the repository", name)
				}
			}
			return fmt.Errorf("failed to delete remote branch: %w", err)
		}
		return nil
	}

	// For local branches, use regular execGit
	_, err = g.execGit(args...)
	return err
}

// isBranchMerged checks if a branch is fully merged into the current branch
func (g *Git) isBranchMerged(name string) (bool, error) {
	// Get the current branch first
	currentBranch, err := g.execGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if the branch is merged into the current branch
	out, err := g.execGit("branch", "--merged", currentBranch)
	if err != nil {
		return false, fmt.Errorf("failed to check merged branches: %w", err)
	}

	// Look for the branch in the merged list
	for _, line := range strings.Split(out, "\n") {
		// Remove leading whitespace and asterisk for current branch
		branch := strings.TrimLeft(strings.TrimSpace(line), "* ")
		if branch == name {
			return true, nil
		}
	}

	return false, nil
}

// branchExists checks if a branch exists
func (g *Git) branchExists(name string, remote bool) (bool, error) {
	var args []string
	if remote {
		args = []string{"ls-remote", "--heads", "origin", name}
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

// ListBranches lists all git branches
func (g *Git) ListBranches() ([]GitBranch, error) {
	// Get current branch for merge checks
	currentBranch, err := g.execGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get merged branches for quick lookup
	mergedOut, err := g.execGit("branch", "--merged", currentBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}
	mergedBranches := make(map[string]bool)
	for _, line := range strings.Split(mergedOut, "\n") {
		branch := strings.TrimLeft(strings.TrimSpace(line), "* ")
		if branch != "" {
			mergedBranches[branch] = true
		}
	}

	// Get local branches
	localOut, err := g.execGit("branch", "--format", "%(refname:short) %(objectname) %(upstream:short) %(if)%(HEAD)%(then)* %(else)  %(end)%(subject)")
	if err != nil {
		return nil, err
	}

	// Get remote branches
	remoteOut, err := g.execGit("branch", "-r", "--format", "%(refname:short) %(objectname) %(upstream:short) %(if)%(HEAD)%(then)* %(else)  %(end)%(subject)")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var branches []GitBranch

	// Process local branches
	for _, line := range strings.Split(localOut, "\n") {
		if line == "" {
			continue
		}

		branch := parseBranchLine(line)
		if branch.Name == "" || strings.HasSuffix(branch.Name, "/HEAD") {
			continue
		}

		if err := ValidateBranchName(branch.Name); err != nil {
			continue
		}

		// Set merged status
		branch.IsMerged = mergedBranches[branch.Name]

		if !seen[branch.Name] {
			seen[branch.Name] = true
			branches = append(branches, branch)
		}
	}

	// Process remote branches
	for _, line := range strings.Split(remoteOut, "\n") {
		if line == "" {
			continue
		}

		branch := parseBranchLine(line)
		if branch.Name == "" || strings.HasSuffix(branch.Name, "/HEAD") {
			continue
		}

		// Clean up remote branch names
		if strings.HasPrefix(branch.Name, "origin/") {
			branch.Name = strings.TrimPrefix(branch.Name, "origin/")
		}

		if err := ValidateBranchName(branch.Name); err != nil {
			continue
		}

		branch.IsRemote = true
		// For remote branches, check if merged using origin/branch notation
		branch.IsMerged = mergedBranches["origin/"+branch.Name]

		if !seen[branch.Name] {
			seen[branch.Name] = true
			branches = append(branches, branch)
		}
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

// parseBranchLine parses a line from git branch -v output
func parseBranchLine(line string) GitBranch {
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return GitBranch{}
	}

	name := parts[0]
	hash := parts[1]
	reference := parts[2]
	info := parts[3]

	// Skip special refs
	if name == "HEAD" || strings.HasPrefix(name, "heads/") {
		return GitBranch{}
	}

	return GitBranch{
		Name:       name,
		CommitHash: hash,
		Reference:  reference,
		IsCurrent:  strings.HasPrefix(info, "*"),
		IsRemote:   strings.HasPrefix(name, "origin/"),
		IsDefault:  isProtectedBranch(name),
		Message:    strings.TrimPrefix(info, "* "),
	}
}
