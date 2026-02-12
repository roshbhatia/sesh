package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	SessionsRoot string
	FZFOpts      string
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		SessionsRoot: GetSessionsRoot(),
		FZFOpts:      GetFZFOpts(),
	}
}

// GetSessionsRoot returns the sessions root directory
func GetSessionsRoot() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "sesh", "sessions")
}

// GetFZFOpts returns custom fzf options from environment
func GetFZFOpts() string {
	return os.Getenv("_SESH_FZF_OPTS")
}

// EnsureSessionsRoot creates the sessions root directory if it doesn't exist
func EnsureSessionsRoot() error {
	root := GetSessionsRoot()
	return os.MkdirAll(root, 0755)
}
