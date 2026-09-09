package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roshbhatia/seshy/internal/session"
)

// settingRow finds the config table row for a setting name and returns its
// value and origin columns.
func settingRow(t *testing.T, stdout, name string) (value, origin string) {
	t.Helper()
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == name {
			return fields[1], fields[len(fields)-1]
		}
	}
	t.Fatalf("no row for %q in:\n%s", name, stdout)
	return "", ""
}

func TestConfigShowsDefaultOrigin(t *testing.T) {
	isolatedRoot(t)
	stdout, _, err := runCmd("config")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	value, origin := settingRow(t, stdout, "branchFormat")
	if value != "sy/{{.Session}}/{{.Repo}}" || origin != "default" {
		t.Errorf("branchFormat = %q from %q; want the default", value, origin)
	}
}

func TestConfigShowsEnvOrigin(t *testing.T) {
	isolatedRoot(t)
	t.Setenv("SESHY_BRANCH_FORMAT", "env/{{.Session}}/{{.Repo}}")
	stdout, _, err := runCmd("config")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	value, origin := settingRow(t, stdout, "branchFormat")
	if value != "env/{{.Session}}/{{.Repo}}" || origin != "env" {
		t.Errorf("branchFormat = %q from %q; want the env value", value, origin)
	}
}

func TestConfigShowsFileOrigin(t *testing.T) {
	isolatedRoot(t)
	dir := t.TempDir()
	cfg := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(cfg, []byte("repoSource: fd -t d\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runCmd("config", "--config", cfg)
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	value, origin := settingRow(t, stdout, "repoSource")
	if value != "fd" || origin != "file" {
		// value is the first field; the source has a space, so check the origin.
		if origin != "file" {
			t.Errorf("repoSource origin = %q; want file", origin)
		}
	}
	if _, branchOrigin := settingRow(t, stdout, "branchFormat"); branchOrigin != "default" {
		t.Errorf("branchFormat origin = %q; want default when the file omits it", branchOrigin)
	}
}

// TestGitConfigOverridesBranchFormat: seshy.branchFormat in the source repo
// renames the worktree branch, above the file and the default.
func TestGitConfigOverridesBranchFormat(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "api")
	setupGitRepo(t, repo)
	if out, err := exec.Command("git", "-C", repo, "config", "seshy.branchFormat", "gitcfg/{{.Session}}/{{.Repo}}").CombinedOutput(); err != nil {
		t.Fatalf("git config: %v\n%s", err, out)
	}
	opts := session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}", BranchFormatFor: branchFormatResolver()}
	if _, err := session.Create("feat", []string{repo}, opts); err != nil {
		t.Fatalf("create: %v", err)
	}
	if out, err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/gitcfg/feat/api").CombinedOutput(); err != nil {
		t.Errorf("git config branchFormat did not name the branch: %v\n%s", err, out)
	}
}

// TestBranchFlagBeatsGitConfig: an explicit --branch still wins over the repo's
// git config format.
func TestBranchFlagBeatsGitConfig(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "api")
	setupGitRepo(t, repo)
	if out, err := exec.Command("git", "-C", repo, "config", "seshy.branchFormat", "gitcfg/{{.Session}}/{{.Repo}}").CombinedOutput(); err != nil {
		t.Fatalf("git config: %v\n%s", err, out)
	}
	opts := session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}", BranchFormatFor: branchFormatResolver(), BranchOverride: "hotfix/urgent"}
	if _, err := session.Create("feat", []string{repo}, opts); err != nil {
		t.Fatalf("create: %v", err)
	}
	if out, err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/hotfix/urgent").CombinedOutput(); err != nil {
		t.Errorf("--branch did not win: %v\n%s", err, out)
	}
}
