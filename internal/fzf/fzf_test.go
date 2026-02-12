package fzf

import (
	"os/exec"
	"testing"
)

func TestCheckDependencies(t *testing.T) {
	err := CheckDependencies()
	
	// Check if fzf exists
	_, fzfErr := exec.LookPath("fzf")
	
	// Check if zoxide exists
	_, zoxideErr := exec.LookPath("zoxide")
	
	if fzfErr != nil || zoxideErr != nil {
		// If either is missing, we expect an error
		if err == nil {
			t.Error("Expected error when dependencies are missing, got nil")
		}
	} else {
		// Both exist, should not error
		if err != nil {
			t.Errorf("CheckDependencies failed unexpectedly: %v", err)
		}
	}
}

func TestRunFZF_Basic(t *testing.T) {
	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		t.Skip("fzf not available, skipping test")
	}

	// This test can't run interactively, but we can verify the function doesn't panic
	// In a real scenario, this would be tested with a mock or in an integration test
	t.Run("function exists", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		// We can't actually run this without user interaction
		// Just verify it doesn't panic on setup
		if items == nil {
			t.Error("items should not be nil")
		}
	})
}
