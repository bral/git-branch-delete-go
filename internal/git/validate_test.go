package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGitArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"valid command", "branch", false},
		{"valid flag", "--format", false},
		{"valid ref", "refs/heads/main", false},
		{"valid format", "%(refname)", false},
		{"valid branch name", "feature/test-123", false},
		{"empty string", "", false},
		{"command injection ;", "branch;ls", true},
		{"command injection &&", "branch&&ls", true},
		{"command injection |", "branch|ls", true},
		{"command injection `", "branch`ls`", true},
		{"command injection $", "branch$PATH", true},
		{"invalid characters", "branch\n", true},
		{"path traversal", "../config", true},
		{"unknown flag", "--unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitArg(tt.arg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{"valid simple", "main", false},
		{"valid with slash", "feature/test", false},
		{"valid with dash", "fix-123", false},
		{"valid with underscore", "feature_test", false},
		{"empty", "", true},
		{"too long", string(make([]byte, 256)), true},
		{"starts with dot", ".hidden", true},
		{"ends with dot", "branch.", true},
		{"ends with slash", "branch/", true},
		{"ends with .lock", "branch.lock", true},
		{"contains space", "feature branch", true},
		{"contains special chars", "feature*test", true},
		{"contains control chars", "feature\ntest", true},
		{"double dots", "feature..test", true},
		{"command injection", "feature;ls", true},
		{"path traversal", "../config", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.branch)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCustomErrors(t *testing.T) {
	t.Run("ErrInvalidBranch", func(t *testing.T) {
		err := newInvalidBranchError("test", "invalid chars")
		assert.EqualError(t, err, "invalid branch 'test': invalid chars")
	})

	t.Run("ErrProtectedBranch", func(t *testing.T) {
		err := newProtectedBranchError("main")
		assert.EqualError(t, err, "cannot modify protected branch 'main'")
	})

	t.Run("ErrUnmergedBranch", func(t *testing.T) {
		err := newUnmergedBranchError("feature")
		assert.EqualError(t, err, "branch 'feature' is not fully merged")
	})

	t.Run("ErrGitCommand", func(t *testing.T) {
		err := newGitCommandError("status", "fatal: not a git repository", assert.AnError)
		assert.Contains(t, err.Error(), "git command 'status' failed")
		assert.Contains(t, err.Error(), "fatal: not a git repository")
	})

	t.Run("ErrTimeout", func(t *testing.T) {
		err := newTimeoutError("fetch", "30s")
		assert.EqualError(t, err, "git command 'fetch' timed out after 30s")
	})
}
