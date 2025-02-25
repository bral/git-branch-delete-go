package git

import (
	"bufio"
	"context"
	"fmt"
)

// BranchStream provides a memory-efficient way to process branches
type BranchStream struct {
	git *Git
}

func NewBranchStream(g *Git) *BranchStream {
	return &BranchStream{git: g}
}

// StreamBranches streams branches one at a time through the channel
func (bs *BranchStream) StreamBranches(ctx context.Context) (branches <-chan Branch, errs <-chan error) {
	branchChan := make(chan Branch)
	errChan := make(chan error, 1)

	go func() {
		defer close(branchChan)
		defer close(errChan)

		// Use --format to get branch info in a parseable format
		cmd, stdout, err := bs.git.execGitWithStdout("for-each-ref", "--format=%(refname) %(objectname) %(upstream:track)", "refs/heads", "refs/remotes")
		if err != nil {
			errChan <- fmt.Errorf("failed to start git command: %w", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				line := scanner.Text()
				branch, err := bs.git.ParseBranchLine(line)
				if err != nil {
					errChan <- fmt.Errorf("failed to parse branch info: %w", err)
					return
				}
				branchChan <- branch
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading branches: %w", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			errChan <- fmt.Errorf("git command failed: %w", err)
			return
		}
	}()

	return branchChan, errChan
}

// CleanupRefs performs repository cleanup and optimization
func (bs *BranchStream) CleanupRefs(ctx context.Context) error {
	// Run cleanup operations in sequence
	ops := []struct {
		name string
		args []string
	}{
		{"Pruning unreachable objects", []string{"prune"}},
		{"Cleaning up loose refs", []string{"pack-refs", "--all", "--prune"}},
		{"Running garbage collection", []string{"gc", "--auto", "--prune=now"}},
	}

	for _, op := range ops {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if _, err := bs.git.execGit(op.args...); err != nil {
				return fmt.Errorf("%s failed: %w", op.name, err)
			}
		}
	}

	return nil
}
