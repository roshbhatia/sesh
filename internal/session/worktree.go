package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// sessionKey is the git config key, under a branch's section, that names the
// session seshy created the branch for. It is the link back from a branch to
// its session once the worktree no longer has the branch checked out.
func sessionKey(branch string) string { return "branch." + branch + ".seshy-session" }

// reusedKey marks a branch that existed before seshy checked it out, so a
// later status can tell a reused branch from a created one.
func reusedKey(branch string) string { return "branch." + branch + ".seshy-reused" }

// recordBranch writes the back-pointer from branch to session in the repo's
// local config. The worktree already exists, so a failed write is not worth
// failing the add over: delete falls back to HEAD without it.
func recordBranch(repoPath, branch, session string, reused bool) {
	_ = git.ConfigSet(repoPath, sessionKey(branch), session)
	if reused {
		_ = git.ConfigSet(repoPath, reusedKey(branch), "true")
	}
}

// recordedBranches lists the branches of mainRepoPath whose back-pointer names
// session.
func recordedBranches(mainRepoPath, session string) []string {
	out, err := git.Output(mainRepoPath, "config", "--local", "--get-regexp", `^branch\..*\.seshy-session$`)
	if err != nil {
		return nil
	}
	var branches []string
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(line, " ")
		if !ok || value != session {
			continue
		}
		branch := strings.TrimSuffix(strings.TrimPrefix(key, "branch."), ".seshy-session")
		branches = append(branches, branch)
	}
	return branches
}

// retargetBranchRecords rewrites the back-pointers of every worktree under
// sessionPath from oldName to newName after a rename.
func retargetBranchRecords(sessionPath, oldName, newName string) {
	entries, _ := os.ReadDir(sessionPath)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		mainRepoPath, err := git.MainWorktree(filepath.Join(sessionPath, e.Name()))
		if err != nil {
			continue
		}
		for _, branch := range recordedBranches(mainRepoPath, oldName) {
			_ = git.ConfigSet(mainRepoPath, sessionKey(branch), newName)
		}
	}
}

// branchesToDelete decides which branches go with the worktree at entryPath.
//
// The checked-out branch is read first: when its back-pointer names this
// session it is the one seshy created, whatever else the user did since.
// Otherwise every branch recorded for the session that no other worktree has
// checked out is seshy's, which is what makes a detached tree still drop its
// branch. Sessions from before the back-pointer existed have no record at all,
// and for them HEAD is the only evidence of what seshy created, so it is
// deleted as it always was.
func branchesToDelete(mainRepoPath, entryPath, session string) []string {
	head, _ := git.Branch(entryPath)
	if head == "HEAD" {
		head = ""
	}
	if head != "" {
		if owner, set, _ := git.ConfigGet(mainRepoPath, sessionKey(head)); set && owner == session {
			return []string{head}
		}
	}

	recorded := recordedBranches(mainRepoPath, session)
	if len(recorded) == 0 {
		// Legacy session: no record was ever written, so the checked-out branch
		// is the only evidence of what seshy created.
		if head == "" {
			return nil
		}
		return []string{head}
	}

	elsewhere := map[string]bool{}
	if trees, err := git.Worktrees(mainRepoPath); err == nil {
		for _, tree := range trees {
			if tree.Branch != "" && realPath(tree.Path) != realPath(entryPath) {
				elsewhere[strings.TrimPrefix(tree.Branch, "refs/heads/")] = true
			}
		}
	}
	var branches []string
	for _, branch := range recorded {
		if !elsewhere[branch] {
			branches = append(branches, branch)
		}
	}
	return branches
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

	// The session is the directory the entry sits in, for archived and live
	// sessions alike.
	branches := branchesToDelete(mainRepoPath, entryPath, filepath.Base(filepath.Dir(entryPath)))

	if err := removeWorktree(mainRepoPath, entryPath, force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	mainBranch, _ := git.Branch(mainRepoPath)
	var errs []error
	for _, branch := range branches {
		// The main worktree may sit on the same branch; git refuses to delete a
		// checked-out branch and the user did not ask us to move it.
		if branch == mainBranch {
			continue
		}
		if err := git.Run(mainRepoPath, "branch", "-D", branch); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete branch %s: %w", branch, err))
		}
	}
	return errors.Join(errs...)
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
