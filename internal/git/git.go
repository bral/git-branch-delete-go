package git

import (
	"bytes"
	"context"
	"fmt"
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
			return "", fmt.Errorf("invalid git argument: %w", err)
		}
	}

	// Validate all branch name arguments
	for _, arg := range args {
		if strings.HasPrefix(arg, "refs/heads/") || strings.HasPrefix(arg, "refs/remotes/") {
			branchName := strings.TrimPrefix(strings.TrimPrefix(arg, "refs/heads/"), "refs/remotes/")
			if err := ValidateBranchName(branchName); err != nil {
				return "", fmt.Errorf("invalid branch name: %w", err)
			}
		}
	}

	// Use absolute path to git executable
	cmd := exec.CommandContext(ctx, g.gitPath, args...)
	cmd.Dir = g.workDir

	// Use separate buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set minimal secure environment
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cmd.Env = []string{
		"HOME=" + home,
		"GIT_TERMINAL_PROMPT=0", // Disable git credential prompting
		"GIT_ASKPASS=", // Disable password prompting
		"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=yes", // Enforce SSH key checking
		"LC_ALL=C", // Use consistent locale
		"GIT_CONFIG_NOSYSTEM=1", // Ignore system config
		"GIT_FLUSH=1", // Disable output buffering
		"GIT_PROTOCOL=version=2", // Use Git protocol v2
		"GIT_TRACE_PACK_ACCESS=", // Disable pack tracing
		"GIT_TRACE_PACKET=", // Disable packet tracing
		"GIT_TRACE=", // Disable general tracing
	}

	// Execute command with timeout
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git command timed out after %v", g.timeout)
		}
		return "", fmt.Errorf("git command failed: %s: %w", stderr.String(), err)
	}

	// Validate output for potential command injection
	output := stdout.String()
	if strings.ContainsAny(output, "\x00\x07\x1B\x9B") {
		return "", fmt.Errorf("git output contains invalid characters")
	}

	return strings.TrimSpace(output), nil
}

// GitBranch represents a git branch
type GitBranch struct {
	Name       string
	CommitHash string
	Reference  string
	IsCurrent  bool
	IsRemote   bool
	IsDefault  bool
	IsMerged   bool
	IsStale    bool
	Message    string
}

// DeleteBranch deletes a git branch
func (g *Git) DeleteBranch(name string, force bool, remote bool) error {
	// Validate branch name
	if err := ValidateBranchName(name); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	// Don't allow deletion of protected branches
	if isProtectedBranch(name) {
		return fmt.Errorf("cannot delete protected branch: %s", name)
	}

	// Verify branch exists before attempting deletion
	exists, err := g.branchExists(name, remote)
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}
	if !exists {
		return fmt.Errorf("branch does not exist: %s", name)
	}

	// Check if branch is fully merged if not force deleting
	if !force && !remote {
		merged, err := g.isBranchMerged(name)
		if err != nil {
			return fmt.Errorf("failed to check if branch is merged: %w", err)
		}
		if !merged {
			return fmt.Errorf("branch %s is not fully merged", name)
		}
	}

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
