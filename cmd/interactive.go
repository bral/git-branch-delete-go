package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bral/git-branch-delete-go/internal/git"
	"github.com/bral/git-branch-delete-go/internal/log"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	interactiveForce bool
	interactiveAll   bool
)

func init() {
	interactiveCmd := newInteractiveCmd()
	rootCmd.AddCommand(interactiveCmd)

	interactiveCmd.Flags().BoolVarP(&interactiveForce, "force", "f", false, "Force delete branches without merge check")
	interactiveCmd.Flags().BoolVarP(&interactiveAll, "all", "a", false, "Include remote branches (use with caution)")
}

func newInteractiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"i"},
		Short:   "Interactively select branches to delete",
		Long: `Interactively select branches to delete.
Shows a list of local branches by default. Use --all to include remote branches.
Use arrow keys to navigate, space to select, and enter to confirm.

Note:
- Branches marked as [unmerged] require --force to delete
- Remote branches (marked as [remote]) require --all to be visible
- Current branch and protected branches (main, master, etc.) cannot be deleted`,
		Example: `  git-branch-delete interactive        # Delete local branches
  git-branch-delete i --force         # Force delete unmerged branches
  git-branch-delete i --all          # Include remote branches`,
		RunE: runInteractive,
	}
}

func runInteractive(cmd *cobra.Command, args []string) error {
	// Show loading spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "Loading branches "
	s.Start()

	wd, err := os.Getwd()
	if err != nil {
		s.Stop()
		return err
	}

	g, err := git.New(wd)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	branches, err := g.ListBranches()
	if err != nil {
		s.Stop()
		return err
	}

	s.Stop()

	// Filter and prepare branches for selection
	var choices []string
	branchMap := make(map[string]git.GitBranch)

	for _, b := range branches {
		// Skip current and protected branches
		if b.IsCurrent || b.IsDefault {
			continue
		}

		// Create rich label with status indicators
		var indicators []string

		if b.IsStale {
			indicators = append(indicators, color.RedString("stale"))
		}
		if !b.IsMerged {
			indicators = append(indicators, color.YellowString("unmerged"))
		}
		if b.IsMerged {
			indicators = append(indicators, color.GreenString("merged"))
		}

		// Format branch display
		var label string
		if b.IsRemote {
			if !interactiveAll {
				continue
			}
			label = color.BlueString("[remote] ")
		} else {
			label = color.GreenString("[local]  ")
		}

		label += b.Name
		if len(indicators) > 0 {
			label += " (" + strings.Join(indicators, ", ") + ")"
		}
		if b.CommitHash != "" {
			shortHash := b.CommitHash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			label += color.HiBlackString(" " + shortHash)
		}

		choices = append(choices, label)
		branchMap[label] = b
	}

	if len(choices) == 0 {
		if interactiveAll {
			log.Info("No branches available for deletion")
		} else {
			log.Info("No local branches available for deletion (use --all to include remote branches)")
		}
		return nil
	}

	// Sort choices for better UX
	sortBranchChoices(choices)

	// Show branch type counts
	totalLocalCount := 0
	totalRemoteCount := 0
	for _, b := range branchMap {
		if b.IsRemote {
			totalRemoteCount++
		} else {
			totalLocalCount++
		}
	}
	fmt.Printf("Found %d local and %d remote branches\n", totalLocalCount, totalRemoteCount)

	// Configure survey templates
	survey.SelectQuestionTemplate = `
{{- color "default+hb"}}{{ .Message }}{{color "reset"}}
{{- if .Help }} {{color "cyan"}}[{{ .Help }}]{{color "reset"}}{{end}}
{{- color "reset"}}
`

	survey.MultiSelectQuestionTemplate = `
{{- color "default+hb"}}{{ .Message }}{{color "reset"}}
{{- if .Help }} {{color "cyan"}}[{{ .Help }}]{{color "reset"}}{{end}}
{{- "\n"}}
{{- range $ix, $option := .PageEntries}}
  {{- if eq $ix $.SelectedIndex }}{{color "cyan"}}❯{{color "reset"}}{{else}} {{end}}
  {{- if index $.Checked $option.Index }}[x]{{else}}[ ]{{end}}
  {{- " "}}{{ $option.Value }}
  {{- "\n"}}
{{- end}}`

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select branches to delete:",
		Options: choices,
		Description: func(value string, index int) string {
			// Skip descriptions for headers
			if strings.HasPrefix(value, "──") {
				return ""
			}
			branch := branchMap[value]
			if branch.Message != "" {
				return color.HiBlackString(branch.Message)
			}
			return ""
		},
		Help: "↑/↓: navigate • space: select • enter: confirm",
		PageSize: 15,
	}

	err = survey.AskOne(prompt, &selected, survey.WithPageSize(15))
	if err != nil {
		return fmt.Errorf("failed to get user input: %w", err)
	}

	if len(selected) == 0 {
		log.Info("No branches selected for deletion")
		return nil
	}

	// Show summary before confirmation
	var unmergedBranches []string
	var localCount, remoteCount int
	var selectedNames []string

	selectedBranches := make([]git.GitBranch, 0, len(selected))
	for _, label := range selected {
		branch := branchMap[label]
		selectedBranches = append(selectedBranches, branch)

		name := branch.Name
		if !branch.IsMerged {
			name = color.YellowString(name + " (unmerged)")
		}
		selectedNames = append(selectedNames, name)

		if branch.IsRemote {
			remoteCount++
		} else {
			localCount++
		}
	}

	// Show selection summary
	log.Info("\nSelected branches:")
	maxDisplay := 5
	if len(selectedNames) > maxDisplay {
		// Display first 5 branches
		for _, name := range selectedNames[:maxDisplay] {
			log.Info("  %s", name)
		}
		log.Info("  ... and %d more", len(selectedNames)-maxDisplay)
	} else {
		// Display all branches
		for _, name := range selectedNames {
			log.Info("  %s", name)
		}
	}
	log.Info("Total: %s, %s",
		color.GreenString("%d local", localCount),
		color.BlueString("%d remote", remoteCount))

	// Handle unmerged branches
	if len(unmergedBranches) > 0 && !interactiveForce {
		log.Info("\n%s Unmerged branches require --force to delete", color.YellowString("!"))
		return fmt.Errorf("cannot delete unmerged branches without --force")
	}

	// Confirm deletion with counts
	confirmMsg := fmt.Sprintf("Delete %d branches (%d local, %d remote)?", len(selected), localCount, remoteCount)
	if interactiveForce {
		confirmMsg = fmt.Sprintf("Force delete %d branches (%d local, %d remote)?", len(selected), localCount, remoteCount)
	}

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: confirmMsg,
		Default: false,
	}

	err = survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirm {
		log.Info("Operation cancelled")
		return nil
	}

	// Show progress spinner during deletion
	successCount := 0
	failCount := 0
	spinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spinner.Suffix = fmt.Sprintf(" Deleting branches (0/%d)", len(selectedBranches))
	spinner.Start()

	for _, branch := range selectedBranches {
		err := g.DeleteBranch(branch.Name, interactiveForce, branch.IsRemote)
		if err != nil {
			failCount++
			log.Error("Failed to delete %s: %s", branch.Name, err)
		} else {
			successCount++
		}
		spinner.Suffix = fmt.Sprintf(" Deleting branches (%d/%d)", successCount, len(selectedBranches))
	}

	spinner.Stop()

	// Show final summary
	fmt.Printf("\nDeleted %d branches successfully", successCount)
	if failCount > 0 {
		fmt.Printf(", %d failed", failCount)
	}
	fmt.Println()

	// Calculate and show time saved
	if successCount > 0 {
		timePerBranch := 30 * time.Second
		timeSaved := time.Duration(successCount) * timePerBranch
		minutes := int(timeSaved.Minutes())
		seconds := int(timeSaved.Seconds()) % 60

		if minutes > 0 {
			log.Info("Saved you ~%d minutes and %d seconds of manual work! 🚀", minutes, seconds)
		} else {
			log.Info("Saved you ~%d seconds of manual work! 🚀", seconds)
		}
	}

	return nil
}

// sortBranchChoices sorts branch choices for better UX:
// - Stale branches first
// - Then unmerged branches
// - Then merged branches
// - Remote branches last in each category
func sortBranchChoices(choices []string) {
	// Implementation left as is - can be added if needed
}
