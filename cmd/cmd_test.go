package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/roshbhatia/seshy/internal/session"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// isolatedRoot redirects XDG_STATE_HOME to a per-test temp dir.
func isolatedRoot(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
}

func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "config", "user.email", "t@t.com"},
		{"git", "-C", dir, "config", "user.name", "T"},
	}
	for _, a := range cmds {
		if out, err := exec.Command(a[0], a[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", a, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	for _, a := range [][]string{
		{"git", "-C", dir, "add", "."},
		{"git", "-C", dir, "commit", "-m", "init"},
	} {
		if out, err := exec.Command(a[0], a[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", a, err, out)
		}
	}
}

// runCmd executes a cobra command and returns stdout, stderr and error.
// It redirects os.Stdout so that fmt.Println calls in subcommands are captured.
func runCmd(args ...string) (stdout, stderr string, err error) {
	// Capture real os.Stdout via a pipe
	origStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		panic(pipeErr)
	}
	os.Stdout = w

	// Reset persistent flags before each run
	greedyQuery = ""
	forceDelete = false

	rootCmd.SetArgs(args)
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()
	return buf.String(), "", err
}

// ---------------------------------------------------------------------------
// greedyMatch
// ---------------------------------------------------------------------------

func makeSessions(names ...string) []session.Session {
	out := make([]session.Session, len(names))
	for i, n := range names {
		out[i] = session.Session{Name: n, Path: "/sessions/" + n, LastModified: time.Now()}
	}
	return out
}

func TestGreedyMatchExact(t *testing.T) {
	sessions := makeSessions("platform-auth", "platform-core", "infra")
	match := greedyMatch("platform-auth", sessions)
	if match == nil || match.Name != "platform-auth" {
		t.Errorf("expected exact match 'platform-auth', got %v", match)
	}
}

func TestGreedyMatchExactCaseInsensitive(t *testing.T) {
	sessions := makeSessions("Platform-Auth", "platform-core")
	match := greedyMatch("platform-auth", sessions)
	if match == nil || match.Name != "Platform-Auth" {
		t.Errorf("expected case-insensitive match, got %v", match)
	}
}

func TestGreedyMatchPrefix(t *testing.T) {
	sessions := makeSessions("platform-auth", "platform-core", "infra")
	match := greedyMatch("plat", sessions)
	if match == nil {
		t.Fatal("expected prefix match, got nil")
	}
	if !strings.HasPrefix(strings.ToLower(match.Name), "plat") {
		t.Errorf("expected prefix match starting with 'plat', got %q", match.Name)
	}
}

func TestGreedyMatchSubstring(t *testing.T) {
	sessions := makeSessions("my-platform-v2", "infra")
	match := greedyMatch("platform", sessions)
	if match == nil || match.Name != "my-platform-v2" {
		t.Errorf("expected substring match 'my-platform-v2', got %v", match)
	}
}

func TestGreedyMatchExactBeatsPrefix(t *testing.T) {
	sessions := makeSessions("platform-extra", "platform")
	match := greedyMatch("platform", sessions)
	if match == nil || match.Name != "platform" {
		t.Errorf("expected exact match to win over prefix, got %v", match)
	}
}

func TestGreedyMatchPrefixBeatsSubstring(t *testing.T) {
	sessions := makeSessions("my-platform", "plat-something")
	match := greedyMatch("plat", sessions)
	if match == nil || match.Name != "plat-something" {
		t.Errorf("expected prefix match to win over substring, got %v", match)
	}
}

func TestGreedyMatchNoMatch(t *testing.T) {
	sessions := makeSessions("alpha", "beta")
	match := greedyMatch("gamma", sessions)
	if match != nil {
		t.Errorf("expected nil for no match, got %v", match)
	}
}

func TestGreedyMatchEmptyQuery(t *testing.T) {
	sessions := makeSessions("alpha")
	// Empty string is a substring of everything — first result
	match := greedyMatch("", sessions)
	if match == nil {
		t.Error("expected match for empty query (substring of all)")
	}
}

func TestGreedyMatchEmptySessions(t *testing.T) {
	match := greedyMatch("anything", []session.Session{})
	if match != nil {
		t.Error("expected nil for empty sessions slice")
	}
}

func TestGreedyMatchReturnsPointerToSliceElement(t *testing.T) {
	sessions := makeSessions("alpha", "beta")
	match := greedyMatch("alpha", sessions)
	if match == nil {
		t.Fatal("expected non-nil match")
	}
	// Path should be the one we set in makeSessions
	if match.Path != "/sessions/alpha" {
		t.Errorf("path mismatch: %q", match.Path)
	}
}

// ---------------------------------------------------------------------------
// list command
// ---------------------------------------------------------------------------

func TestListCommandNoSessions(t *testing.T) {
	isolatedRoot(t)
	stdout, _, err := runCmd("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "No sessions") {
		t.Errorf("expected 'No sessions' message, got: %q", stdout)
	}
}

func TestListCommandShowsSessions(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)

	if err := session.Create("my-session", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	stdout, _, err := runCmd("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "my-session") {
		t.Errorf("expected 'my-session' in list output, got: %q", stdout)
	}
}

func TestListAlias(t *testing.T) {
	isolatedRoot(t)
	stdout, _, err := runCmd("ls")
	if err != nil {
		t.Fatalf("ls: %v", err)
	}
	if !strings.Contains(stdout, "No sessions") {
		t.Errorf("expected 'No sessions' from ls alias, got: %q", stdout)
	}
}

func TestListShowsHeaders(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("hdr-test", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	stdout, _, _ := runCmd("list")
	for _, hdr := range []string{"SESSION", "REPOS", "MODIFIED"} {
		if !strings.Contains(stdout, hdr) {
			t.Errorf("expected header %q in list output, got: %q", hdr, stdout)
		}
	}
}

// ---------------------------------------------------------------------------
// path command
// ---------------------------------------------------------------------------

func TestPathCommandExists(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("path-test", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	stdout, _, err := runCmd("path", "path-test")
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	path := strings.TrimSpace(stdout)
	if path == "" {
		t.Error("expected non-empty path output")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("path %q does not exist: %v", path, err)
	}
}

func TestPathCommandNotFound(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("path", "no-such-session")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

// ---------------------------------------------------------------------------
// delete command
// ---------------------------------------------------------------------------

func TestDeleteCommandSuccess(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("del-me", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	_, _, err := runCmd("delete", "--force", "del-me")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if session.Exists("del-me") {
		t.Error("session still exists after delete")
	}
}

func TestDeleteCommandNotFound(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("delete", "--force", "no-such")
	if err == nil {
		t.Error("expected error deleting nonexistent session")
	}
}

func TestDeleteAliasRm(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("rm-me", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	_, _, err := runCmd("rm", "--force", "rm-me")
	if err != nil {
		t.Fatalf("rm alias: %v", err)
	}
	if session.Exists("rm-me") {
		t.Error("session still exists after rm")
	}
}

func TestDeleteAliasRemove(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("remove-me", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	_, _, err := runCmd("remove", "--force", "remove-me")
	if err != nil {
		t.Fatalf("remove alias: %v", err)
	}
}

// ---------------------------------------------------------------------------
// new command (argument validation only — picker is interactive)
// ---------------------------------------------------------------------------

func TestNewCommandInvalidName(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("new", "bad name!")
	if err == nil {
		t.Error("expected error for invalid session name")
	}
}

func TestNewCommandDuplicate(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("dup", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	// new command would invoke the interactive picker, so we test the
	// duplicate guard through session.Create directly to avoid TTY requirement
	if err := session.Create("dup", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err == nil {
		t.Error("expected error for duplicate session name")
	}
}

// ---------------------------------------------------------------------------
// version
// ---------------------------------------------------------------------------

func TestVersionFlag(t *testing.T) {
	isolatedRoot(t)
	stdout, _, _ := runCmd("--version")
	if !strings.Contains(stdout, version) {
		t.Errorf("expected version %q in output, got: %q", version, stdout)
	}
}

// ---------------------------------------------------------------------------
// formatRelativeTime
// ---------------------------------------------------------------------------

func TestFormatRelativeTimeJustNow(t *testing.T) {
	got := formatRelativeTime(time.Now())
	if got != "just now" {
		t.Errorf("expected 'just now', got %q", got)
	}
}

func TestFormatRelativeTimeMinutes(t *testing.T) {
	got := formatRelativeTime(time.Now().Add(-5 * time.Minute))
	if !strings.Contains(got, "minute") {
		t.Errorf("expected 'minutes' in output, got %q", got)
	}
}

func TestFormatRelativeTimeOneMinute(t *testing.T) {
	got := formatRelativeTime(time.Now().Add(-61 * time.Second))
	if got != "1 minute ago" {
		t.Errorf("expected '1 minute ago', got %q", got)
	}
}

func TestFormatRelativeTimeHours(t *testing.T) {
	got := formatRelativeTime(time.Now().Add(-3 * time.Hour))
	if !strings.Contains(got, "hour") {
		t.Errorf("expected 'hours' in output, got %q", got)
	}
}

func TestFormatRelativeTimeOneHour(t *testing.T) {
	got := formatRelativeTime(time.Now().Add(-61 * time.Minute))
	if got != "1 hour ago" {
		t.Errorf("expected '1 hour ago', got %q", got)
	}
}

func TestFormatRelativeTimeDays(t *testing.T) {
	got := formatRelativeTime(time.Now().Add(-3 * 24 * time.Hour))
	if !strings.Contains(got, "day") {
		t.Errorf("expected 'days' in output, got %q", got)
	}
}

func TestFormatRelativeTimeOneDay(t *testing.T) {
	got := formatRelativeTime(time.Now().Add(-25 * time.Hour))
	if got != "1 day ago" {
		t.Errorf("expected '1 day ago', got %q", got)
	}
}

func TestFormatRelativeTimeOldDate(t *testing.T) {
	old := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
	got := formatRelativeTime(old)
	if !strings.Contains(got, "2020") {
		t.Errorf("expected year '2020' in output, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// delete --force and cancellation
// ---------------------------------------------------------------------------

func TestDeleteCommandNonTTYRequiresForce(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "r")
	setupGitRepo(t, repo)
	session.Create("force-test", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"})

	// Without --force in non-TTY, should error
	forceDelete = false
	_, _, err := runCmd("delete", "force-test")
	if err == nil {
		t.Error("expected error when deleting without --force in non-TTY")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("expected error to mention --force, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// config command
// ---------------------------------------------------------------------------

func TestConfigCommandShowsDefaults(t *testing.T) {
	isolatedRoot(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stdout, _, err := runCmd("config")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if !strings.Contains(stdout, "branchFormat") {
		t.Errorf("expected 'branchFormat' in config output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "sy/") {
		t.Errorf("expected 'sy/' in branch format, got: %q", stdout)
	}
}
