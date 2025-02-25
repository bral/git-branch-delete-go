package ui

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/bral/git-branch-delete-go/pkg/git"
)

// SimpleSelectBranches presents a list of branches to the user and returns their selection
func SimpleSelectBranches(branches []git.Branch, in io.Reader, out io.Writer) ([]git.Branch, error) {
	if len(branches) == 0 {
		return nil, nil
	}

	// Display branches
	fmt.Fprintln(out, "Select branches to delete (comma-separated numbers, or empty to cancel):")
	for i, b := range branches {
		status := ""
		if b.IsDefault {
			status = "(default)"
		} else if !b.IsMerged {
			status = "(not merged)"
		}
		fmt.Fprintf(out, "[%d] %s %s\n", i+1, b.Name, status)
	}

	// Read selection
	reader := bufio.NewReader(in)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	// Parse selection
	selections := strings.Split(input, ",")
	selected := make([]git.Branch, 0, len(selections))
	for _, s := range selections {
		s = strings.TrimSpace(s)
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid selection: %s", s)
		}
		if i < 1 || i > len(branches) {
			return nil, fmt.Errorf("invalid selection: %d", i)
		}
		selected = append(selected, branches[i-1])
	}

	return selected, nil
}

// SimpleConfirmDeletion asks the user to confirm deletion of the specified branches
func SimpleConfirmDeletion(branches []git.Branch, in io.Reader, out io.Writer) (bool, error) {
	if len(branches) == 0 {
		return false, nil
	}

	fmt.Fprintln(out, "\nThe following branches will be deleted:")
	for _, b := range branches {
		fmt.Fprintf(out, "  %s\n", b.Name)
	}
	fmt.Fprint(out, "\nAre you sure? [y/N] ")

	reader := bufio.NewReader(in)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	if input != "y" && input != "yes" {
		return false, nil
	}

	return true, nil
}
