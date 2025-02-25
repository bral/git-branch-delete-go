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
		{"git", "checkout", "--orphan", "main"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
		// Create and setup feature branches
		{"git", "branch", "feature/test"},
		{"git", "branch", "feature/test2"},
		// Create unmerged branch
		{"git", "checkout", "-b", "unmerged"},
		{"git", "commit", "--allow-empty", "-m", "Unmerged commit"},
		{"git", "checkout", "main"},
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

	// Count unique branch names
	branchNames := make(map[string]bool)
	for _, b := range branches {
		branchNames[b.Name] = true
	}

	// Should have main, unmerged, and two feature branches
	assert.Len(t, branchNames, 4, "Expected 4 unique branches")

	var hasMain, hasFeature1, hasFeature2, hasUnmerged bool
	for _, b := range branches {
		switch b.Name {
		case "main":
			hasMain = true
			assert.True(t, b.IsDefault, "main should be marked as default")
		case "feature/test":
			hasFeature1 = true
		case "feature/test2":
			hasFeature2 = true
		case "unmerged":
			hasUnmerged = true
			assert.False(t, b.IsMerged, "unmerged branch should be marked as not merged")
		}
	}

	assert.True(t, hasMain, "main branch not found")
	assert.True(t, hasFeature1, "feature/test branch not found")
	assert.True(t, hasFeature2, "feature/test2 branch not found")
	assert.True(t, hasUnmerged, "unmerged branch not found")
}

func TestListRemoteBranches(t *testing.T) {
	t.Skip("Remote branch tests require a remote repository setup")
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
	require.NoError(t, os.MkdirAll(invalidDir, 0o755))

	g = New(invalidDir)
	err = g.verifyRepo()
	assert.Error(t, err)
	assert.IsType(t, &ErrNotGitRepo{}, err)

	// Test inaccessible directory
	inaccessibleDir := filepath.Join(t.TempDir(), "no-access")
	require.NoError(t, os.MkdirAll(inaccessibleDir, 0o000))
	defer os.Chmod(inaccessibleDir, 0o755) // Restore permissions for cleanup

	g = New(inaccessibleDir)
	err = g.verifyRepo()
	assert.Error(t, err)
}

func TestDeleteBranch(t *testing.T) {
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
			name:       "delete local branch",
			branchName: "feature/test",
		},
		{
			name:        "delete unmerged branch without force",
			branchName:  "unmerged",
			shouldError: true,
		},
		{
			name:       "force delete unmerged branch",
			branchName: "unmerged",
			force:      true,
		},
		{
			name:        "delete non-existent branch",
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
				// Verify branch is gone
				branches, err := g.ListBranches()
				require.NoError(t, err)
				for _, b := range branches {
					assert.NotEqual(t, tt.branchName, b.Name)
				}
			}
		})
	}
}

func TestDeleteBranchErrors(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := New(dir)

	// Test deleting branch with invalid characters
	err := g.DeleteBranch("invalid/;branch", false, false)
	assert.Error(t, err)

	// Test deleting protected branch
	err = g.DeleteBranch("main", false, false)
	assert.Error(t, err)

	// Test deleting unmerged branch without force
	err = g.DeleteBranch("unmerged", false, false)
	assert.Error(t, err)

	// Test deleting non-existent branch
	err = g.DeleteBranch("does-not-exist", false, false)
	assert.Error(t, err)
}
