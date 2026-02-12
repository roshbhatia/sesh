package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSessionsRoot(t *testing.T) {
	// Save original env
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

func TestGetFZFOpts(t *testing.T) {
	originalOpts := os.Getenv("_SESH_FZF_OPTS")
	defer func() {
		if originalOpts != "" {
			os.Setenv("_SESH_FZF_OPTS", originalOpts)
		} else {
			os.Unsetenv("_SESH_FZF_OPTS")
		}
	}()

	t.Run("with _SESH_FZF_OPTS set", func(t *testing.T) {
		os.Setenv("_SESH_FZF_OPTS", "--height 40% --reverse")
		opts := GetFZFOpts()
		if opts != "--height 40% --reverse" {
			t.Errorf("expected '--height 40%% --reverse', got '%s'", opts)
		}
	})

	t.Run("without _SESH_FZF_OPTS", func(t *testing.T) {
		os.Unsetenv("_SESH_FZF_OPTS")
		opts := GetFZFOpts()
		if opts != "" {
			t.Errorf("expected empty string, got '%s'", opts)
		}
	})
}

func TestNew(t *testing.T) {
	cfg := New()
	if cfg == nil {
		t.Fatal("New() returned nil")
	}
	if cfg.SessionsRoot == "" {
		t.Error("SessionsRoot is empty")
	}
}
