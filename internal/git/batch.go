package git

import (
	"context"
	"sync"
)

const batchSize = 10

type batchOperation struct {
	changes  []change
	rollback func() error
}

type change struct {
	branch   GitBranch
	action   string
	rollback func() error
}

// BatchProcessor handles batch operations on git branches.
// It provides concurrent processing with proper error handling.
type BatchProcessor struct {
	git *Git
}

// NewBatchProcessor creates a new BatchProcessor instance.
//
// Parameters:
//   - g: Git instance to use for operations
//
// Returns:
//   - *BatchProcessor: New processor instance
func NewBatchProcessor(g *Git) *BatchProcessor {
	return &BatchProcessor{
		git: g,
	}
}

// ProcessBranches processes multiple branches concurrently.
//
// Parameters:
//   - ctx: Context for cancellation
//   - branches: List of branches to process
//   - fn: Function to apply to each branch
//
// Returns an error if:
//   - Context is cancelled
//   - Any branch operation fails
//
// Example:
//
//	processor := NewBatchProcessor(git)
//	err := processor.ProcessBranches(ctx, branches, func(b GitBranch) error {
//	    return git.DeleteBranch(b.Name, false, false)
//	})
func (bp *BatchProcessor) ProcessBranches(ctx context.Context, branches []GitBranch, fn func(GitBranch) error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(branches))

	for i := 0; i < len(branches); i += batchSize {
		end := min(i+batchSize, len(branches))
		batch := branches[i:end]

		wg.Add(1)
		go func(batch []GitBranch) {
			defer wg.Done()
			for _, branch := range batch {
				if err := fn(branch); err != nil {
					errChan <- err
					return
				}
			}
		}(batch)
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
