package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/bral/git-branch-delete-go/pkg/git"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	showRemote bool
	showAll    bool
)

func init() {
	listCmd := newListCmd()
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&showRemote, "remote", "r", false, "Show remote branches")
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show both local and remote branches")
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List git branches",
		Long: `List git branches with their current status.
Shows local branches by default.`,
		Example: `  git-branch-delete list
  git-branch-delete list --remote
  git-branch-delete list --all`,
		RunE: runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	log.Debug("Starting branch listing")

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current directory: %v", err)
		return err
	}

	// Initialize git client
	gitClient := git.New(dir)
	branches, err := gitClient.ListBranches()
	if err != nil {
		log.Error("Failed to list branches: %v", err)
		return err
	}

	log.Debug("Retrieved branches: %d total", len(branches))

	// Filter branches based on flags
	var filteredBranches []git.Branch
	for _, branch := range branches {
		if showAll ||
			(showRemote && branch.IsRemote) ||
			(!showRemote && !branch.IsRemote) {
			filteredBranches = append(filteredBranches, branch)
		}
	}

	log.Debug("Filtered branches: %d total", len(filteredBranches))

	if len(filteredBranches) == 0 {
		log.Info("No branches found matching criteria")
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Branch\tCommit\tStatus\tMessage")
	fmt.Fprintln(w, "------\t------\t------\t-------")

	for _, branch := range filteredBranches {
		status := []string{}
		if branch.IsCurrent {
			status = append(status, color.GreenString("current"))
		}
		if branch.IsDefault {
			status = append(status, color.BlueString("default"))
		}
		if branch.IsMerged {
			status = append(status, color.YellowString("merged"))
		}
		if branch.IsStale {
			status = append(status, color.RedString("stale"))
		}

		statusStr := strings.Join(status, ", ")
		if statusStr == "" {
			statusStr = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			branch.Name,
			branch.CommitHash,
			statusStr,
			branch.Message,
		)
	}

	if err := w.Flush(); err != nil {
		log.Error("Failed to flush output: %v", err)
		return err
	}

	log.Debug("Successfully listed branches")
	return nil
}
