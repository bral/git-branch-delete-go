package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// Git handles git operations
type Git struct {
	workDir string
}

// New creates a new Git instance
func New(workDir string) *Git {
	return &Git{workDir: workDir}
}

// ListBranches returns all branches with detailed information
func (g *Git) ListBranches() ([]Branch, error) {
	if err := g.verifyRepo(); err != nil {
		return nil, err
	}

	// Get all branches with their commit info
	cmd := exec.Command("git", "for-each-ref", "--sort=-committerdate", "refs/heads/", "refs/remotes/", "--format=%(if)%(HEAD)%(then)*%(else) %(end)%(refname:short):::%(objectname:short):::%(subject)")
	cmd.Dir = g.workDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	currentBranch, err := g.getCurrentBranch()
	if err != nil {
		return nil, err
	}

	defaultBranch, err := g.getDefaultBranch()
	if err != nil {
		// Don't fail if we can't determine default branch
		defaultBranch = ""
	}

	// Get merged branches
	mergedBranches, err := g.getMergedBranches()
	if err != nil {
		// Non-fatal error, continue without merged info
		mergedBranches = make(map[string]bool)
	}

	var branches []Branch
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":::")
		if len(parts) < 3 {
			continue
		}

		prefix := parts[0]
		name := strings.TrimSpace(prefix)
		isCurrent := strings.HasPrefix(prefix, "*") || name == currentBranch
		if isCurrent {
			name = strings.TrimLeft(name, "* ")
		}

		isRemote := strings.HasPrefix(name, "remotes/")
		if isRemote {
			name = strings.TrimPrefix(name, "remotes/")
		}

		branch := Branch{
			Name:       name,
			CommitHash: parts[1],
			Message:    parts[2],
			IsLocal:    !isRemote,
			IsRemote:   isRemote,
			IsCurrent:  isCurrent,
			IsDefault:  defaultBranch != "" && (name == defaultBranch || name == "origin/"+defaultBranch),
			IsMerged:   mergedBranches[name],
		}

		branches = append(branches, branch)
	}

	// Check for stale branches (non-fatal)
	_ = g.markStaleBranches(branches)

	return branches, nil
}

// DeleteBranch deletes a branch locally and/or remotely
func (g *Git) DeleteBranch(name string, force, remote bool) error {
	if err := g.verifyRepo(); err != nil {
		return err
	}

	if remote {
		remoteName := strings.Split(name, "/")[0]
		branchName := strings.Join(strings.Split(name, "/")[1:], "/")

		args := []string{"push", remoteName, "--delete", branchName}
		cmd := exec.Command("git", args...)
		cmd.Dir = g.workDir

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to delete remote branch: %w", err)
		}
	}

	flag := "-d"
	if force {
		flag = "-D"
	}

	cmd := exec.Command("git", "branch", flag, name)
	cmd.Dir = g.workDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete local branch: %w", err)
	}

	return nil
}

func (g *Git) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (g *Git) getDefaultBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = g.workDir

	output, err := cmd.Output()
	if err != nil {
		// Fallback to common default branch names
		for _, name := range []string{"main", "master"} {
			if g.branchExists(name) {
				return name, nil
			}
		}
		return "", fmt.Errorf("failed to determine default branch")
	}

	ref := strings.TrimSpace(string(output))
	return strings.TrimPrefix(ref, "refs/remotes/origin/"), nil
}

func (g *Git) getMergedBranches() (map[string]bool, error) {
	mergedBranches := make(map[string]bool)

	// Try main first, then master
	for _, base := range []string{"main", "master"} {
		cmd := exec.Command("git", "branch", "--merged", base)
		cmd.Dir = g.workDir
		output, err := cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(output)))
			for scanner.Scan() {
				branch := strings.TrimSpace(scanner.Text())
				branch = strings.TrimPrefix(branch, "*") // Remove current branch marker
				branch = strings.TrimSpace(branch)       // Remove any remaining whitespace
				mergedBranches[branch] = true
			}
			return mergedBranches, nil
		}
	}

	return mergedBranches, fmt.Errorf("failed to get merged branches")
}

func (g *Git) branchExists(name string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", name)
	cmd.Dir = g.workDir
	return cmd.Run() == nil
}

func (g *Git) verifyRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.workDir

	if err := cmd.Run(); err != nil {
		return &ErrNotGitRepo{Dir: g.workDir}
	}
	return nil
}

func (g *Git) markStaleBranches(branches []Branch) error {
	for i := range branches {
		if branches[i].IsRemote {
			continue
		}

		cmd := exec.Command("git", "branch", "-v", "--format", "%(upstream:track)", branches[i].Name)
		cmd.Dir = g.workDir

		output, err := cmd.Output()
		if err != nil {
			// Don't fail if we can't check upstream status
			// This can happen with new repos or branches without upstream
			branches[i].IsStale = false
			continue
		}

		branches[i].IsStale = strings.Contains(string(output), "gone")
	}
	return nil
}
