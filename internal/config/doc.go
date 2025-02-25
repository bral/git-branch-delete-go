/*
Package config provides configuration management for the git-branch-delete tool.

This package handles loading and managing configuration from multiple sources:
  - Configuration files (YAML)
  - Environment variables
  - Command-line flags

Configuration File:

The default configuration file is located at ~/.config/git-branch-delete.yaml:

	default_branch: main
	protected_branches:
	  - main
	  - master
	  - develop
	default_remote: origin
	auto_confirm: false
	dry_run: false

Environment Variables:

All configuration options can be set via environment variables with the GBD_ prefix:

	GBD_DEFAULT_BRANCH=main
	GBD_DEFAULT_REMOTE=origin
	GBD_AUTO_CONFIRM=true
	GBD_DRY_RUN=true

Usage:

Basic usage of the config package:

	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Access configuration
	if config.DryRun {
		fmt.Println("Running in dry-run mode")
	}

	// Check protected branches
	for _, branch := range config.ProtectedBranches {
		fmt.Printf("Protected branch: %s\n", branch)
	}

Configuration Priority:

The configuration is loaded with the following priority (highest to lowest):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

Default Values:

If no configuration is provided, the following defaults are used:
  - DefaultRemote: "origin"
  - ProtectedBranches: ["main", "master"]
  - AutoConfirm: false
  - DryRun: false
*/
package config
