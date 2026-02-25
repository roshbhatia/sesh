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

	if string(content) != shellNixTemplate {
		t.Error("shell.nix content does not match template")
	}

	// Verify the template contains essential structure
	if !contains(string(content), "mkShell") {
		t.Error("shell.nix missing mkShell")
	}
	if !contains(string(content), "buildInputs") {
		t.Error("shell.nix missing buildInputs")
	}
	if !contains(string(content), "git") {
		t.Error("shell.nix missing git package")
	}
}

func TestGetTemplate(t *testing.T) {
	tmpl := GetTemplate()
	if tmpl == "" {
		t.Error("GetTemplate returned empty string")
	}
	if tmpl != shellNixTemplate {
		t.Error("GetTemplate returned different content than expected")
	}

	// Verify content
	if !contains(tmpl, "{ pkgs ?") {
		t.Error("GetTemplate missing Nix function signature")
	}
}

func TestTemplateIsNotEmpty(t *testing.T) {
	if shellNixTemplate == "" {
		t.Fatal("embedded shell.nix template is empty - embed directive may not be working")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
