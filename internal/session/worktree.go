package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/roshbhatia/go-utils/git"
)

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

// CreateWorktree creates a git worktree for the given repo in the session
// directory. branchName is a pre-rendered branch name (from template or
// --branch flag). The branch starts from HEAD, and an unborn repository is an
// error rather than an orphan worktree. When the branch already exists git
// checks it out instead of creating it, and reused reports that so the caller
// can surface the DWIM.
func CreateWorktree(repoPath, sessionPath, branchName string) (worktreePath string, reused bool, err error) {
	worktreePath = filepath.Join(sessionPath, disambiguatedName(repoPath, sessionPath))
	reused, err = git.WorktreeAdd(repoPath, worktreePath, git.WorktreeAddOptions{Branch: branchName, Start: "HEAD", Reuse: true})
	if err != nil {
		return "", false, err
	}
	return worktreePath, reused, nil
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

// removeWorktree unregisters a worktree from its main repo.
//
// One --force discards local changes. force adds the second --force that git
// demands before it will remove a locked worktree; without it a single locked
// worktree leaves both the registration and the branch behind.
func removeWorktree(mainRepoPath, worktreePath string, force bool) error {
	level := 1
	if force {
		level = 2
	}
	removeErr := git.WorktreeRemove(mainRepoPath, worktreePath, level)
	if removeErr == nil {
		return nil
	}

	// prune clears registrations whose directory is already gone, which is the
	// common reason remove fails. It exits 0 even when it clears nothing, so
	// confirm the registration actually went away rather than trusting it.
	_ = git.WorktreePrune(mainRepoPath)
	if worktreeRegistered(mainRepoPath, worktreePath) {
		return removeErr
	}
	return nil
}

// worktreeRegistered reports whether mainRepo still lists worktreePath.
func worktreeRegistered(mainRepoPath, worktreePath string) bool {
	trees, err := git.Worktrees(mainRepoPath)
	if err != nil {
		return false
	}
	target := realPath(worktreePath)
	for _, tree := range trees {
		if realPath(tree.Path) == target {
			return true
		}
	}
	return false
}

// realPath resolves symlinks where it can, falling back to a lexical clean for
// paths that no longer exist.
func realPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// errStandaloneClone marks a session entry that is a repository of its own
// rather than a worktree registered elsewhere. git worktree remove would
// refuse it, so the caller deletes the directory instead.
var errStandaloneClone = errors.New("entry is a standalone clone")

// teardownWorktree unregisters the worktree at entryPath from its main repo
// and deletes the branch it was checked out on. force is passed through to the
// removal so locked worktrees can be torn down.
func teardownWorktree(entryPath string, force bool) error {
	mainRepoPath, err := git.MainWorktree(entryPath)
	if err != nil {
		return fmt.Errorf("could not locate main repo: %w", err)
	}
	if filepath.Clean(mainRepoPath) == filepath.Clean(entryPath) {
		return errStandaloneClone
	}

	branchName, _ := git.Branch(entryPath)

	if err := removeWorktree(mainRepoPath, entryPath, force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	if branchName == "" || branchName == "HEAD" {
		return nil
	}
	// The main worktree may sit on the same branch; git refuses to delete a
	// checked-out branch and the user did not ask us to move it.
	if mainBranch, _ := git.Branch(mainRepoPath); mainBranch == branchName {
		return nil
	}
	if err := git.Run(mainRepoPath, "branch", "-D", branchName); err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branchName, err)
	}
	return nil
}

// CleanupWorktrees removes all git worktrees in a session directory and deletes their branches.
// Continues on individual failures and returns a combined error if any worktree could not be removed.
// force is passed through to worktree removal so locked worktrees can be torn down.
func CleanupWorktrees(sessionPath string, force bool) error {
	entries, err := os.ReadDir(sessionPath)
	if err != nil {
		return fmt.Errorf("failed to read session directory: %w", err)
	}

	var errs []error

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(sessionPath, entry.Name())
		if !git.IsRepo(entryPath) {
			continue
		}

		// A standalone clone is removed with the session directory itself.
		if err := teardownWorktree(entryPath, force); err != nil && !errors.Is(err, errStandaloneClone) {
			errs = append(errs, fmt.Errorf("%s: %w", entry.Name(), err))
		}
	}

	return errors.Join(errs...)
}

// RemoveRepoEntry removes a single repo entry from a session directory.
// For git worktrees: removes the worktree and deletes the branch.
// For symlinks: removes the symlink.
// force is passed through to worktree removal so locked worktrees can be removed.
func RemoveRepoEntry(sessionPath, repoName string, force bool) error {
	entryPath := filepath.Join(sessionPath, repoName)
	info, err := os.Lstat(entryPath)
	if err != nil {
		return fmt.Errorf("repo %q not found in session: %w", repoName, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(entryPath)
	}

	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory or symlink", repoName)
	}

	if !git.IsRepo(entryPath) {
		return os.RemoveAll(entryPath)
	}

	err = teardownWorktree(entryPath, force)
	if errors.Is(err, errStandaloneClone) {
		return os.RemoveAll(entryPath)
	}
	return err
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
			sources = append(sources, realPath(target))
			continue
		}

		if !info.IsDir() {
			continue
		}

		mainRepoPath, err := git.MainWorktree(entryPath)
		if err != nil {
			continue
		}
		sources = append(sources, realPath(mainRepoPath))
	}

	return sources, nil
}
