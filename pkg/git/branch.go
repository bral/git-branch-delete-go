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
