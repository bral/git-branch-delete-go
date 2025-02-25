package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/spf13/cobra"
)

var (
	pruneForce bool
)

func init() {
	pruneCmd := newPruneCmd()
	rootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "Force delete branches without confirmation")
}

func newPruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Delete stale branches",
		Long: `Delete branches that have been merged or deleted from remote.
By default, asks for confirmation before deleting.`,
		Example: `  git-branch-delete prune
  git-branch-delete prune --force`,
		RunE: runPrune,
	}
}

func runPrune(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	g, err := git.New(wd)
	if err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	branches, err := g.ListBranches()
	if err != nil {
		return err
	}

	var staleBranches []git.GitBranch
	for _, b := range branches {
		if b.IsStale && !b.IsCurrent && !b.IsDefault {
			staleBranches = append(staleBranches, b)
		}
	}

	if len(staleBranches) == 0 {
		log.Info("No stale branches found")
		return nil
	}

	log.Info("Found %d stale branches:", len(staleBranches))
	for _, b := range staleBranches {
		log.Info("  %s", b.Name)
	}

	if !pruneForce {
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Delete %d stale branches?", len(staleBranches)),
		}

		err = survey.AskOne(prompt, &confirm)
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirm {
			log.Info("Operation cancelled")
			return nil
		}
	}

	for _, b := range staleBranches {
		log.Info("Deleting branch: %s", b.Name)

		err := g.DeleteBranch(b.Name, true, b.IsRemote)
		if err != nil {
			log.Error("Failed to delete branch %s: %v", b.Name, err)
			continue
		}

		log.Success("Deleted branch: %s", b.Name)
	}

	return nil
}
