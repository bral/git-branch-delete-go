package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/log"
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

	// Filter branches based on flags
	var filteredBranches []git.GitBranch
	for _, b := range branches {
		if !showAll {
			if showRemote && !b.IsRemote {
				continue
			}
			if !showRemote && b.IsRemote {
				continue
			}
		}
		filteredBranches = append(filteredBranches, b)
	}

	if len(filteredBranches) == 0 {
		if showRemote {
			log.Info("No remote branches found")
		} else if showAll {
			log.Info("No branches found")
		} else {
			log.Info("No local branches found")
		}
		return nil
	}

	// Set up tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	log.Info("Found %d branches:", len(filteredBranches))

	for _, branch := range filteredBranches {
		// Format branch name
		prefix := "  "
		if branch.IsCurrent {
			prefix = color.GreenString("* ")
		}
		name := prefix + branch.Name

		// Add indicators
		var indicators []string
		if branch.IsDefault {
			indicators = append(indicators, color.YellowString("default"))
		}
		if branch.IsRemote {
			indicators = append(indicators, color.BlueString("remote"))
		}
		if branch.IsStale {
			indicators = append(indicators, color.RedString("stale"))
		}
		if branch.IsMerged {
			indicators = append(indicators, color.GreenString("merged"))
		}

		status := ""
		if len(indicators) > 0 {
			status = "\t(" + strings.Join(indicators, ", ") + ")"
		}

		// Add commit info if available
		commitInfo := ""
		if branch.CommitHash != "" {
			shortHash := branch.CommitHash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			commitInfo = "\t" + shortHash
			if branch.Message != "" {
				commitInfo += "\t" + branch.Message
			}
		}

		log.Info("%s%s%s", name, status, commitInfo)
	}

	return nil
}
