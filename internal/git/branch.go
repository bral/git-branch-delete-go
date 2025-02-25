package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// GetBranches returns a list of all git branches in the current repository
func GetBranches() ([]Branch, error) {
	// Get all branches with their commit info
	cmd := exec.Command("git", "for-each-ref", "--sort=-committerdate", "refs/heads/", "--format=%(if)%(HEAD)%(then)*%(else) %(end)%(refname:short):::%(objectname:short):::%(subject)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches: %w", err)
	}

	// Get list of merged branches
	mergedCmd := exec.Command("git", "branch", "--merged", "main")
	mergedOutput, err := mergedCmd.Output()
	if err != nil {
		// If main doesn't exist, try master
		mergedCmd = exec.Command("git", "branch", "--merged", "master")
		mergedOutput, err = mergedCmd.Output()
		if err != nil {
			mergedOutput = []byte{} // No merged branches if we can't determine
		}
	}

	// Parse merged branches into a map for quick lookup
	mergedBranches := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(mergedOutput)))
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		branch = strings.TrimPrefix(branch, "*") // Remove current branch marker
		branch = strings.TrimSpace(branch)       // Remove any remaining whitespace
		mergedBranches[branch] = true
	}

	var branches []Branch
	scanner = bufio.NewScanner(strings.NewReader(string(output)))
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
		isCurrent := strings.HasPrefix(prefix, "*")
		name := strings.TrimLeft(prefix, " *")
		commitHash := parts[1]
		message := parts[2]

		// Truncate message if longer than 30 chars
		if len(message) > 30 {
			message = message[:27] + "..."
		}

		branch := Branch{
			Name:       name,
			CommitHash: commitHash,
			Message:    message,
			IsCurrent:  isCurrent,
			IsMerged:   mergedBranches[name],
		}
		branches = append(branches, branch)
	}

	return branches, nil
}

// DeleteBranches deletes the specified branches
func DeleteBranches(branches []string) []BranchDeletionResult {
	results := make([]BranchDeletionResult, 0, len(branches))

	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "-D", branch)
		output, err := cmd.CombinedOutput()

		result := BranchDeletionResult{
			Name: branch,
		}

		if err != nil {
			result.Success = false
			result.Error = string(output)
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}
