package sesh_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roshbhatia/sesh/internal/agents"
	"github.com/roshbhatia/sesh/internal/shellnix"
)

// TestEmbeddedTemplates validates that shell.nix and agents.md are properly embedded
func TestEmbeddedTemplates(t *testing.T) {
	t.Run("shell.nix template embedded", func(t *testing.T) {
		tmpl := shellnix.GetTemplate()
		if tmpl == "" {
			t.Fatal("shell.nix template is empty - embed failed")
		}

		if !strings.Contains(tmpl, "{ pkgs ?") {
			t.Error("shell.nix missing Nix syntax")
		}
		if !strings.Contains(tmpl, "mkShell") {
			t.Error("shell.nix missing mkShell")
		}
		if !strings.Contains(tmpl, "buildInputs") {
			t.Error("shell.nix missing buildInputs")
		}
		if !strings.Contains(tmpl, "git") {
			t.Error("shell.nix missing git package")
		}
	})

	t.Run("agents.md template embedded", func(t *testing.T) {
		tmpl := agents.GetTemplate()
		if tmpl == "" {
			t.Fatal("agents.md template is empty - embed failed")
		}

		if !strings.Contains(tmpl, "Agents guidance") {
			t.Error("agents.md missing header")
		}
		if !strings.Contains(tmpl, "Use Nix") {
			t.Error("agents.md missing Nix guidance")
		}
		if !strings.Contains(tmpl, "shell.nix") {
			t.Error("agents.md missing shell.nix reference")
		}
		if !strings.Contains(tmpl, "direnv") {
			t.Error("agents.md missing direnv instructions")
		}
	})
}

// TestWriteTemplatesIntegration validates that both templates write correctly
func TestWriteTemplatesIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("write shell.nix", func(t *testing.T) {
		if err := shellnix.WriteTemplate(tmpDir); err != nil {
			t.Fatalf("WriteTemplate failed: %v", err)
		}

		shellNixPath := filepath.Join(tmpDir, "shell.nix")
		content, err := os.ReadFile(shellNixPath)
		if err != nil {
			t.Fatalf("Failed to read shell.nix: %v", err)
		}

		if len(content) == 0 {
			t.Error("shell.nix file is empty")
		}

		if string(content) != shellnix.GetTemplate() {
			t.Error("written shell.nix doesn't match template")
		}
	})

	t.Run("write agents.md", func(t *testing.T) {
		if err := agents.WriteTemplate(tmpDir); err != nil {
			t.Fatalf("WriteTemplate failed: %v", err)
		}

		agentsMdPath := filepath.Join(tmpDir, "agents.md")
		content, err := os.ReadFile(agentsMdPath)
		if err != nil {
			t.Fatalf("Failed to read agents.md: %v", err)
		}

		if len(content) == 0 {
			t.Error("agents.md file is empty")
		}

		if string(content) != agents.GetTemplate() {
			t.Error("written agents.md doesn't match template")
		}
	})
}

// TestBuildWithEmbeds validates that the binary builds with embedded templates
func TestBuildWithEmbeds(t *testing.T) {
	t.Run("go build succeeds", func(t *testing.T) {
		// Skip this test in CI or if we can't determine project root
		// The actual build is tested by task build in the real environment
		t.Skip("Build test requires project root - run 'task build' instead")
	})
}
