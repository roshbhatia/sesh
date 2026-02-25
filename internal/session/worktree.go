package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepo checks if a path is a git repository
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetRepoBasename returns the basename of a repository path
func GetRepoBasename(path string) string {
	return filepath.Base(path)
}

// CreateWorktree creates a git worktree for the given repo in the session directory
func CreateWorktree(repoPath, sessionPath, sessionName string) (string, error) {
	basename := GetRepoBasename(repoPath)
	worktreeName := fmt.Sprintf("%s-%s", basename, sessionName)
	worktreePath := filepath.Join(sessionPath, worktreeName)

	// Get the default branch name
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(output))

	// Create the worktree on the same branch (this will create a detached HEAD at the same commit)
	cmd = exec.Command("git", "-C", repoPath, "worktree", "add", "--detach", worktreePath, currentBranch)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	// Now checkout the branch in the worktree
	cmd = exec.Command("git", "-C", worktreePath, "checkout", "-B", currentBranch)
	if err := cmd.Run(); err != nil {
		// If branch creation fails, just checkout existing branch
		cmd = exec.Command("git", "-C", worktreePath, "checkout", currentBranch)
		cmd.Run()
	}

	return worktreePath, nil
}

// CreateSymlink creates a symlink for non-git directories
func CreateSymlink(target, sessionPath string) (string, error) {
	basename := filepath.Base(target)
	linkPath := filepath.Join(sessionPath, basename)

	if err := os.Symlink(target, linkPath); err != nil {
		return "", fmt.Errorf("failed to create symlink: %w", err)
	}

	return linkPath, nil
}

// CleanupWorktrees removes all git worktrees in a session directory
func CleanupWorktrees(sessionPath string) error {
	entries, err := os.ReadDir(sessionPath)
	if err != nil {
		return fmt.Errorf("failed to read session directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(sessionPath, entry.Name())

		// Check if this is a git worktree
		cmd := exec.Command("git", "-C", entryPath, "rev-parse", "--is-inside-work-tree")
		if cmd.Run() != nil {
			continue
		}

		// Get the git directory to find the main repo
		cmd = exec.Command("git", "-C", entryPath, "rev-parse", "--git-common-dir")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		gitCommonDir := strings.TrimSpace(string(output))

		// git-common-dir returns the .git dir of the main repo; its parent is the repo root
		mainRepoPath := filepath.Dir(gitCommonDir)

		// Remove the worktree
		cmd = exec.Command("git", "-C", mainRepoPath, "worktree", "remove", entryPath, "--force")
		if err := cmd.Run(); err != nil {
			// If removal fails, try to prune
			exec.Command("git", "-C", mainRepoPath, "worktree", "prune").Run()
		}
	}

	return nil
}

// GetWorktreeMainRepo finds the main repository for a worktree
func GetWorktreeMainRepo(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git common dir: %w", err)
	}
	gitCommonDir := strings.TrimSpace(string(output))

	// The main repo is the parent of the .git directory
	mainRepoPath := filepath.Dir(gitCommonDir)
	return mainRepoPath, nil
}
