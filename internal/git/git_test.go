package git

import (
	"fmt"
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
		{"git", "config", "init.defaultBranch", "main"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
		{"git", "branch", "feature/test"},
		{"git", "branch", "feature/test2"},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		require.NoError(t, c.Run())
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
	assert.Len(t, branches, 3)

	var hasMain, hasFeature1, hasFeature2 bool
	for _, b := range branches {
		switch b.Name {
		case "main":
			hasMain = true
			assert.True(t, b.IsDefault)
		case "feature/test":
			hasFeature1 = true
		case "feature/test2":
			hasFeature2 = true
		}
	}

	assert.True(t, hasMain)
	assert.True(t, hasFeature1)
	assert.True(t, hasFeature2)
}

func TestVerifyRepo(t *testing.T) {
	// Test valid repo
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	g := New(dir)
	err := g.verifyRepo()
	assert.NoError(t, err)

	// Test invalid repo
	invalidDir := filepath.Join(dir, "not-a-repo")
	require.NoError(t, os.Mkdir(invalidDir, 0755))

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

func setupBenchmarkRepo(b *testing.B) (string, func()) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "git-bench-*")
	require.NoError(b, err)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	// Initialize git repo with many branches
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "init.defaultBranch", "main"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		require.NoError(b, c.Run())
	}

	// Create many branches
	for i := 0; i < 100; i++ {
		cmd := exec.Command("git", "branch", fmt.Sprintf("feature/test-%d", i))
		cmd.Dir = dir
		require.NoError(b, cmd.Run())
	}

	return dir, cleanup
}

func BenchmarkListBranches(b *testing.B) {
	dir, cleanup := setupBenchmarkRepo(b)
	defer cleanup()

	g := New(dir)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		branches, err := g.ListBranches()
		require.NoError(b, err)
		require.NotEmpty(b, branches)
	}
}

func BenchmarkGetCurrentBranch(b *testing.B) {
	dir, cleanup := setupBenchmarkRepo(b)
	defer cleanup()

	g := New(dir)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		branch, err := g.getCurrentBranch()
		require.NoError(b, err)
		require.NotEmpty(b, branch)
	}
}

func BenchmarkGetDefaultBranch(b *testing.B) {
	dir, cleanup := setupBenchmarkRepo(b)
	defer cleanup()

	g := New(dir)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		branch, err := g.getDefaultBranch()
		require.NoError(b, err)
		require.NotEmpty(b, branch)
	}
}

func BenchmarkMarkStaleBranches(b *testing.B) {
	dir, cleanup := setupBenchmarkRepo(b)
	defer cleanup()

	g := New(dir)
	branches, err := g.ListBranches()
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := g.markStaleBranches(branches)
		require.NoError(b, err)
	}
}

func BenchmarkDeleteBranch(b *testing.B) {
	dir, cleanup := setupBenchmarkRepo(b)
	defer cleanup()

	g := New(dir)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create a new branch for each iteration
		branchName := fmt.Sprintf("bench-branch-%d", i)
		cmd := exec.Command("git", "branch", branchName)
		cmd.Dir = dir
		require.NoError(b, cmd.Run())
		b.StartTimer()

		err := g.DeleteBranch(branchName, true, false)
		require.NoError(b, err)
	}
}
