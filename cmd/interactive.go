package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
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

// Add constants for better maintainability
const (
	maxDisplayBranches = 5
	timePerBranchDelete = 30 * time.Second
	maxBranchesWarningThreshold = 10
	spinnerUpdateInterval = 100 * time.Millisecond
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
	// Validate no args were provided
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %v", args)
	}

	// Show loading spinner
	s := spinner.New(spinner.CharSets[14], spinnerUpdateInterval)
	s.Prefix = "Loading branches "
	s.Start()
	defer s.Stop() // Ensure spinner stops even on error

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Initialize git with cleanup
	g, err := git.New(wd)
	if err != nil {
		return fmt.Errorf("failed to initialize git in %s: %w", wd, err)
	}

	// List branches with proper error context
	branches, err := g.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	s.Stop()

	// Pre-allocate slices with expected capacity
	choices := make([]string, 0, len(branches))
	branchMap := make(map[string]git.GitBranch, len(branches))

	// First find and display current branch
	var currentBranch string
	for _, b := range branches {
		if b.IsCurrent {
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

			currentBranch = fmt.Sprintf("%s %s%s",
				color.CyanString("*"),
				color.HiWhiteString(b.Name),
				func() string {
					if len(indicators) > 0 {
						return " (" + strings.Join(indicators, ", ") + ")"
					}
					return ""
				}(),
			)
			break
		}
	}

	// Then process other branches
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

	// Show branch type counts and current branch
	totalLocalCount := 0
	totalRemoteCount := 0
	for _, b := range branchMap {
		if b.IsRemote {
			totalRemoteCount++
		} else {
			totalLocalCount++
		}
	}
	fmt.Printf("\n%s\n", color.HiBlackString("─── Current Branch ───────────────────────"))
	fmt.Printf("  %s\n", currentBranch)
	fmt.Printf("\n")
	fmt.Printf("%s\n", color.HiBlackString("─── Available Branches ────────────────────"))
	fmt.Printf("Found %d local and %d remote branches\n", totalLocalCount, totalRemoteCount)
	fmt.Printf("\n")

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
  {{- if index $.Checked $option.Index }}{{color "green"}}✓{{color "reset"}}{{else}}{{color "default"}}○{{color "reset"}}{{end}}
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
		// The survey package has built-in filtering that can't be fully disabled.
		// This is a workaround that preserves all options by always returning true,
		// effectively neutralizing the filtering behavior while maintaining the
		// selection state. This prevents the issue where typing would cause
		// selections to disappear.
		Filter: func(filter string, value string, index int) bool {
			return true
		},
	}

	err = survey.AskOne(prompt, &selected, survey.WithPageSize(15))
	if err != nil {
		if err == terminal.InterruptErr {
			log.Info("Operation cancelled by user")
			return nil
		}
		return fmt.Errorf("failed to get branch selection: %w", err)
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
	fmt.Printf("\nSelected branches:\n\n")
	maxDisplay := 5
	if len(selectedNames) > maxDisplay {
		// Display first 5 branches
		for i, name := range selectedNames[:maxDisplay] {
			branch := selectedBranches[i]
			indicator := color.GreenString("[local]")
			if branch.IsRemote {
				indicator = color.BlueString("[remote]")
			}
			fmt.Printf("  %s %s %s%s\n", color.GreenString("✓"), indicator, name, formatCommitHash(branch.CommitHash))
		}
		fmt.Printf("  ... and %d more\n", len(selectedNames)-maxDisplay)
	} else {
		// Display all branches
		for i, name := range selectedNames {
			branch := selectedBranches[i]
			indicator := color.GreenString("[local]")
			if branch.IsRemote {
				indicator = color.BlueString("[remote]")
			}
			fmt.Printf("  %s %s %s%s\n", color.GreenString("✓"), indicator, name, formatCommitHash(branch.CommitHash))
		}
	}
	fmt.Printf("\nTotal: %s, %s\n",
		color.GreenString("%d local", localCount),
		color.BlueString("%d remote", remoteCount))

	// Handle unmerged branches
	if len(unmergedBranches) > 0 && !interactiveForce {
		log.Info("\n%s Unmerged branches require --force to delete", color.YellowString("!"))
		return fmt.Errorf("cannot delete unmerged branches without --force")
	}

	// Safety check: don't allow deleting all branches
	if len(selectedBranches) >= len(branches)-1 {
		log.Warn("Cannot delete all branches, at least one branch must remain")
		return fmt.Errorf("refusing to delete all branches")
	}

	// Safety check: warn about large deletions
	if len(selectedBranches) > 10 {
		log.Warn("You are about to delete %d branches. This is a large operation.", len(selectedBranches))
		var proceed bool
		proceedPrompt := &survey.Confirm{
			Message: "Are you sure you want to proceed?",
			Default: false,
		}
		if err := survey.AskOne(proceedPrompt, &proceed); err != nil || !proceed {
			log.Info("Operation cancelled")
			return nil
		}
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
		if err == terminal.InterruptErr {
			log.Info("Operation cancelled by user")
			return nil
		}
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

	// Use a buffered channel for parallel branch deletion with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type deleteResult struct {
		branch string
		err    error
	}
	results := make(chan deleteResult, len(selectedBranches))

	// Process branches in parallel with a worker pool
	const maxWorkers = 4
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, branch := range selectedBranches {
		wg.Add(1)
		go func(b git.GitBranch) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			select {
			case sem <- struct{}{}: // Acquire semaphore
				err := g.DeleteBranch(b.Name, interactiveForce, b.IsRemote)
				results <- deleteResult{branch: b.Name, err: err}
			case <-ctx.Done():
				results <- deleteResult{branch: b.Name, err: ctx.Err()}
			}
		}(branch)
	}

	// Wait for all workers in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with timeout
	var errs []string
loop:
	for {
		select {
		case result, ok := <-results:
			if !ok {
				break loop
			}
			if result.err != nil {
				failCount++
				errs = append(errs, fmt.Sprintf("%s: %s", result.branch, result.err))
			} else {
				successCount++
			}
			spinner.Suffix = fmt.Sprintf(" Deleting branches (%d/%d)", successCount+failCount, len(selectedBranches))
		case <-ctx.Done():
			log.Error("Operation timed out after 30 seconds")
			return ctx.Err()
		}
	}

	spinner.Stop()

	// Show final summary with detailed errors if any
	fmt.Printf("\nDeleted %d branches successfully", successCount)
	if failCount > 0 {
		fmt.Printf(", %d failed", failCount)
		fmt.Println("\nFailed branches:")
		for _, err := range errs {
			fmt.Printf("  - %s\n", err)
		}
	}
	fmt.Println()

	// Calculate and show time saved
	if successCount > 0 {
		timeSaved := time.Duration(successCount) * timePerBranchDelete
		minutes := int(timeSaved.Minutes())
		seconds := int(timeSaved.Seconds()) % 60

		if minutes > 0 {
			fmt.Printf("Saved you ~%d minutes and %d seconds of manual work! 🚀\n", minutes, seconds)
		} else {
			fmt.Printf("Saved you ~%d seconds of manual work! 🚀\n", seconds)
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
	type branchScore struct {
		index int
		score int
		value string
	}

	scores := make([]branchScore, len(choices))
	for i, choice := range choices {
		score := 0

		// Priority scoring (higher is more important)
		switch {
		case strings.Contains(choice, color.RedString("stale")):
			score += 8000
		case strings.Contains(choice, color.YellowString("unmerged")):
			score += 4000
		case strings.Contains(choice, color.GreenString("merged")):
			score += 2000
		}

		// Deprioritize remote branches within their categories
		if strings.Contains(choice, color.BlueString("[remote]")) {
			score -= 1000
		}

		// Use original index as tiebreaker for stable sort
		score = score*10000 + (10000 - i)

		scores[i] = branchScore{
			index: i,
			score: score,
			value: choice,
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Update choices array in place
	sorted := make([]string, len(choices))
	for i, s := range scores {
		sorted[i] = s.value
	}
	copy(choices, sorted)
}

// Add helper function at the end of the file
func formatCommitHash(hash string) string {
	if hash == "" {
		return ""
	}
	if len(hash) > 7 {
		hash = hash[:7]
	}
	return color.HiBlackString(" " + hash)
}
