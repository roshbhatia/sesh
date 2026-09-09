// Package config handles seshy configuration via YAML files with XDG support.
package config

import (
	"os"
	"path/filepath"

	sharedconfig "github.com/roshbhatia/go-utils/config"
	"github.com/roshbhatia/go-utils/paths"
	"go.yaml.in/yaml/v3"
)

// HooksConfig holds lifecycle hook commands.
type HooksConfig struct {
	PostCreate []string `json:"postCreate,omitempty" yaml:"postCreate"`
	PostAdd    []string `json:"postAdd,omitempty" yaml:"postAdd"`
	PreDelete  []string `json:"preDelete,omitempty" yaml:"preDelete"`
}

// Config holds all seshy configuration.
type Config struct {
	BranchFormat  string      `json:"branchFormat,omitempty" yaml:"branchFormat"`
	SessionsDir   string      `json:"sessionsDir,omitempty" yaml:"sessionsDir"`
	ArchiveDir    string      `json:"archiveDir,omitempty" yaml:"archiveDir"`
	RepoSource    string      `json:"repoSource,omitempty" yaml:"repoSource"`
	Picker        string      `json:"picker,omitempty" yaml:"picker"`
	SessionPicker string      `json:"sessionPicker,omitempty" yaml:"sessionPicker"`
	DefaultRepos  []string    `json:"defaultRepos,omitempty" yaml:"defaultRepos"`
	Hooks         HooksConfig `json:"hooks,omitempty" yaml:"hooks"`
}

func defaults() Config {
	return Config{
		BranchFormat:  "sy/{{.Session}}/{{.Repo}}",
		SessionsDir:   "",
		ArchiveDir:    "",
		RepoSource:    "zoxide query --list",
		Picker:        "fzf --multi --height=40% --reverse --prompt='repo > '",
		SessionPicker: "fzf --height=40% --reverse --prompt='session > '",
	}
}

// ConfigDir returns the seshy config directory.
func ConfigDir() string {
	return filepath.Join(paths.ConfigHome(), "seshy")
}

// ConfigPath returns the path to config.yaml.
func ConfigPath() string {
	path, err := sharedconfig.Path(configOptions())
	if err != nil {
		return filepath.Join(ConfigDir(), "config.yaml")
	}
	return path
}

func configOptions() sharedconfig.Options {
	return sharedconfig.Options{Name: "seshy", EnvPrefix: "SESHY"}
}

// Load reads config from disk and merges with defaults.
func Load() (*Config, error) {
	cfg, err := sharedconfig.Load(defaults(), configOptions())
	if err != nil {
		return nil, err
	}

	for i, p := range cfg.DefaultRepos {
		cfg.DefaultRepos[i] = paths.ExpandHome(p)
	}

	return &cfg, nil
}

// Schema emits the JSON Schema used by YAML language servers.
func Schema() ([]byte, error) {
	return sharedconfig.Schema[Config]("Seshy configuration")
}

// WriteDefault writes a default config file.
func WriteDefault() error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	cfg := defaults()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetSessionsRoot returns the sessions directory. config.yaml wins, then the
// sysinit paths manifest, which is the value nix wrote into config.yaml anyway.
func GetSessionsRoot() string {
	if cfg, err := Load(); err == nil && cfg.SessionsDir != "" {
		return paths.ExpandHome(cfg.SessionsDir)
	}
	return paths.SeshySessions()
}

// EnsureSessionsRoot creates the sessions directory if it doesn't exist.
func EnsureSessionsRoot() error {
	return os.MkdirAll(GetSessionsRoot(), 0755)
}

// GetArchiveRoot returns the directory that holds archived sessions.
// Respects archiveDir in config if set. Otherwise it sits beside the sessions
// directory, which keeps archiving a same-filesystem rename.
func GetArchiveRoot() string {
	if cfg, err := Load(); err == nil && cfg.ArchiveDir != "" {
		return paths.ExpandHome(cfg.ArchiveDir)
	}
	return filepath.Join(filepath.Dir(GetSessionsRoot()), "archive")
}

// EnsureArchiveRoot creates the archive directory if it doesn't exist.
func EnsureArchiveRoot() error {
	return os.MkdirAll(GetArchiveRoot(), 0755)
}
