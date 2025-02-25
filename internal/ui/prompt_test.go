package ui

import (
	"testing"

	"github.com/bral/git-branch-delete-go/pkg/git"
	"github.com/stretchr/testify/assert"
)

func TestSelectBranches(t *testing.T) {
	// Test empty branches
	selected, err := SelectBranches([]git.Branch{})
	assert.Error(t, err)
	assert.Nil(t, selected)
}

func TestConfirmDeletion(t *testing.T) {
	// Test empty branches
	confirmed, err := ConfirmDeletion([]string{})
	assert.NoError(t, err)
	assert.False(t, confirmed)

	// Since ConfirmDeletion is interactive, we can only test error cases
	// and the empty branches case above
}
