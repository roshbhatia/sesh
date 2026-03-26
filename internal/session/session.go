package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/tmpl"
)

// Session represents a seshy session.
type Session struct {
	Name         string
	Path         string
	RepoCount    int
	LastModified time.Time
}

// RepoInfo describes a repo that was created in a session.
type RepoInfo struct {
	Name       string // basename in session dir
	Path       string // absolute worktree/symlink path
	SourcePath string // absolute original repo path
	Branch     string // rendered branch name (empty for non-git)
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

// ValidateSessionName checks if a session name is valid.
func ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("session name must contain only letters, numbers, hyphens, and underscores")
		}
	}
	return nil
}

// GetPath returns the absolute path to a session.
func GetPath(name string) (string, error) {
	root := config.GetSessionsRoot()
	sessionPath := filepath.Join(root, name)
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session '%s' not found", name)
	}
	return sessionPath, nil
}

// Exists checks if a session exists.
func Exists(name string) bool {
	root := config.GetSessionsRoot()
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}

// branchForRepo computes the branch name for a repo.
func branchForRepo(branchFormat, branchOverride, sessionName, repoPath string) (string, error) {
	if branchOverride != "" {
		if err := ValidateBranchName(branchOverride); err != nil {
			return "", err
		}
		return branchOverride, nil
	}
	return RenderBranchName(branchFormat, sessionName, GetRepoBasename(repoPath))
}

// CreateOpts holds options for session creation.
type CreateOpts struct {
	BranchFormat   string
	BranchOverride string
}

// Create creates a new session with the given repos. Atomic: on failure,
// all previously created worktrees and branches are cleaned up.
// Returns RepoInfo for each successfully created repo.
func Create(name string, repoPaths []string, opts CreateOpts) ([]RepoInfo, error) {
	if err := ValidateSessionName(name); err != nil {
		return nil, err
	}
	if Exists(name) {
		return nil, fmt.Errorf("session '%s' already exists", name)
	}

	root := config.GetSessionsRoot()
	sessionPath := filepath.Join(root, name)

	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	type created struct {
		worktreePath string
		repoPath     string
		branchName   string
	}
	var createdList []created
	var repoInfos []RepoInfo

	cleanup := func() {
		for i := len(createdList) - 1; i >= 0; i-- {
			c := createdList[i]
			if IsGitRepo(c.repoPath) && c.branchName != "" {
				gitExec(c.repoPath, "worktree", "remove", c.worktreePath, "--force")
				gitExec(c.repoPath, "branch", "-D", c.branchName)
			}
		}
		os.RemoveAll(sessionPath)
	}

	for _, repoPath := range repoPaths {
		if IsGitRepo(repoPath) {
			branch, err := branchForRepo(opts.BranchFormat, opts.BranchOverride, name, repoPath)
			if err != nil {
				cleanup()
				return nil, fmt.Errorf("branch name for %s: %w", repoPath, err)
			}

			wtPath, err := CreateWorktree(repoPath, sessionPath, branch)
			if err != nil {
				cleanup()
				return nil, fmt.Errorf("failed to create worktree for %s: %w", repoPath, err)
			}
			createdList = append(createdList, created{worktreePath: wtPath, repoPath: repoPath, branchName: branch})
			repoInfos = append(repoInfos, RepoInfo{
				Name:       filepath.Base(wtPath),
				Path:       wtPath,
				SourcePath: repoPath,
				Branch:     branch,
			})
		} else {
			linkPath, err := CreateSymlink(repoPath, sessionPath)
			if err != nil {
				cleanup()
				return nil, fmt.Errorf("failed to create symlink for %s: %w", repoPath, err)
			}
			repoInfos = append(repoInfos, RepoInfo{
				Name:       filepath.Base(linkPath),
				Path:       linkPath,
				SourcePath: repoPath,
			})
		}
	}

	return repoInfos, nil
}

// List returns all sessions.
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
				if strings.HasPrefix(se.Name(), ".") {
					continue
				}
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

func resolveRepoPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

// AddRepos adds repositories to an existing session (best-effort).
// Returns AddResult, RepoInfo for newly added repos, and error.
func AddRepos(name string, repoPaths []string, opts CreateOpts) (AddResult, []RepoInfo, error) {
	sessionPath, err := GetPath(name)
	if err != nil {
		return AddResult{}, nil, err
	}

	result := AddResult{Errors: make(map[string]error)}
	var newRepos []RepoInfo

	existingSources, err := ListRepoSources(sessionPath)
	if err != nil {
		existingSources = nil
	}
	existingSet := make(map[string]bool, len(existingSources))
	for _, s := range existingSources {
		existingSet[resolveRepoPath(s)] = true
	}

	for _, repoPath := range repoPaths {
		resolved := resolveRepoPath(repoPath)
		if existingSet[resolved] {
			result.Skipped = append(result.Skipped, repoPath)
			continue
		}

		if IsGitRepo(repoPath) {
			branch, err := branchForRepo(opts.BranchFormat, opts.BranchOverride, name, repoPath)
			if err != nil {
				result.Errors[repoPath] = err
				continue
			}
			wtPath, err := CreateWorktree(repoPath, sessionPath, branch)
			if err != nil {
				result.Errors[repoPath] = err
				continue
			}
			newRepos = append(newRepos, RepoInfo{
				Name:       filepath.Base(wtPath),
				Path:       wtPath,
				SourcePath: repoPath,
				Branch:     branch,
			})
		} else {
			linkPath, err := CreateSymlink(repoPath, sessionPath)
			if err != nil {
				result.Errors[repoPath] = err
				continue
			}
			newRepos = append(newRepos, RepoInfo{
				Name:       filepath.Base(linkPath),
				Path:       linkPath,
				SourcePath: repoPath,
			})
		}

		result.Added = append(result.Added, repoPath)
		existingSet[resolved] = true
	}

	return result, newRepos, nil
}

// BuildTemplateData creates TemplateData from session info.
func BuildTemplateData(name, sessionPath string, repos []RepoInfo) tmpl.TemplateData {
	repoData := make([]tmpl.RepoData, len(repos))
	for i, r := range repos {
		repoData[i] = tmpl.RepoData{
			Name:   r.Name,
			Path:   r.Path,
			Source: r.SourcePath,
			Branch: r.Branch,
		}
	}
	return tmpl.NewTemplateData(name, sessionPath, repoData)
}

// Delete removes a session and cleans up worktrees + branches.
func Delete(name string) error {
	sessionPath, err := GetPath(name)
	if err != nil {
		return err
	}

	if err := CleanupWorktrees(sessionPath); err != nil {
		return fmt.Errorf("failed to cleanup worktrees: %w", err)
	}

	if err := os.RemoveAll(sessionPath); err != nil {
		return fmt.Errorf("failed to remove session directory: %w", err)
	}

	return nil
}
