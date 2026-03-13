package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/roshbhatia/sesh/internal/config"
)

// Session represents a sesh session
type Session struct {
	Name         string
	Path         string
	RepoCount    int
	LastModified time.Time
}

// AddResult holds the outcome of adding multiple repos.
type AddResult struct {
	Added   []string
	Skipped []string
	Errors  map[string]error
}

// Err returns a combined error if any repos failed to add, or nil.
func (r AddResult) Err() error {
	if len(r.Errors) == 0 {
		return nil
	}
	var parts []string
	for repo, err := range r.Errors {
		parts = append(parts, fmt.Sprintf("  %s: %v", repo, err))
	}
	return fmt.Errorf("failed to add %d repo(s):\n%s", len(r.Errors), strings.Join(parts, "\n"))
}

// ValidateSessionName checks if a session name is valid
func ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	// Check for invalid characters
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("session name must contain only letters, numbers, hyphens, and underscores")
		}
	}

	return nil
}

// GetPath returns the absolute path to a session
func GetPath(name string) (string, error) {
	root := config.GetSessionsRoot()
	sessionPath := filepath.Join(root, name)

	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session '%s' not found", name)
	}

	return sessionPath, nil
}

// Exists checks if a session exists
func Exists(name string) bool {
	root := config.GetSessionsRoot()
	sessionPath := filepath.Join(root, name)
	_, err := os.Stat(sessionPath)
	return err == nil
}

// Create creates a new session with the given repos
func Create(name string, repoPaths []string) error {
	if err := ValidateSessionName(name); err != nil {
		return err
	}

	if Exists(name) {
		return fmt.Errorf("session '%s' already exists", name)
	}

	root := config.GetSessionsRoot()
	sessionPath := filepath.Join(root, name)

	// Create session directory
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Add repos/worktrees
	for _, repoPath := range repoPaths {
		if IsGitRepo(repoPath) {
			if _, err := CreateWorktree(repoPath, sessionPath, name); err != nil {
				// Clean up on failure
				os.RemoveAll(sessionPath)
				return fmt.Errorf("failed to create worktree for %s: %w", repoPath, err)
			}
		} else {
			if _, err := CreateSymlink(repoPath, sessionPath); err != nil {
				// Clean up on failure
				os.RemoveAll(sessionPath)
				return fmt.Errorf("failed to create symlink for %s: %w", repoPath, err)
			}
		}
	}

	return nil
}

// List returns all sessions
func List() ([]Session, error) {
	root := config.GetSessionsRoot()

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return []Session{}, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	sessions := make([]Session, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionPath := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		repoCount := 0
		sessionEntries, err := os.ReadDir(sessionPath)
		if err == nil {
			for _, se := range sessionEntries {
				// Skip hidden files like .DS_Store
				if strings.HasPrefix(se.Name(), ".") {
					continue
				}
				// Only count directories and symlinks (actual repos)
				if se.IsDir() || se.Type()&os.ModeSymlink != 0 {
					repoCount++
				}
			}
		}

		sessions = append(sessions, Session{
			Name:         entry.Name(),
			Path:         sessionPath,
			RepoCount:    repoCount,
			LastModified: info.ModTime(),
		})
	}

	return sessions, nil
}

// resolveRepoPath resolves a repo path to its real filesystem path.
func resolveRepoPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

// AddRepos adds repositories to an existing session.
// It detects duplicates, handles partial failures, and returns an AddResult.
func AddRepos(name string, repoPaths []string) error {
	sessionPath, err := GetPath(name)
	if err != nil {
		return err
	}

	result := AddResult{
		Errors: make(map[string]error),
	}

	// Build set of already-present repo source paths
	existingSources, err := ListRepoSources(sessionPath)
	if err != nil {
		existingSources = nil // non-fatal: proceed without duplicate detection
	}
	existingSet := make(map[string]bool, len(existingSources))
	for _, s := range existingSources {
		existingSet[resolveRepoPath(s)] = true
	}

	// Add repos/worktrees
	for _, repoPath := range repoPaths {
		resolved := resolveRepoPath(repoPath)

		// Duplicate detection
		if existingSet[resolved] {
			result.Skipped = append(result.Skipped, repoPath)
			continue
		}

		if IsGitRepo(repoPath) {
			if _, err := CreateWorktree(repoPath, sessionPath, name); err != nil {
				result.Errors[repoPath] = err
				continue
			}
		} else {
			if _, err := CreateSymlink(repoPath, sessionPath); err != nil {
				result.Errors[repoPath] = err
				continue
			}
		}

		result.Added = append(result.Added, repoPath)
		existingSet[resolved] = true
	}

	return result.Err()
}

// AddReposResult adds repositories to an existing session and returns detailed results.
func AddReposResult(name string, repoPaths []string) (AddResult, error) {
	sessionPath, err := GetPath(name)
	if err != nil {
		return AddResult{}, err
	}

	result := AddResult{
		Errors: make(map[string]error),
	}

	// Build set of already-present repo source paths
	existingSources, err := ListRepoSources(sessionPath)
	if err != nil {
		existingSources = nil
	}
	existingSet := make(map[string]bool, len(existingSources))
	for _, s := range existingSources {
		existingSet[resolveRepoPath(s)] = true
	}

	// Add repos/worktrees
	for _, repoPath := range repoPaths {
		resolved := resolveRepoPath(repoPath)

		// Duplicate detection
		if existingSet[resolved] {
			result.Skipped = append(result.Skipped, repoPath)
			continue
		}

		if IsGitRepo(repoPath) {
			if _, err := CreateWorktree(repoPath, sessionPath, name); err != nil {
				result.Errors[repoPath] = err
				continue
			}
		} else {
			if _, err := CreateSymlink(repoPath, sessionPath); err != nil {
				result.Errors[repoPath] = err
				continue
			}
		}

		result.Added = append(result.Added, repoPath)
		existingSet[resolved] = true
	}

	return result, nil
}

// Delete removes a session and cleans up worktrees
func Delete(name string) error {
	sessionPath, err := GetPath(name)
	if err != nil {
		return err
	}

	// Clean up any git worktrees
	if err := CleanupWorktrees(sessionPath); err != nil {
		return fmt.Errorf("failed to cleanup worktrees: %w", err)
	}

	// Remove the session directory
	if err := os.RemoveAll(sessionPath); err != nil {
		return fmt.Errorf("failed to remove session directory: %w", err)
	}

	return nil
}
