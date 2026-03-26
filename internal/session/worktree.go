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

// IsGitRepo checks if a path is a git repository.
func IsGitRepo(path string) bool {
	_, err := gitExec(path, "rev-parse", "--git-dir")
	return err == nil
}

// GetRepoBasename returns the basename of a repository path.
func GetRepoBasename(path string) string {
	return filepath.Base(path)
}

// disambiguatedName generates a unique worktree directory name using bare basename.
// Tries: basename → <parent>-<basename> → <basename>-2, -3, etc.
func disambiguatedName(repoPath, sessionPath string) string {
	basename := GetRepoBasename(repoPath)

	if _, err := os.Stat(filepath.Join(sessionPath, basename)); os.IsNotExist(err) {
		return basename
	}

	parent := filepath.Base(filepath.Dir(repoPath))
	if parent != "" && parent != "." && parent != "/" {
		candidate := fmt.Sprintf("%s-%s", parent, basename)
		if _, err := os.Stat(filepath.Join(sessionPath, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", basename, i)
		if _, err := os.Stat(filepath.Join(sessionPath, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
}

// CreateWorktree creates a git worktree for the given repo in the session directory.
// branchName is a pre-rendered branch name (from template or --branch flag).
func CreateWorktree(repoPath, sessionPath, branchName string) (string, error) {
	worktreeName := disambiguatedName(repoPath, sessionPath)
	worktreePath := filepath.Join(sessionPath, worktreeName)

	// Primary strategy: create a new branch from HEAD
	_, err := gitExec(repoPath, "worktree", "add", worktreePath, "-b", branchName, "HEAD")
	if err != nil {
		// Fallback: branch already exists, reuse it
		_, err2 := gitExec(repoPath, "worktree", "add", worktreePath, branchName)
		if err2 != nil {
			return "", fmt.Errorf("failed to create worktree (primary: %w) (fallback: %v)", err, err2)
		}
	}

	return worktreePath, nil
}

// CreateSymlink creates a symlink for non-git directories.
func CreateSymlink(target, sessionPath string) (string, error) {
	basename := filepath.Base(target)
	linkPath := filepath.Join(sessionPath, basename)

	if _, err := os.Stat(linkPath); err == nil {
		parent := filepath.Base(filepath.Dir(target))
		if parent != "" && parent != "." && parent != "/" {
			linkPath = filepath.Join(sessionPath, fmt.Sprintf("%s-%s", parent, basename))
		}
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

// CleanupWorktrees removes all git worktrees in a session directory and deletes their branches.
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

		_, err := gitExec(entryPath, "rev-parse", "--is-inside-work-tree")
		if err != nil {
			continue
		}

		// Get the branch name before removing the worktree
		branchName, _ := gitExec(entryPath, "rev-parse", "--abbrev-ref", "HEAD")

		gitCommonDir, err := gitExec(entryPath, "rev-parse", "--git-common-dir")
		if err != nil {
			continue
		}

		mainRepoPath := filepath.Dir(gitCommonDir)

		// Remove the worktree
		_, removeErr := gitExec(mainRepoPath, "worktree", "remove", entryPath, "--force")
		if removeErr != nil {
			gitExec(mainRepoPath, "worktree", "prune")
		}

		// Delete the branch (non-fatal on failure)
		if branchName != "" && branchName != "HEAD" {
			// Check if this branch is the current branch of the main repo
			mainBranch, _ := gitExec(mainRepoPath, "rev-parse", "--abbrev-ref", "HEAD")
			if mainBranch != branchName {
				gitExec(mainRepoPath, "branch", "-D", branchName)
			}
			// If it's the current branch, skip deletion (can't delete checked-out branch)
		}
	}

	return nil
}

// GetWorktreeMainRepo finds the main repository for a worktree.
func GetWorktreeMainRepo(worktreePath string) (string, error) {
	gitCommonDir, err := gitExec(worktreePath, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Dir(gitCommonDir), nil
}

// ListRepoSources returns the resolved real paths for all repo sources in a session.
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
