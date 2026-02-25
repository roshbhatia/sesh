package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSessionsRoot(t *testing.T) {
	originalXDG := os.Getenv("XDG_STATE_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_STATE_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_STATE_HOME")
		}
	}()

	t.Run("with XDG_STATE_HOME set", func(t *testing.T) {
		os.Setenv("XDG_STATE_HOME", "/tmp/test-state")
		root := GetSessionsRoot()
		expected := "/tmp/test-state/sesh/sessions"
		if root != expected {
			t.Errorf("expected %s, got %s", expected, root)
		}
	})

	t.Run("without XDG_STATE_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_STATE_HOME")
		root := GetSessionsRoot()
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".local", "state", "sesh", "sessions")
		if root != expected {
			t.Errorf("expected %s, got %s", expected, root)
		}
	})
}
