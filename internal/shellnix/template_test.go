package shellnix

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTemplate(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Write the template
	err := WriteTemplate(tmpDir)
	if err != nil {
		t.Fatalf("WriteTemplate failed: %v", err)
	}

	// Check that the file exists
	shellNixPath := filepath.Join(tmpDir, "shell.nix")
	if _, err := os.Stat(shellNixPath); os.IsNotExist(err) {
		t.Fatal("shell.nix was not created")
	}

	// Read the file and verify content
	content, err := os.ReadFile(shellNixPath)
	if err != nil {
		t.Fatalf("Failed to read shell.nix: %v", err)
	}

	if string(content) != template {
		t.Error("shell.nix content does not match template")
	}
}

func TestGetTemplate(t *testing.T) {
	tmpl := GetTemplate()
	if tmpl == "" {
		t.Error("GetTemplate returned empty string")
	}
	if tmpl != template {
		t.Error("GetTemplate returned different content than expected")
	}
}
