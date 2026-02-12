package shellnix

import (
	"os"
	"path/filepath"
)

const template = `{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    git
    # Add your development tools here
    # Examples:
    # nodejs
    # go
    # python3
    # rustc
  ];
  
  shellHook = ''
    echo "Session: $(basename $PWD)"
  '';
}
`

// WriteTemplate writes the shell.nix template to the given directory
func WriteTemplate(dir string) error {
	path := filepath.Join(dir, "shell.nix")
	return os.WriteFile(path, []byte(template), 0644)
}

// GetTemplate returns the shell.nix template content
func GetTemplate() string {
	return template
}
