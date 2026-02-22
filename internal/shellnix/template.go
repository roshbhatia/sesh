package shellnix

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed shell.nix
var shellNixTemplate string

// WriteTemplate writes the shell.nix template to the given directory
func WriteTemplate(dir string) error {
	path := filepath.Join(dir, "shell.nix")
	return os.WriteFile(path, []byte(shellNixTemplate), 0644)
}

// GetTemplate returns the shell.nix template content
func GetTemplate() string {
	return shellNixTemplate
}
