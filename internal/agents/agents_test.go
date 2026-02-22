package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	err := WriteTemplate(tmpDir)
	if err != nil {
		t.Fatalf("WriteTemplate failed: %v", err)
	}

	agentsMdPath := filepath.Join(tmpDir, "agents.md")
	if _, err := os.Stat(agentsMdPath); os.IsNotExist(err) {
		t.Fatal("agents.md was not created")
	}

	// Verify file permissions are readable
	info, err := os.Stat(agentsMdPath)
	if err != nil {
		t.Fatalf("Failed to stat agents.md: %v", err)
	}
	if info.Mode()&0444 == 0 {
		t.Error("agents.md not readable")
	}

	content, err := os.ReadFile(agentsMdPath)
	if err != nil {
		t.Fatalf("Failed to read agents.md: %v", err)
	}

	if string(content) != agentsMdTemplate {
		t.Error("agents.md content does not match template")
	}

	// Verify essential structure
	contentStr := string(content)
	if !strings.Contains(contentStr, "# Agents guidance") {
		t.Error("agents.md missing header")
	}
	if !strings.Contains(contentStr, "Use Nix when possible") {
		t.Error("agents.md missing Nix guidance")
	}
	if !strings.Contains(contentStr, "shell.nix") {
		t.Error("agents.md missing shell.nix reference")
	}
	if !strings.Contains(contentStr, "direnv allow") {
		t.Error("agents.md missing direnv instructions")
	}
}

func TestGetTemplate(t *testing.T) {
	tmpl := GetTemplate()
	if tmpl == "" {
		t.Error("GetTemplate returned empty string")
	}
	if tmpl != agentsMdTemplate {
		t.Error("GetTemplate returned different content than expected")
	}

	// Verify core sections exist
	if !strings.Contains(tmpl, "Agent workflow") {
		t.Error("GetTemplate missing Agent workflow section")
	}
	if !strings.Contains(tmpl, "session") {
		t.Error("GetTemplate missing session references")
	}
}

func TestTemplateIsNotEmpty(t *testing.T) {
	if agentsMdTemplate == "" {
		t.Fatal("embedded agents.md template is empty - embed directive may not be working")
	}
}

func TestTemplateLength(t *testing.T) {
	if len(agentsMdTemplate) < 100 {
		t.Errorf("agents.md template suspiciously small: %d bytes", len(agentsMdTemplate))
	}
}
