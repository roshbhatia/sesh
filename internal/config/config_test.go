package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// GetSessionsRoot
// ---------------------------------------------------------------------------

func TestGetSessionsRootWithXDG(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/custom-state")
	got := GetSessionsRoot()
	want := "/tmp/custom-state/sesh/sessions"
	if got != want {
		t.Errorf("GetSessionsRoot() = %q, want %q", got, want)
	}
}

func TestGetSessionsRootWithoutXDG(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	os.Unsetenv("XDG_STATE_HOME")
	got := GetSessionsRoot()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "state", "sesh", "sessions")
	if got != want {
		t.Errorf("GetSessionsRoot() = %q, want %q", got, want)
	}
}

func TestGetSessionsRootXDGOverridesDefault(t *testing.T) {
	home, _ := os.UserHomeDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	got := GetSessionsRoot()
	want := filepath.Join(home, ".local", "state", "sesh", "sessions")
	if got != want {
		t.Errorf("GetSessionsRoot() = %q, want %q", got, want)
	}
}

func TestGetSessionsRootSuffix(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/a/b")
	got := GetSessionsRoot()
	if filepath.Base(got) != "sessions" {
		t.Errorf("expected last segment 'sessions', got %q", filepath.Base(got))
	}
	if filepath.Base(filepath.Dir(got)) != "sesh" {
		t.Errorf("expected second-to-last segment 'sesh', got %q", filepath.Base(filepath.Dir(got)))
	}
}

// ---------------------------------------------------------------------------
// EnsureSessionsRoot
// ---------------------------------------------------------------------------

func TestEnsureSessionsRootCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	if err := EnsureSessionsRoot(); err != nil {
		t.Fatalf("EnsureSessionsRoot() error: %v", err)
	}

	root := GetSessionsRoot()
	info, err := os.Stat(root)
	if err != nil {
		t.Fatalf("stat sessions root: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected sessions root to be a directory")
	}
}

func TestEnsureSessionsRootIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	// Call twice — should not error
	if err := EnsureSessionsRoot(); err != nil {
		t.Fatalf("first EnsureSessionsRoot: %v", err)
	}
	if err := EnsureSessionsRoot(); err != nil {
		t.Fatalf("second EnsureSessionsRoot: %v", err)
	}
}

func TestEnsureSessionsRootPermissions(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	if err := EnsureSessionsRoot(); err != nil {
		t.Fatalf("EnsureSessionsRoot: %v", err)
	}

	root := GetSessionsRoot()
	info, err := os.Stat(root)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm&0700 != 0700 {
		t.Errorf("expected at least 0700 permissions, got %04o", perm)
	}
}

func TestEnsureSessionsRootNestedPath(t *testing.T) {
	tmp := t.TempDir()
	// Point XDG to a deeply nested path that does not yet exist
	deep := filepath.Join(tmp, "a", "b", "c")
	t.Setenv("XDG_STATE_HOME", deep)

	if err := EnsureSessionsRoot(); err != nil {
		t.Fatalf("EnsureSessionsRoot with nested path: %v", err)
	}

	root := GetSessionsRoot()
	if _, err := os.Stat(root); err != nil {
		t.Errorf("nested sessions root not created: %v", err)
	}
}
