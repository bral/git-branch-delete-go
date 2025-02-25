package git

// Branch represents a git branch
type Branch struct {
	Name       string
	CommitHash string
	Message    string
	IsLocal    bool
	IsRemote   bool
	IsDefault  bool
	IsCurrent  bool
	IsStale    bool
	IsMerged   bool
}

// BranchDeletionResult represents the result of a branch deletion operation
type BranchDeletionResult struct {
	Name    string
	Success bool
	Error   string
}
