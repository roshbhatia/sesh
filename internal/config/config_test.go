package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BranchFormat != "sy/{{.Session}}/{{.Repo}}" {
		t.Errorf("unexpected BranchFormat: %q", cfg.BranchFormat)
	}
	if !strings.HasSuffix(cfg.SessionsDir, filepath.Join("seshy", "sessions")) {
		t.Errorf("unexpected SessionsDir: %q", cfg.SessionsDir)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BranchFormat != "sy/{{.Session}}/{{.Repo}}" {
		t.Errorf("expected default BranchFormat, got %q", cfg.BranchFormat)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "seshy")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(""), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BranchFormat != "sy/{{.Session}}/{{.Repo}}" {
		t.Errorf("expected default BranchFormat for empty file, got %q", cfg.BranchFormat)
	}
}

func TestLoadPartialConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "seshy")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("branchFormat: \"custom/{{.Repo}}\"\n"), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BranchFormat != "custom/{{.Repo}}" {
		t.Errorf("expected custom BranchFormat, got %q", cfg.BranchFormat)
	}
	// SessionsDir should be default
	if !strings.HasSuffix(cfg.SessionsDir, filepath.Join("seshy", "sessions")) {
		t.Errorf("expected default SessionsDir, got %q", cfg.SessionsDir)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "seshy")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("{{invalid yaml"), 0644)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "config.yaml") {
		t.Errorf("error should mention file path, got: %v", err)
	}
}

func TestConfigDirXDGOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	dir := ConfigDir()
	if dir != "/custom/config/seshy" {
		t.Errorf("expected /custom/config/seshy, got %q", dir)
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	p := ConfigPath()
	if p != "/custom/config/seshy/config.yaml" {
		t.Errorf("expected /custom/config/seshy/config.yaml, got %q", p)
	}
}

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seshy", "config.yaml")
	if err := WriteDefault(path); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "branchFormat") {
		t.Errorf("expected branchFormat in default config, got: %s", data)
	}
}

func TestGetSessionsRootXDGOverride(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/custom-state")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no config file
	root := GetSessionsRoot()
	if root != "/tmp/custom-state/seshy/sessions" {
		t.Errorf("got %q", root)
	}
}

func TestGetSessionsRootDefault(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := GetSessionsRoot()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "state", "seshy", "sessions")
	if root != expected {
		t.Errorf("expected %q, got %q", expected, root)
	}
}
