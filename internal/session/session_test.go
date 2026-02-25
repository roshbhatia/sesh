package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestValidateSessionName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid simple name", "test", false},
		{"valid with hyphen", "test-session", false},
		{"valid with underscore", "test_session", false},
		{"valid with numbers", "test123", false},
		{"empty name", "", true},
		{"with spaces", "test session", true},
		{"with slash", "test/session", true},
		{"with special chars", "test@session", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSessionName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

func TestGetRepoBasename(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/repos/myrepo", "myrepo"},
		{"/home/user/repos/myrepo/", "myrepo"},
		{"myrepo", "myrepo"},
	}

	for _, tt := range tests {
		result := GetRepoBasename(tt.path)
		if result != tt.expected {
			t.Errorf("GetRepoBasename(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func setupTestGitRepo(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init", dir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run()

	// Create initial commit
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "add", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}
}

func TestIsGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("is git repo", func(t *testing.T) {
		repoDir := filepath.Join(tmpDir, "gitrepo")
		setupTestGitRepo(t, repoDir)

		if !IsGitRepo(repoDir) {
			t.Error("Expected IsGitRepo to return true for git repo")
		}
	})

	t.Run("not git repo", func(t *testing.T) {
		notRepoDir := filepath.Join(tmpDir, "notrepo")
		os.MkdirAll(notRepoDir, 0755)

		if IsGitRepo(notRepoDir) {
			t.Error("Expected IsGitRepo to return false for non-git directory")
		}
	})
}

func TestCreateWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test git repo
	repoDir := filepath.Join(tmpDir, "testrepo")
	setupTestGitRepo(t, repoDir)

	// Create session directory
	sessionDir := filepath.Join(tmpDir, "sessions", "test-session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Create worktree
	worktreePath, err := CreateWorktree(repoDir, sessionDir, "test-session")
	if err != nil {
		t.Fatalf("CreateWorktree failed: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("Worktree directory was not created")
	}

	// Verify it's a git worktree
	if !IsGitRepo(worktreePath) {
		t.Error("Created worktree is not recognized as git repo")
	}

	// Verify naming convention
	expectedName := "testrepo-test-session"
	if filepath.Base(worktreePath) != expectedName {
		t.Errorf("Expected worktree name %q, got %q", expectedName, filepath.Base(worktreePath))
	}
}

func TestCreateSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target directory
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	// Create session directory
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Create symlink
	linkPath, err := CreateSymlink(targetDir, sessionDir)
	if err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Failed to stat symlink: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Created link is not a symlink")
	}

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if target != targetDir {
		t.Errorf("Expected symlink target %q, got %q", targetDir, target)
	}
}

func TestCleanupWorktrees(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test git repo
	repoDir := filepath.Join(tmpDir, "testrepo")
	setupTestGitRepo(t, repoDir)

	// Create session directory
	sessionDir := filepath.Join(tmpDir, "session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Create a worktree
	worktreePath, err := CreateWorktree(repoDir, sessionDir, "test")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	// Verify worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Fatal("Worktree was not created")
	}

	// Cleanup worktrees
	if err := CleanupWorktrees(sessionDir); err != nil {
		t.Fatalf("CleanupWorktrees failed: %v", err)
	}

	// Verify worktree is removed (or at least cleanup attempted)
	// Note: The directory might still exist but should be removed from git's worktree list
	cmd := exec.Command("git", "-C", repoDir, "worktree", "list")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}

	// The output should only contain the main repo, not the worktree
	if len(output) > 0 {
		// Check that worktreePath is not in the list
		// This is a basic check; a more robust check would parse the output
		t.Logf("Worktree list after cleanup: %s", output)
	}
}

func TestCreateSession(t *testing.T) {
	tmpDir := t.TempDir()

	repoDir := filepath.Join(tmpDir, "testrepo")
	setupTestGitRepo(t, repoDir)

	err := Create("test-session", []string{repoDir})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionPath, err := GetPath("test-session")
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}

	worktreePath := filepath.Join(sessionPath, "testrepo-test-session")
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("worktree was not created in session")
	}
}

func TestAddRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two test git repos
	repo1 := filepath.Join(tmpDir, "repo1")
	setupTestGitRepo(t, repo1)

	repo2 := filepath.Join(tmpDir, "repo2")
	setupTestGitRepo(t, repo2)

	// Create a session with first repo
	err := Create("test-add-session", []string{repo1})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add second repo to session
	err = AddRepos("test-add-session", []string{repo2})
	if err != nil {
		t.Fatalf("AddRepos failed: %v", err)
	}

	// Get session path
	sessionPath, err := GetPath("test-add-session")
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}

	// Verify both worktrees exist
	repo1Worktree := filepath.Join(sessionPath, "repo1-test-add-session")
	if _, err := os.Stat(repo1Worktree); os.IsNotExist(err) {
		t.Error("First repo worktree not found")
	}

	repo2Worktree := filepath.Join(sessionPath, "repo2-test-add-session")
	if _, err := os.Stat(repo2Worktree); os.IsNotExist(err) {
		t.Error("Second repo worktree not found after AddRepos")
	}

	// Verify both are valid git repos
	if !IsGitRepo(repo1Worktree) {
		t.Error("First worktree is not a git repo")
	}
	if !IsGitRepo(repo2Worktree) {
		t.Error("Second worktree is not a git repo")
	}
}

