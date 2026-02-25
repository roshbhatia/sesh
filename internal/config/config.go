package config

import (
	"os"
	"path/filepath"
)

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

// EnsureSessionsRoot creates the sessions root directory if it doesn't exist
func EnsureSessionsRoot() error {
	root := GetSessionsRoot()
	return os.MkdirAll(root, 0755)
}
