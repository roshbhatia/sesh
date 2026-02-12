package fzf

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/roshbhatia/sesh/internal/config"
)

// SelectRepos launches fzf to select repositories from zoxide database
func SelectRepos() ([]string, error) {
	// Get directories from zoxide
	cmd := exec.Command("zoxide", "query", "--list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get zoxide directories: %w", err)
	}

	dirs := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no directories found in zoxide database")
	}

	// Launch fzf with multi-select
	selected, err := runFZF(dirs, true, "Select repositories (Space to select, Enter to confirm)")
	if err != nil {
		return nil, err
	}

	return selected, nil
}

// SelectSession launches fzf to select a session
func SelectSession(sessions []string) (string, error) {
	if len(sessions) == 0 {
		return "", fmt.Errorf("no sessions available")
	}

	selected, err := runFZF(sessions, false, "Select session")
	if err != nil {
		return "", err
	}

	if len(selected) == 0 {
		return "", fmt.Errorf("no session selected")
	}

	return selected[0], nil
}

// runFZF runs fzf with the given items
func runFZF(items []string, multiSelect bool, prompt string) ([]string, error) {
	// Build fzf command
	args := []string{
		"--prompt", prompt + " > ",
		"--height", "40%",
		"--reverse",
		"--border",
	}

	if multiSelect {
		args = append(args, "--multi")
	}

	// Add custom FZF options from environment
	customOpts := config.GetFZFOpts()
	if customOpts != "" {
		// Split custom options and add them
		opts := strings.Fields(customOpts)
		args = append(args, opts...)
	}

	cmd := exec.Command("fzf", args...)
	
	// Prepare input
	input := strings.Join(items, "\n")
	cmd.Stdin = strings.NewReader(input)
	
	// Capture output
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr

	// Run fzf
	if err := cmd.Run(); err != nil {
		// fzf returns exit code 130 when user cancels (Ctrl-C)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return nil, fmt.Errorf("selection cancelled")
		}
		return nil, fmt.Errorf("fzf failed: %w", err)
	}

	// Parse output
	output := strings.TrimSpace(outBuf.String())
	if output == "" {
		return []string{}, nil
	}

	selected := strings.Split(output, "\n")
	return selected, nil
}

// CheckDependencies verifies that required external tools are available
func CheckDependencies() error {
	// Check for fzf
	if _, err := exec.LookPath("fzf"); err != nil {
		return fmt.Errorf("fzf not found in PATH. Please install fzf: https://github.com/junegunn/fzf")
	}

	// Check for zoxide
	if _, err := exec.LookPath("zoxide"); err != nil {
		return fmt.Errorf("zoxide not found in PATH. Please install zoxide: https://github.com/ajeetdsouza/zoxide")
	}

	return nil
}
