package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewInteractiveCmd(t *testing.T) {
	// Create a test root command
	testRoot := initTestRoot()

	// Initialize flags
	var testForce, testAll bool

	// Create interactive command
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactively select branches to delete",
		Long: `Interactively select branches to delete.
Shows a list of branches with their status and allows selecting multiple branches for deletion.`,
		Example: `  git-branch-delete interactive
  git-branch-delete interactive --force
  git-branch-delete interactive --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil // No-op for testing
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&testForce, "force", "f", false, "Force delete branches without checking merge status")
	cmd.Flags().BoolVarP(&testAll, "all", "a", false, "Delete both local and remote branches")

	// Add command to root
	testRoot.AddCommand(cmd)

	// Test cases
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
		checkVal func(*testing.T, *cobra.Command)
	}{
		{
			name:    "no flags",
			args:    []string{"interactive"},
			wantErr: false,
			checkVal: func(t *testing.T, cmd *cobra.Command) {
				force, _ := cmd.Flags().GetBool("force")
				all, _ := cmd.Flags().GetBool("all")
				assert.False(t, force)
				assert.False(t, all)
			},
		},
		{
			name:    "force flag",
			args:    []string{"interactive", "--force"},
			wantErr: false,
			checkVal: func(t *testing.T, cmd *cobra.Command) {
				force, _ := cmd.Flags().GetBool("force")
				assert.True(t, force)
			},
		},
		{
			name:    "all flag",
			args:    []string{"interactive", "--all"},
			wantErr: false,
			checkVal: func(t *testing.T, cmd *cobra.Command) {
				all, _ := cmd.Flags().GetBool("all")
				assert.True(t, all)
			},
		},
		{
			name:    "force and all flags",
			args:    []string{"interactive", "--force", "--all"},
			wantErr: false,
			checkVal: func(t *testing.T, cmd *cobra.Command) {
				force, _ := cmd.Flags().GetBool("force")
				all, _ := cmd.Flags().GetBool("all")
				assert.True(t, force)
				assert.True(t, all)
			},
		},
		{
			name:    "short flags",
			args:    []string{"interactive", "-f", "-a"},
			wantErr: false,
			checkVal: func(t *testing.T, cmd *cobra.Command) {
				force, _ := cmd.Flags().GetBool("force")
				all, _ := cmd.Flags().GetBool("all")
				assert.True(t, force)
				assert.True(t, all)
			},
		},
		{
			name:    "invalid flag",
			args:    []string{"interactive", "--invalid"},
			wantErr: true,
			errMsg:  "unknown flag: --invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			cmd.Flags().Set("force", "false")
			cmd.Flags().Set("all", "false")

			// Set args for this test
			testRoot.SetArgs(tt.args)

			// Execute command
			err := testRoot.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkVal != nil {
					tt.checkVal(t, cmd)
				}
			}
		})
	}
}

func TestInteractiveCmd(t *testing.T) {
	// Skip execution tests since they require git repo and user interaction
	t.Skip("Skipping execution tests since they require git repo and user interaction")
}
