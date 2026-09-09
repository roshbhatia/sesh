package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roshbhatia/go-utils/git"
	"github.com/roshbhatia/seshy/internal/config"
)

// PruneAction is one thing prune did, or would do under a dry run. Reason
// says why; Target is the branch, worktree, or symlink it applies to.
type PruneAction struct {
	Kind   string // "worktree", "branch", or "symlink"
	Target string // path or branch name
	Repo   string // source repo for a branch or worktree; empty for a symlink
	Reason string
}

func (a PruneAction) String() string {
	switch a.Kind {
	case "branch":
		return fmt.Sprintf("delete branch %s in %s (%s)", a.Target, a.Repo, a.Reason)
	case "worktree":
		return fmt.Sprintf("prune worktree %s (%s)", a.Target, a.Reason)
	case "symlink":
		return fmt.Sprintf("remove dangling symlink %s (%s)", a.Target, a.Reason)
	default:
		return fmt.Sprintf("%s %s", a.Kind, a.Target)
	}
}

// Prune reclaims what a removed session left in git. For each repo it prunes
// worktree registrations whose directories are gone, then deletes the seshy
// branches whose recorded session directory no longer exists. It also removes
// symlinks under the sessions root whose targets are gone. With no repos it
// visits every source repo of every session.
//
// dryRun computes the same actions without taking them. It returns one action
// per thing done or planned, in visit order, and joins any failures.
func Prune(repoArgs []string, dryRun bool) ([]PruneAction, error) {
	repos, err := pruneTargets(repoArgs)
	if err != nil {
		return nil, err
	}

	var actions []PruneAction
	var problems []error

	for _, repo := range repos {
		worktreeActions, err := prunableWorktrees(repo)
		if err != nil {
			problems = append(problems, err)
		}
		actions = append(actions, worktreeActions...)
		if len(worktreeActions) > 0 && !dryRun {
			if err := git.WorktreePrune(repo); err != nil {
				problems = append(problems, fmt.Errorf("prune worktrees in %s: %w", repo, err))
			}
		}

		for _, branch := range orphanBranches(repo) {
			actions = append(actions, PruneAction{Kind: "branch", Target: branch, Repo: repo, Reason: "session directory is gone"})
			if dryRun {
				continue
			}
			if err := git.Run(repo, "branch", "-d", branch); err != nil {
				problems = append(problems, fmt.Errorf("delete branch %s in %s: %w", branch, repo, err))
			}
		}
	}

	symlinkActions, err := danglingSymlinks()
	if err != nil {
		problems = append(problems, err)
	}
	for _, action := range symlinkActions {
		actions = append(actions, action)
		if dryRun {
			continue
		}
		if err := os.Remove(action.Target); err != nil {
			problems = append(problems, fmt.Errorf("remove %s: %w", action.Target, err))
		}
	}

	return actions, errors.Join(problems...)
}

// pruneTargets resolves the repos to visit: each argument's git toplevel, or
// every source repo of every session when none are given. A path that is not
// a git repository has no worktrees or branches to prune and is an error.
func pruneTargets(repoArgs []string) ([]string, error) {
	seen := make(map[string]bool)
	var repos []string
	add := func(path string) {
		clean := realPath(path)
		if !seen[clean] {
			seen[clean] = true
			repos = append(repos, clean)
		}
	}

	if len(repoArgs) > 0 {
		for _, arg := range repoArgs {
			path, err := NormalizeRepoPath(arg)
			if err != nil {
				return nil, err
			}
			if !git.IsRepo(path) {
				return nil, fmt.Errorf("not a git repository: %s", arg)
			}
			add(path)
		}
		return repos, nil
	}

	sessions, err := List()
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		sources, err := ListRepoSources(s.Path)
		if err != nil {
			continue
		}
		for _, source := range sources {
			if git.IsRepo(source) {
				add(source)
			}
		}
	}
	return repos, nil
}

// prunableWorktrees lists the worktree registrations of repo whose directories
// git reports as gone.
func prunableWorktrees(repo string) ([]PruneAction, error) {
	trees, err := git.Worktrees(repo)
	if err != nil {
		return nil, fmt.Errorf("list worktrees in %s: %w", repo, err)
	}
	var actions []PruneAction
	for _, tree := range trees {
		if tree.Prunable {
			reason := tree.PruneReason
			if reason == "" {
				reason = "gone"
			}
			actions = append(actions, PruneAction{Kind: "worktree", Target: tree.Path, Repo: repo, Reason: reason})
		}
	}
	return actions, nil
}

// orphanBranches lists the seshy branches of repo whose recorded session
// directory no longer exists under the sessions root.
func orphanBranches(repo string) []string {
	out, err := git.Output(repo, "config", "--local", "--get-regexp", `^branch\..*\.seshy-session$`)
	if err != nil {
		return nil
	}
	root := config.GetSessionsRoot()
	var branches []string
	for _, line := range strings.Split(out, "\n") {
		key, sessionName, ok := strings.Cut(line, " ")
		if !ok || sessionName == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, sessionName)); !os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(filepath.Join(config.GetArchiveRoot(), sessionName)); !os.IsNotExist(err) {
			continue
		}
		branch := strings.TrimSuffix(strings.TrimPrefix(key, "branch."), ".seshy-session")
		if reused, _, err := git.ConfigGet(repo, reusedKey(branch)); err != nil || reused == "true" {
			continue
		}
		branches = append(branches, branch)
	}
	return branches
}

// danglingSymlinks lists the symlinks under the sessions root whose targets no
// longer exist. Such a link is a repo entry whose source moved or was deleted.
func danglingSymlinks() ([]PruneAction, error) {
	root := config.GetSessionsRoot()
	sessions, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read sessions root: %w", err)
	}
	var actions []PruneAction
	for _, session := range sessions {
		if !session.IsDir() {
			continue
		}
		sessionPath := filepath.Join(root, session.Name())
		entries, err := os.ReadDir(sessionPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink == 0 {
				continue
			}
			entryPath := filepath.Join(sessionPath, entry.Name())
			if _, err := os.Stat(entryPath); err == nil {
				continue
			}
			actions = append(actions, PruneAction{Kind: "symlink", Target: entryPath, Reason: "target is gone"})
		}
	}
	return actions, nil
}
