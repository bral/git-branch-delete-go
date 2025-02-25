package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/bral/git-branch-delete-go/pkg/git"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

type branchItem struct {
	Name       string
	CommitHash string
	Message    string
	IsCurrent  bool
	IsMerged   bool
	Selected   bool
	IsSpecial  bool // For special options like "Delete All But Current"
}

func (b branchItem) String() string {
	if b.IsSpecial {
		check := " "
		if b.Selected {
			check = "✓"
		}
		return fmt.Sprintf("[%s] \033[1;33m%s\033[0m", check, b.Name)
	}

	check := " "
	if b.Selected {
		check = "✓"
	}
	return fmt.Sprintf("[%s] %s [%s] %s (%s)", check, b.Name, b.CommitHash, b.Message, b.IsMerged)
}

// SelectBranches presents an interactive prompt for selecting branches to delete
func SelectBranches(branches []git.Branch) ([]string, error) {
	var current *git.Branch
	var others []git.Branch

	// Separate current branch from others
	for i, b := range branches {
		if b.IsCurrent {
			current = &branches[i]
		} else {
			others = append(others, b)
		}
	}

	// Put terminal in raw mode for the entire selection process
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("could not set terminal to raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Clear screen and hide cursor
	fmt.Print("\033[H\033[2J\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor when done

	if current != nil {
		mergeStatus := "(not merged)"
		if current.IsMerged {
			mergeStatus = "(merged)"
		}
		fmt.Printf("\033[33mCurrent branch: %s [%s] %s %s\033[0m\r\n\r\n",
			current.Name,
			current.CommitHash,
			current.Message,
			mergeStatus,
		)
	}

	if len(others) == 0 {
		fmt.Print("\r\n")
		return nil, fmt.Errorf("no branches available for deletion")
	}

	// Create items list with special "Delete All But Current" option
	items := make([]branchItem, len(others)+1)
	items[0] = branchItem{
		Name:      "Delete All But Current Branch",
		IsSpecial: true,
		Selected:  false,
	}

	// Add regular branch items
	for i, b := range others {
		items[i+1] = branchItem{
			Name:       b.Name,
			CommitHash: b.CommitHash,
			Message:    b.Message,
			IsCurrent:  b.IsCurrent,
			IsMerged:   b.IsMerged,
			Selected:   false,
		}
	}

	currentIndex := 0
	var selected []string

	// Initial prompt
	fmt.Print("SPACE to select/unselect, ENTER to confirm\r\n")
	fmt.Print("Press Ctrl+C to exit without deleting\r\n\r\n")

	// Get terminal height
	_, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("could not get terminal size: %v", err)
	}

	// Calculate visible items
	headerHeight := 4 // Basic prompt height
	if current != nil {
		headerHeight += 2 // Add current branch height
	}
	visibleItems := height - headerHeight - 1 // -1 for prompt at bottom
	if visibleItems < 1 {
		visibleItems = 1
	}

	// Track scroll position
	scrollPos := 0

	for {
		// Calculate visible range
		start := scrollPos
		end := scrollPos + visibleItems
		if end > len(items) {
			end = len(items)
			start = end - visibleItems
			if start < 0 {
				start = 0
			}
		}

		// Clear items area only (leave prompt intact)
		fmt.Print("\033[" + fmt.Sprint(headerHeight) + ";1H") // Move to start of items
		fmt.Print("\033[J")                                   // Clear from cursor to end

		// Draw visible items
		var buf strings.Builder
		for i := start; i < end; i++ {
			item := items[i]
			if item.Selected {
				buf.WriteString("\033[32m✓\033[0m ") // Green checkmark for selected
			} else {
				buf.WriteString("\033[90m✓\033[0m ") // Gray checkmark for unselected
			}

			if item.IsSpecial {
				if i == currentIndex {
					buf.WriteString(fmt.Sprintf("\033[4;33m%s\033[0m\r\n", item.Name)) // Underlined yellow
					// Add separator after special option
					buf.WriteString("\033[90m" + strings.Repeat("─", 50) + "\033[0m\r\n")
				} else {
					buf.WriteString(fmt.Sprintf("\033[33m%s\033[0m\r\n", item.Name)) // Yellow
					// Add separator after special option
					buf.WriteString("\033[90m" + strings.Repeat("─", 50) + "\033[0m\r\n")
				}
				continue
			}

			// Format branch info
			mergeStatus := "\033[31m(not merged)\033[0m"
			if item.IsMerged {
				mergeStatus = "\033[32m(merged)\033[0m"
			}

			if i == currentIndex {
				// Current item gets cyan underline
				buf.WriteString(fmt.Sprintf("\033[4;36m%s\033[0m [%s] %s %s\r\n",
					item.Name,
					item.CommitHash,
					item.Message,
					mergeStatus,
				))
			} else {
				if item.Selected {
					// Selected items are cyan
					buf.WriteString(fmt.Sprintf("\033[36m%s\033[0m [%s] %s %s\r\n",
						item.Name,
						item.CommitHash,
						item.Message,
						mergeStatus,
					))
				} else {
					buf.WriteString(fmt.Sprintf("%s [%s] %s %s\r\n",
						item.Name,
						item.CommitHash,
						item.Message,
						mergeStatus,
					))
				}
			}
		}

		// Write buffer to terminal
		fmt.Print(buf.String())

		// Read input
		b := make([]byte, 3) // Buffer for escape sequences
		n, err := os.Stdin.Read(b)
		if err != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
			return nil, fmt.Errorf("error reading input: %v", err)
		}

		switch {
		case n == 1 && b[0] == 3: // Ctrl+C
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Print("\r\n\033[35mExiting without deleting any branches.\033[0m\r\n")
			return nil, nil
		case n == 1 && b[0] == 13: // Enter
			// Collect selected items
			for _, item := range items {
				if item.Selected && !item.IsSpecial {
					selected = append(selected, item.Name)
				}
			}
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Print("\r\n")
			return selected, nil
		case n == 1 && b[0] == 32: // Space
			if currentIndex == 0 { // Special "Delete All But Current" option
				// Toggle all items except current branch
				items[0].Selected = !items[0].Selected
				for i := 1; i < len(items); i++ {
					items[i].Selected = items[0].Selected
				}
			} else {
				// Toggle selection of current item
				items[currentIndex].Selected = !items[currentIndex].Selected
				// If any regular item is deselected, deselect the "Delete All" option
				if !items[currentIndex].Selected {
					items[0].Selected = false
				}
			}
		case n == 1 && b[0] == 14: // Ctrl+N (next)
			if currentIndex < len(items)-1 {
				currentIndex++
			}
		case n == 1 && b[0] == 16: // Ctrl+P (previous)
			if currentIndex > 0 {
				currentIndex--
			}
		case n == 3 && b[0] == 27 && b[1] == 91: // Arrow keys
			switch b[2] {
			case 65: // Up arrow (27,91,65)
				if currentIndex > 0 {
					currentIndex--
					// Adjust scroll position if needed
					if currentIndex < scrollPos {
						scrollPos = currentIndex
					}
				}
			case 66: // Down arrow (27,91,66)
				if currentIndex < len(items)-1 {
					currentIndex++
					// Adjust scroll position if needed
					if currentIndex >= scrollPos+visibleItems {
						scrollPos = currentIndex - visibleItems + 1
					}
				}
			}
		}
	}
}

// ConfirmDeletion asks for confirmation before deleting branches
func ConfirmDeletion(branches []string) (bool, error) {
	if len(branches) == 0 {
		return false, nil
	}

	color.Red("You have selected these branches to delete:")
	for i, name := range branches {
		fmt.Printf(" %d. %s\n", i+1, name)
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Delete these %d branches", len(branches)),
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return false, nil
	}

	return strings.ToLower(result) == "y", nil
}
