package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config holds the application configuration
type Config struct {
	DefaultBranch     string   `json:"defaultBranch"`
	ProtectedBranches []string `json:"protectedBranches"`
	DefaultRemote     string   `json:"defaultRemote"`
	AutoConfirm       bool     `json:"autoConfirm"`
	MaxBranchLength   int      `json:"maxBranchLength"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultBranch:     "main",
		ProtectedBranches: []string{"main", "master", "develop"},
		DefaultRemote:     "origin",
		AutoConfirm:       false,
		MaxBranchLength:   255, // Git's limit
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate protected branches
	for _, branch := range c.ProtectedBranches {
		if strings.TrimSpace(branch) == "" {
			return fmt.Errorf("protected branch name cannot be empty")
		}
		if len(branch) > c.MaxBranchLength {
			return fmt.Errorf("protected branch name too long: %s", branch)
		}
	}

	// Validate remote name
	if strings.TrimSpace(c.DefaultRemote) == "" {
		return fmt.Errorf("default remote cannot be empty")
	}
	if strings.ContainsAny(c.DefaultRemote, " \t\n\r") {
		return fmt.Errorf("default remote contains invalid characters")
	}

	// Validate max branch length
	if c.MaxBranchLength <= 0 || c.MaxBranchLength > 255 {
		return fmt.Errorf("invalid max branch length: %d", c.MaxBranchLength)
	}

	return nil
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	// Check if config file exists
	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Check file permissions
	if info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("config file has too-broad permissions %#o (should be 0600)", info.Mode().Perm())
	}

	// Open file with restricted permissions
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0o600)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to open config: %w", err)
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to decode config: %w", err)
	}

	// Validate loaded config
	if err := config.Validate(); err != nil {
		return DefaultConfig(), fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	// Validate before saving
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create temporary file for atomic write
	tmpFile, err := os.CreateTemp(configDir, "config.*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up in case of error

	// Set restrictive permissions
	if err := tmpFile.Chmod(0o600); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Write config with indentation for readability
	enc := json.NewEncoder(tmpFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Ensure all data is written
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync file: %w", err)
	}
	tmpFile.Close()

	// Atomic rename
	if err := os.Rename(tmpPath, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Ensure the path is clean and absolute
	configDir = filepath.Clean(configDir)
	if !filepath.IsAbs(configDir) {
		return "", fmt.Errorf("config directory path must be absolute")
	}

	// Use platform-specific paths
	var configPath string
	switch runtime.GOOS {
	case "windows":
		configPath = filepath.Join(configDir, "git-branch-delete", "config.json")
	default:
		configPath = filepath.Join(configDir, ".git-branch-delete", "config.json")
	}

	// Verify the final path is still under the config directory
	if !strings.HasPrefix(filepath.Clean(configPath), configDir) {
		return "", fmt.Errorf("invalid config path")
	}

	return configPath, nil
}
