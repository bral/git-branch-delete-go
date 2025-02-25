package ui

import (
	"bytes"
	"testing"

	"github.com/bral/git-branch-delete-go/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleSelectBranches(t *testing.T) {
	tests := []struct {
		name     string
		branches []git.Branch
		input    string
		want     []git.Branch
		wantErr  bool
	}{
		{
			name: "select single branch",
			branches: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test2"},
			},
			input: "1\n",
			want: []git.Branch{
				{Name: "feature/test1"},
			},
		},
		{
			name: "select multiple branches",
			branches: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test2"},
				{Name: "feature/test3"},
			},
			input: "1,3\n",
			want: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test3"},
			},
		},
		{
			name: "select all branches",
			branches: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test2"},
			},
			input: "1,2\n",
			want: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test2"},
			},
		},
		{
			name: "select no branches",
			branches: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test2"},
			},
			input: "\n",
			want:  nil,
		},
		{
			name: "invalid selection",
			branches: []git.Branch{
				{Name: "feature/test1"},
			},
			input:   "2\n",
			wantErr: true,
		},
		{
			name: "invalid format",
			branches: []git.Branch{
				{Name: "feature/test1"},
			},
			input:   "invalid\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewBufferString(tt.input)
			out := &bytes.Buffer{}

			got, err := SimpleSelectBranches(tt.branches, in, out)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSimpleConfirmDeletion(t *testing.T) {
	tests := []struct {
		name     string
		branches []git.Branch
		input    string
		want     bool
		wantErr  bool
	}{
		{
			name: "confirm with y",
			branches: []git.Branch{
				{Name: "feature/test"},
			},
			input: "y\n",
			want:  true,
		},
		{
			name: "confirm with Y",
			branches: []git.Branch{
				{Name: "feature/test"},
			},
			input: "Y\n",
			want:  true,
		},
		{
			name: "deny with n",
			branches: []git.Branch{
				{Name: "feature/test"},
			},
			input: "n\n",
			want:  false,
		},
		{
			name: "deny with N",
			branches: []git.Branch{
				{Name: "feature/test"},
			},
			input: "N\n",
			want:  false,
		},
		{
			name: "invalid input",
			branches: []git.Branch{
				{Name: "feature/test"},
			},
			input:   "invalid\n",
			want:    false,
		},
		{
			name:     "no branches",
			branches: []git.Branch{},
			input:    "y\n",
			want:     false,
		},
		{
			name: "multiple branches",
			branches: []git.Branch{
				{Name: "feature/test1"},
				{Name: "feature/test2"},
			},
			input: "y\n",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewBufferString(tt.input)
			out := &bytes.Buffer{}

			got, err := SimpleConfirmDeletion(tt.branches, in, out)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
