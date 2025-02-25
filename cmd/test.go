package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/spf13/cobra"
)

var (
	testCount int
)

func init() {
	testCmd := newTestCmd()
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().IntVarP(&testCount, "count", "n", 5, "Number of test branches to create")
}

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Create random test branches",
		Long: `Create random test branches for testing purposes.
This will create both local and remote branches.`,
		Example: `  git-branch-delete test      # Create 5 test branches
  git-branch-delete test -n 10  # Create 10 test branches`,
		RunE: runTest,
	}
}

func generateRandomName() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("test_%s", hex.EncodeToString(bytes)), nil
}

func runTest(cmd *cobra.Command, args []string) error {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Initialize git
	g, err := git.New(wd)
	if err != nil {
		return fmt.Errorf("failed to initialize git in %s: %w", wd, err)
	}

	log.Info("Creating %d test branches...", testCount)

	for i := 0; i < testCount; i++ {
		// Generate random branch name
		name, err := generateRandomName()
		if err != nil {
			return fmt.Errorf("failed to generate branch name: %w", err)
		}

		// Create branch with test commit
		if err := g.CreateBranch(name, true); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", name, err)
		}

		// Push to remote
		if err := g.PushBranch(name); err != nil {
			log.Warn("Failed to push branch %s: %v", name, err)
		} else {
			log.Info("Created and pushed branch: %s", name)
		}
	}

	// Return to original branch
	if err := g.CheckoutBranch("-"); err != nil {
		return fmt.Errorf("failed to return to original branch: %w", err)
	}

	log.Info("\nCreated %d test branches successfully! 🎉", testCount)
	log.Info("Run 'git-branch-delete interactive --all' to clean them up")

	return nil
}
