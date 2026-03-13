package session

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitExec runs a git command and returns its stdout. If the command fails,
// the error wraps the subcommand name, repo path, and stderr content.
func gitExec(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", fullArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		subcmd := ""
		if len(args) > 0 {
			subcmd = args[0]
		}
		return "", fmt.Errorf("git %s failed for %s: %s: %w", subcmd, repoPath, strings.TrimSpace(stderr.String()), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// IsGitRepo checks if a path is a git repository
func IsGitRepo(path string) bool {
	_, err := gitExec(path, "rev-parse", "--git-dir")
	return err == nil
}

// GetRepoBasename returns the basename of a repository path
func GetRepoBasename(path string) string {
	return filepath.Base(path)
}

// disambiguatedName generates a unique worktree directory name for a repo
// within a session directory. If the simple basename collides with an existing
// entry, it prefixes with the parent directory name. If that also collides,
// it appends a numeric suffix.
func disambiguatedName(repoPath, sessionPath, sessionName string) string {
	basename := GetRepoBasename(repoPath)
	candidate := fmt.Sprintf("%s-%s", basename, sessionName)

	// Check if the simple name is available
	if _, err := os.Stat(filepath.Join(sessionPath, candidate)); os.IsNotExist(err) {
		return candidate
	}

	// Try parent directory prefix
	parent := filepath.Base(filepath.Dir(repoPath))
	if parent != "" && parent != "." && parent != "/" {
		candidate = fmt.Sprintf("%s-%s-%s", parent, basename, sessionName)
		if _, err := os.Stat(filepath.Join(sessionPath, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}

	// Numeric suffix fallback
	for i := 2; ; i++ {
		candidate = fmt.Sprintf("%s-%d-%s", basename, i, sessionName)
		if _, err := os.Stat(filepath.Join(sessionPath, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
}

// CreateWorktree creates a git worktree for the given repo in the session directory.
// It uses a session-scoped branch named sesh/<session>/<basename> from the current HEAD.
func CreateWorktree(repoPath, sessionPath, sessionName string) (string, error) {
	worktreeName := disambiguatedName(repoPath, sessionPath, sessionName)
	worktreePath := filepath.Join(sessionPath, worktreeName)
	basename := GetRepoBasename(repoPath)
	branchName := fmt.Sprintf("sesh/%s/%s", sessionName, basename)

	// Primary strategy: create a new branch from HEAD
	_, err := gitExec(repoPath, "worktree", "add", worktreePath, "-b", branchName, "HEAD")
	if err != nil {
		// Fallback: branch already exists (e.g. from a previously deleted session), reuse it
		_, err2 := gitExec(repoPath, "worktree", "add", worktreePath, branchName)
		if err2 != nil {
			return "", fmt.Errorf("failed to create worktree (primary: %w) (fallback: %v)", err, err2)
		}
	}

	return worktreePath, nil
}

// CreateSymlink creates a symlink for non-git directories.
// Uses disambiguation to avoid collisions with existing entries.
func CreateSymlink(target, sessionPath string) (string, error) {
	basename := filepath.Base(target)
	linkPath := filepath.Join(sessionPath, basename)

	// If simple name collides, disambiguate
	if _, err := os.Stat(linkPath); err == nil {
		parent := filepath.Base(filepath.Dir(target))
		if parent != "" && parent != "." && parent != "/" {
			linkPath = filepath.Join(sessionPath, fmt.Sprintf("%s-%s", parent, basename))
		}
		// If still collides, numeric suffix
		if _, err := os.Stat(linkPath); err == nil {
			for i := 2; ; i++ {
				candidate := filepath.Join(sessionPath, fmt.Sprintf("%s-%d", basename, i))
				if _, err := os.Stat(candidate); os.IsNotExist(err) {
					linkPath = candidate
					break
				}
			}
		}
	}

	if err := os.Symlink(target, linkPath); err != nil {
		return "", fmt.Errorf("failed to create symlink for %s: %w", target, err)
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
		_, err := gitExec(entryPath, "rev-parse", "--is-inside-work-tree")
		if err != nil {
			continue
		}

		// Get the git directory to find the main repo
		gitCommonDir, err := gitExec(entryPath, "rev-parse", "--git-common-dir")
		if err != nil {
			continue
		}

		// git-common-dir returns the .git dir of the main repo; its parent is the repo root
		mainRepoPath := filepath.Dir(gitCommonDir)

		// Remove the worktree
		_, removeErr := gitExec(mainRepoPath, "worktree", "remove", entryPath, "--force")
		if removeErr != nil {
			// If removal fails, try to prune
			gitExec(mainRepoPath, "worktree", "prune")
		}
	}

	return nil
}

// GetWorktreeMainRepo finds the main repository for a worktree
func GetWorktreeMainRepo(worktreePath string) (string, error) {
	gitCommonDir, err := gitExec(worktreePath, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}

	// The main repo is the parent of the .git directory
	mainRepoPath := filepath.Dir(gitCommonDir)
	return mainRepoPath, nil
}

// ListRepoSources returns the resolved real paths for all repo sources in a session.
// For worktrees, it queries git rev-parse --git-common-dir to find the parent repo.
// For symlinks, it resolves the link target.
func ListRepoSources(sessionPath string) ([]string, error) {
	entries, err := os.ReadDir(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	var sources []string
	for _, entry := range entries {
		entryPath := filepath.Join(sessionPath, entry.Name())

		info, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// Symlink: resolve target
			target, err := os.Readlink(entryPath)
			if err != nil {
				continue
			}
			resolved, err := filepath.EvalSymlinks(target)
			if err != nil {
				resolved = target
			}
			sources = append(sources, resolved)
			continue
		}

		if !info.IsDir() {
			continue
		}

		// Directory: check if it's a git worktree
		gitCommonDir, err := gitExec(entryPath, "rev-parse", "--git-common-dir")
		if err != nil {
			continue
		}

		mainRepoPath := filepath.Dir(gitCommonDir)
		resolved, err := filepath.EvalSymlinks(mainRepoPath)
		if err != nil {
			resolved = mainRepoPath
		}
		sources = append(sources, resolved)
	}

	return sources, nil
}
