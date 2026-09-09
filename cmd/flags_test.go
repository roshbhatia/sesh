package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roshbhatia/seshy/internal/session"
)

// TestYesSkipsDeletePromptWithoutForcing: --yes deletes without a terminal and
// without the warning --force prints.
func TestYesSkipsDeletePrompt(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "gone", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	devnull, _ := os.Open(os.DevNull)
	defer devnull.Close()
	withStdin(t, devnull)

	_, stderr, err := runCmd("delete", "--yes", "gone")
	if err != nil {
		t.Fatalf("delete --yes: %v", err)
	}
	if session.Exists("gone") {
		t.Error("--yes did not delete the session")
	}
	if strings.Contains(stderr, "--force implies --yes") {
		t.Errorf("--yes must not print the force warning:\n%s", stderr)
	}
}

func TestYesShortFlagSkipsDeletePrompt(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "gone", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	devnull, _ := os.Open(os.DevNull)
	defer devnull.Close()
	withStdin(t, devnull)
	if _, _, err := runCmd("delete", "-y", "gone"); err != nil {
		t.Fatalf("delete -y: %v", err)
	}
	if session.Exists("gone") {
		t.Error("-y did not delete the session")
	}
}

// TestForceWarnsItImpliesYes: --force still skips the prompt and says --yes
// would have too, exactly once.
func TestForceWarnsItImpliesYes(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "gone", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	devnull, _ := os.Open(os.DevNull)
	defer devnull.Close()
	withStdin(t, devnull)

	_, stderr, err := runCmd("delete", "--force", "gone")
	if err != nil {
		t.Fatalf("delete --force: %v", err)
	}
	if session.Exists("gone") {
		t.Error("--force did not delete the session")
	}
	const want = "warning: --force implies --yes; pass --yes to skip the prompt without forcing"
	if n := strings.Count(stderr, want); n != 1 {
		t.Errorf("force warning printed %d times, want 1:\n%s", n, stderr)
	}
}

func TestYesSkipsRemovePrompt(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "api")
	setupGitRepo(t, repo)
	if _, err := session.Create("s", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	devnull, _ := os.Open(os.DevNull)
	defer devnull.Close()
	withStdin(t, devnull)

	if _, _, err := runCmd("remove", "--yes", "s", "api"); err != nil {
		t.Fatalf("remove --yes: %v", err)
	}
	sessionPath, _ := session.Resolve("s")
	if len(session.GetSessionRepoInfos(sessionPath)) != 0 {
		t.Error("--yes did not remove the repo")
	}
}

// TestConfigFlagSelectsFile: --config points config at a file the XDG default
// would not find.
func TestConfigFlagSelectsFile(t *testing.T) {
	isolatedRoot(t)
	dir := t.TempDir()
	cfg := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(cfg, []byte("branchFormat: chosen/{{.Session}}/{{.Repo}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runCmd("config", "--config", cfg)
	if err != nil {
		t.Fatalf("config --config: %v", err)
	}
	if !strings.Contains(stdout, "chosen/") {
		t.Errorf("--config did not select the file:\n%s", stdout)
	}
}

// TestDashRepoArgReadsStdin: a "-" argument reads repo paths from stdin.
func TestDashRepoArgReadsStdin(t *testing.T) {
	repos, fromStdin := readRepoArgs([]string{"-"}, false, strings.NewReader("/a\n/b\n"))
	if !fromStdin {
		t.Error("a - argument must mark stdin consulted")
	}
	if strings.Join(repos, ",") != "/a,/b" {
		t.Errorf("repos = %v, want [/a /b]", repos)
	}
}
