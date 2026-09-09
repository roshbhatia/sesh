package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
)

// TestStatusFormatJSONDescribesEveryEntryKind: a worktree, a symlink, and a
// standalone clone each report their kind, and the worktree carries its lock
// state and whether seshy reused its branch.
func TestStatusFormatJSONDescribesEveryEntryKind(t *testing.T) {
	isolatedRoot(t)
	// Sources are reported resolved, and macOS aliases /var to /private/var.
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(tmp, "api")
	setupGitRepo(t, repo)
	plain := filepath.Join(tmp, "notes")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	// The branch exists before the session does, so git reuses it.
	if out, err := exec.Command("git", "-C", repo, "branch", "sy/st/api").CombinedOutput(); err != nil {
		t.Fatalf("git branch: %v\n%s", err, out)
	}
	if _, err := session.Create("st", []string{repo, plain}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	sessionPath, _ := session.Resolve("st")
	if out, err := exec.Command("git", "-C", repo, "worktree", "lock", filepath.Join(sessionPath, "api")).CombinedOutput(); err != nil {
		t.Fatalf("git worktree lock: %v\n%s", err, out)
	}
	clone := filepath.Join(sessionPath, "solo")
	setupGitRepo(t, clone)

	stdout, _, err := runCmd("status", "st", "--format", "json")
	if err != nil {
		t.Fatalf("status --format json: %v", err)
	}
	var got statusJSON
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode: %v\n%s", err, stdout)
	}
	if got.Version != "seshy.status/v1" || got.Name != "st" || got.Path != sessionPath {
		t.Errorf("header = %+v", got)
	}
	byName := map[string]statusRepoJSON{}
	for _, r := range got.Repos {
		byName[r.Name] = r
	}
	if len(byName) != 3 {
		t.Fatalf("want 3 repos, got %d: %s", len(byName), stdout)
	}
	api := byName["api"]
	if api.Kind != "worktree" || api.Branch != "sy/st/api" || !api.Locked || !api.BranchReused || api.Detached {
		t.Errorf("api = %+v; want a locked worktree on a reused branch", api)
	}
	if api.Source != repo || api.Path != filepath.Join(sessionPath, "api") {
		t.Errorf("api paths = %+v", api)
	}
	if notes := byName["notes"]; notes.Kind != "symlink" || notes.Branch != "" || notes.Source != plain {
		t.Errorf("notes = %+v; want a symlink to %s", notes, plain)
	}
	if solo := byName["solo"]; solo.Kind != "clone" || solo.Source != clone || solo.Locked || solo.BranchReused {
		t.Errorf("solo = %+v; want a clone", solo)
	}
}

func TestStatusFormatJSONFreshBranchIsNotReused(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "r")
	setupGitRepo(t, repo)
	if _, err := session.Create("fresh", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	stdout, _, err := runCmd("status", "fresh", "--format", "json")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	var got statusJSON
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Repos) != 1 || got.Repos[0].BranchReused || got.Repos[0].Locked {
		t.Errorf("repos = %+v; want one unlocked worktree on a fresh branch", got.Repos)
	}
}

func TestStatusFormatJSONEmptySessionHasEmptyRepos(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "bare", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	stdout, _, err := runCmd("status", "bare", "--format", "json")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	repos, ok := got["repos"].([]any)
	if !ok || len(repos) != 0 {
		t.Errorf("repos = %v; want an empty array, not null", got["repos"])
	}
}

func TestStatusFormatJSONMissingSessionIsNotFound(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("status", "nope", "--format", "json")
	if !errors.Is(err, exitcode.ErrNotFound) {
		t.Errorf("status nope = %v; want ErrNotFound", err)
	}
}

func TestStatusTableUnchangedWithoutFormat(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "tbl", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	plain, _, err := runCmd("status", "tbl")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	table, _, err := runCmd("status", "tbl", "--format", "table")
	if err != nil {
		t.Fatalf("status --format table: %v", err)
	}
	if plain != table {
		t.Errorf("--format table differs from the default:\n%s\n%s", plain, table)
	}
}
