package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	dir, err := os.MkdirTemp("", "git-test-*")
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	// Initialize git repo
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "--local", "init.defaultBranch", "main"},
		{"git", "config", "--local", "core.autocrlf", "false"},
	}

	// Run initial commands
	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		if err := c.Run(); err != nil {
			cleanup()
			t.Fatalf("Failed to run command %v: %v", cmd, err)
		}
	}

	// Create initial commit and branches
	branchCmds := [][]string{
		// Create initial commit on main
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
		// Create and setup feature branches
		{"git", "branch", "feature/test"},
		{"git", "branch", "feature/test2"},
	}

	for _, cmd := range branchCmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		if err := c.Run(); err != nil {
			cleanup()
			t.Fatalf("Failed to run command %v: %v", cmd, err)
		}
	}

	return dir, cleanup
}

func TestNew(t *testing.T) {
	dir := "/test/dir"
	g := New(dir)
	assert.Equal(t, dir, g.workDir)
}

func TestListBranches(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := New(dir)
	branches, err := g.ListBranches()
	require.NoError(t, err)

	// Should have main and two feature branches
	assert.Len(t, branches, 3, "Expected 3 branches")

	var hasMain, hasFeature1, hasFeature2 bool
	for _, b := range branches {
		switch b.Name {
		case "main":
			hasMain = true
			assert.True(t, b.IsDefault, "main should be marked as default")
		case "feature/test":
			hasFeature1 = true
		case "feature/test2":
			hasFeature2 = true
		}
	}

	assert.True(t, hasMain, "main branch not found")
	assert.True(t, hasFeature1, "feature/test branch not found")
	assert.True(t, hasFeature2, "feature/test2 branch not found")
}

func TestVerifyRepo(t *testing.T) {
	// Test valid repo
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := New(dir)
	err := g.verifyRepo()
	assert.NoError(t, err)

	// Test invalid repo
	invalidDir := filepath.Join(t.TempDir(), "not-a-repo")
	require.NoError(t, os.MkdirAll(invalidDir, 0755))

	g = New(invalidDir)
	err = g.verifyRepo()
	assert.Error(t, err)
	assert.IsType(t, &ErrNotGitRepo{}, err)
}

func TestDeleteBranch(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := New(dir)

	// Try deleting a branch
	err := g.DeleteBranch("feature/test", false, false)
	require.NoError(t, err)

	// Verify branch is gone
	branches, err := g.ListBranches()
	require.NoError(t, err)

	for _, b := range branches {
		assert.NotEqual(t, "feature/test", b.Name)
	}
}

func TestDeleteBranchErrors(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := New(dir)

	tests := []struct {
		name        string
		branchName  string
		force       bool
		remote      bool
		shouldError bool
	}{
		{
			name:        "non-existent branch",
			branchName:  "does-not-exist",
			shouldError: true,
		},
		{
			name:        "delete main branch",
			branchName:  "main",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.DeleteBranch(tt.branchName, tt.force, tt.remote)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
