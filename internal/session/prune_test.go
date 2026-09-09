package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roshbhatia/go-utils/git"
)

// isolatePrune points the sessions root at a temp dir for one test.
func isolatePrune(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_STATE_HOME", root)
	// A real config with an absolute sessionsDir would override XDG_STATE_HOME.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return filepath.Join(root, "seshy", "sessions")
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", dir},
		{"-C", dir, "config", "user.email", "t@t.com"},
		{"-C", dir, "config", "user.name", "T"},
	} {
		if err := git.Run("", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := git.Run(dir, "add", "."); err != nil {
		t.Fatal(err)
	}
	if err := git.Run(dir, "commit", "-m", "init"); err != nil {
		t.Fatal(err)
	}
}

// TestPruneDeletesOrphanBranchAfterSessionRemoved: a session's directory is
// removed out from under seshy; prune drops its worktree registration and the
// branch it left behind.
func TestPruneDeletesOrphanBranchAfterSessionRemoved(t *testing.T) {
	isolatePrune(t)
	repo := filepath.Join(t.TempDir(), "api")
	initRepo(t, repo)
	if _, err := Create("feat", []string{repo}, CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	sessionPath, _ := Resolve("feat")
	if !branchExists(t, repo, "sy/feat/api") {
		t.Fatal("branch was not created")
	}
	// Remove the session directory without seshy's cleanup.
	if err := os.RemoveAll(sessionPath); err != nil {
		t.Fatal(err)
	}

	dry, err := Prune([]string{repo}, true)
	if err != nil {
		t.Fatalf("dry prune: %v", err)
	}
	if !hasBranchAction(dry, "sy/feat/api") {
		t.Errorf("dry run did not list the orphan branch: %v", dry)
	}
	if !branchExists(t, repo, "sy/feat/api") {
		t.Error("dry run deleted the branch")
	}

	actions, err := Prune([]string{repo}, false)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if !hasBranchAction(actions, "sy/feat/api") {
		t.Errorf("prune did not act on the orphan branch: %v", actions)
	}
	if branchExists(t, repo, "sy/feat/api") {
		t.Error("prune left the orphan branch behind")
	}
}

// TestPruneKeepsBranchOfLivingSession: a branch whose session still exists is
// left alone.
func TestPruneKeepsBranchOfLivingSession(t *testing.T) {
	isolatePrune(t)
	repo := filepath.Join(t.TempDir(), "api")
	initRepo(t, repo)
	if _, err := Create("live", []string{repo}, CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	actions, err := Prune([]string{repo}, false)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if hasBranchAction(actions, "sy/live/api") {
		t.Errorf("prune touched a living session's branch: %v", actions)
	}
	if !branchExists(t, repo, "sy/live/api") {
		t.Error("prune deleted a living session's branch")
	}
}

// TestPruneRemovesDanglingSymlink: a symlinked non-git entry whose target is
// deleted is a dangling link prune removes.
func TestPruneRemovesDanglingSymlink(t *testing.T) {
	sessionsRoot := isolatePrune(t)
	target := filepath.Join(t.TempDir(), "notes")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Create("scratch", []string{target}, CreateOpts{}); err != nil {
		t.Fatalf("create: %v", err)
	}
	link := filepath.Join(sessionsRoot, "scratch", "notes")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("symlink missing: %v", err)
	}
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}

	dry, err := Prune(nil, true)
	if err != nil {
		t.Fatalf("dry prune: %v", err)
	}
	if !hasSymlinkAction(dry, link) {
		t.Errorf("dry run did not list the dangling symlink: %v", dry)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Error("dry run removed the symlink")
	}

	if _, err := Prune(nil, false); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("prune left the dangling symlink: %v", err)
	}
}

func TestPruneRejectsNonRepoArgument(t *testing.T) {
	isolatePrune(t)
	plain := t.TempDir()
	if _, err := Prune([]string{plain}, true); err == nil {
		t.Error("prune accepted a non-git directory")
	}
}

func hasBranchAction(actions []PruneAction, branch string) bool {
	for _, a := range actions {
		if a.Kind == "branch" && a.Target == branch {
			return true
		}
	}
	return false
}

func hasSymlinkAction(actions []PruneAction, path string) bool {
	for _, a := range actions {
		if a.Kind == "symlink" && a.Target == path {
			return true
		}
	}
	return false
}
