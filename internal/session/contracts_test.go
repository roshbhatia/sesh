package session

import (
	"github.com/roshbhatia/go-utils/git"
	"os"
	"path/filepath"
	"testing"
)

func TestReferenceDoesNotOwnCheckout(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "repo")
	setupTestGitRepo(t, repo)
	before, err := git.Worktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := Create("reference", []string{repo}, CreateOpts{Reference: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 1 || infos[0].Kind != KindSymlink {
		t.Fatalf("members: %+v", infos)
	}
	after, err := git.Worktrees(repo)
	if err != nil || len(before) != len(after) {
		t.Fatalf("worktrees changed: %+v %v", after, err)
	}
	if err := os.WriteFile(filepath.Join(repo, "keep.txt"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Delete("reference", false); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(repo, "keep.txt")); err != nil || string(data) != "keep" {
		t.Fatalf("reference target changed: %q %v", data, err)
	}
}

func TestDeleteRefusesDirtyWorktree(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "repo")
	setupTestGitRepo(t, repo)
	infos, err := Create("dirty", []string{repo}, defaultOpts())
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(infos[0].Path, "untracked.txt")
	if err := os.WriteFile(file, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Delete("dirty", false); err == nil {
		t.Fatal("dirty worktree was removed")
	}
	if data, err := os.ReadFile(file); err != nil || string(data) != "keep" {
		t.Fatalf("work was lost: %q %v", data, err)
	}
}

func TestDeleteRetainsUnmergedCommit(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "repo")
	setupTestGitRepo(t, repo)
	infos, err := Create("unmerged", []string{repo}, defaultOpts())
	if err != nil {
		t.Fatal(err)
	}
	wt := infos[0].Path
	if err := os.WriteFile(filepath.Join(wt, "work.txt"), []byte("work"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := git.Run(wt, "add", "work.txt"); err != nil {
		t.Fatal(err)
	}
	if err := git.Run(wt, "commit", "-m", "work"); err != nil {
		t.Fatal(err)
	}
	head, err := git.Output(wt, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if err := Delete("unmerged", false); err != nil {
		t.Fatal(err)
	}
	retained, err := git.Output(repo, "rev-parse", infos[0].Branch)
	if err != nil || retained != head {
		t.Fatalf("commit lost: %s %v", retained, err)
	}
	if _, err := Prune([]string{repo}, false); err == nil {
		t.Fatal("prune must refuse an unmerged branch")
	}
	if !branchExists(t, repo, infos[0].Branch) {
		t.Fatal("prune removed unmerged work")
	}
}

func TestBranchChoiceIsExplicit(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "repo")
	setupTestGitRepo(t, repo)
	if err := git.Run(repo, "branch", "existing"); err != nil {
		t.Fatal(err)
	}
	if _, err := Create("new", []string{repo}, CreateOpts{BranchOverride: "existing"}); err == nil {
		t.Fatal("new reused an existing branch")
	}
	if !branchExists(t, repo, "existing") {
		t.Fatal("failed creation removed existing branch")
	}
	if _, err := Create("reuse", []string{repo}, CreateOpts{BranchOverride: "existing", ExistingBranch: true}); err != nil {
		t.Fatal(err)
	}
}

func TestStartPointSelectsCommit(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "repo")
	setupTestGitRepo(t, repo)
	first, err := git.Output(repo, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if err := git.Run(repo, "commit", "--allow-empty", "-m", "second"); err != nil {
		t.Fatal(err)
	}
	infos, err := Create("start", []string{repo}, CreateOpts{BranchOverride: "start", StartPoint: "HEAD~1"})
	if err != nil {
		t.Fatal(err)
	}
	head, err := git.Output(infos[0].Path, "rev-parse", "HEAD")
	if err != nil || head != first {
		t.Fatalf("wrong start point: %s %v", head, err)
	}
}

func TestPrunePreservesReusedBranch(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "repo")
	setupTestGitRepo(t, repo)
	if err := git.Run(repo, "branch", "existing"); err != nil {
		t.Fatal(err)
	}
	if _, err := Create("reuse", []string{repo}, CreateOpts{BranchOverride: "existing", ExistingBranch: true}); err != nil {
		t.Fatal(err)
	}
	if err := Delete("reuse", false); err != nil {
		t.Fatal(err)
	}
	if _, err := Prune([]string{repo}, false); err != nil {
		t.Fatal(err)
	}
	if !branchExists(t, repo, "existing") {
		t.Fatal("prune removed a pre-existing branch")
	}
}
