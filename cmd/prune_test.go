package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roshbhatia/seshy/internal/session"
)

// TestPruneDryRunListsOrphanBranch: the porcelain proof. A session directory
// is removed by hand; prune --dry-run names the orphan branch on stderr and
// leaves it in place.
func TestPruneDryRunListsOrphanBranch(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "api")
	setupGitRepo(t, repo)
	if _, err := session.Create("feat", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	sessionPath, _ := session.Resolve("feat")
	if err := os.RemoveAll(sessionPath); err != nil {
		t.Fatal(err)
	}

	_, stderr, err := runCmd("prune", "--dry-run", repo)
	if err != nil {
		t.Fatalf("prune --dry-run: %v", err)
	}
	if !strings.Contains(stderr, "would delete branch sy/feat/api") {
		t.Errorf("dry run did not name the orphan branch:\n%s", stderr)
	}
	if out, err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/sy/feat/api").CombinedOutput(); err != nil {
		t.Errorf("dry run deleted the branch: %v\n%s", err, out)
	}
}

func TestPruneNothingToDoReports(t *testing.T) {
	isolatedRoot(t)
	_, stderr, err := runCmd("prune")
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if !strings.Contains(stderr, "Nothing to prune") {
		t.Errorf("empty prune did not report: %q", stderr)
	}
}
