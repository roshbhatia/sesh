package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all seshy configuration.
type Config struct {
	BranchFormat string `yaml:"branchFormat"`
	SessionsDir  string `yaml:"sessionsDir"`
}

// DefaultConfig returns a Config with all defaults populated.
func DefaultConfig() *Config {
	return &Config{
		BranchFormat: "sy/{{.Session}}/{{.Repo}}",
		SessionsDir:  defaultSessionsDir(),
	}
}

func defaultSessionsDir() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "seshy", "sessions")
}

// ConfigDir returns the config directory path following XDG.
func ConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "seshy")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Load reads the config file and merges with defaults.
// Missing file → defaults. Invalid YAML → error with path.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", ConfigPath(), err)
	}

	if len(data) == 0 {
		return cfg, nil
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", ConfigPath(), err)
	}

	// Merge: file values override defaults when non-zero
	if fileCfg.BranchFormat != "" {
		cfg.BranchFormat = fileCfg.BranchFormat
	}
	if fileCfg.SessionsDir != "" {
		cfg.SessionsDir = fileCfg.SessionsDir
	}

	return cfg, nil
}

// WriteDefault writes a commented default config file to the given path.
func WriteDefault(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	content := `# seshy configuration
# See: sy config

# Branch naming template for worktrees.
# Variables: {{.Session}}, {{.Repo}}, {{.User}}
# branchFormat: "sy/{{.Session}}/{{.Repo}}"

# Sessions storage directory.
# sessionsDir: "` + defaultSessionsDir() + `"
`
	return os.WriteFile(path, []byte(content), 0644)
}

// GetSessionsRoot returns the sessions root directory (convenience wrapper).
func GetSessionsRoot() string {
	cfg, err := Load()
	if err != nil {
		return defaultSessionsDir()
	}
	return cfg.SessionsDir
}

// EnsureSessionsRoot creates the sessions root directory if it doesn't exist.
func EnsureSessionsRoot() error {
	return os.MkdirAll(GetSessionsRoot(), 0755)
}
