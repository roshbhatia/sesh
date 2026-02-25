package agents

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed agents.md
var agentsMdTemplate string

// WriteTemplate writes the agents.md template to the given directory
func WriteTemplate(dir string) error {
	path := filepath.Join(dir, "agents.md")
	return os.WriteFile(path, []byte(agentsMdTemplate), 0644)
}

// GetTemplate returns the agents.md template content
func GetTemplate() string {
	return agentsMdTemplate
}
