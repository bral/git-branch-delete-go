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

type branchSelector struct {
	branches     []git.Branch
	selected     map[int]bool
	currentIndex int
	oldState     *term.State
}

func newBranchSelector(branches []git.Branch, oldState *term.State) *branchSelector {
	return &branchSelector{
		branches:     branches,
		selected:     make(map[int]bool),
		currentIndex: 0,
		oldState:     oldState,
	}
}

func (s *branchSelector) handleInput(b []byte) ([]string, error) {
	switch {
	case len(b) == 1 && b[0] == 3: // Ctrl+C
		return s.handleCtrlC()
	case len(b) == 1 && b[0] == 13: // Enter
		return s.handleEnter()
	case len(b) == 1 && b[0] == 32: // Space
		s.handleSpace()
	case len(b) == 3 && b[0] == 27 && b[1] == 91: // Arrow keys
		s.handleArrowKey(b[2])
	}
	return nil, nil
}

func (s *branchSelector) handleCtrlC() ([]string, error) {
	if err := term.Restore(int(os.Stdin.Fd()), s.oldState); err != nil {
		fmt.Fprintf(os.Stderr, "failed to restore terminal: %v\n", err)
	}
	fmt.Print("\r\n\033[35mExiting without deleting any branches.\033[0m\r\n")
	return nil, nil
}

func (s *branchSelector) handleEnter() ([]string, error) {
	selected := make([]string, 0, len(s.selected))
	for i := range s.branches {
		if s.selected[i] {
			selected = append(selected, s.branches[i].Name)
		}
	}
	if err := term.Restore(int(os.Stdin.Fd()), s.oldState); err != nil {
		fmt.Fprintf(os.Stderr, "failed to restore terminal: %v\n", err)
	}
	fmt.Print("\r\n")
	return selected, nil
}

func (s *branchSelector) handleSpace() {
	s.selected[s.currentIndex] = !s.selected[s.currentIndex]
	if s.currentIndex < len(s.branches)-1 {
		s.currentIndex++
	}
}

func (s *branchSelector) handleArrowKey(key byte) {
	switch key {
	case 65: // Up arrow
		if s.currentIndex > 0 {
			s.currentIndex--
		}
	case 66: // Down arrow
		if s.currentIndex < len(s.branches)-1 {
			s.currentIndex++
		}
	}
}

func (s *branchSelector) render() {
	// Clear screen
	fmt.Print("\033[2J\033[H")
	fmt.Println("\033[36mSelect branches to delete (use arrow keys and space to select, enter to confirm):\033[0m\r")

	for i, branch := range s.branches {
		if i == s.currentIndex {
			fmt.Print("\033[36m> \033[0m") // Highlight current line
		} else {
			fmt.Print("  ")
		}

		if s.selected[i] {
			fmt.Print("\033[32m[x] \033[0m") // Green checkmark
		} else {
			fmt.Print("[ ] ")
		}

		fmt.Printf("%s\r\n", branch.Name)
	}
}

func SelectBranches(branches []git.Branch) ([]string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			fmt.Fprintf(os.Stderr, "failed to restore terminal: %v\n", err)
		}
	}()

	selector := newBranchSelector(branches, oldState)

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	for {
		selector.render()

		b := make([]byte, 3)
		n, err := os.Stdin.Read(b)
		if err != nil {
			if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore terminal: %v\n", err)
			}
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		if selected, err := selector.handleInput(b[:n]); err != nil || selected != nil {
			return selected, err
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

	return strings.EqualFold(result, "y"), nil
}
