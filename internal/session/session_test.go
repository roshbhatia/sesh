package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// isolatedRoot sets XDG_STATE_HOME to a temp dir scoped to this test and
// returns a cleanup function.
func isolatedRoot(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)
	return tmp
}

func setupTestGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "config", "user.email", "test@example.com"},
		{"git", "-C", dir, "config", "user.name", "Test User"},
	}
	for _, args := range cmds {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	for _, args := range [][]string{
		{"git", "-C", dir, "add", "."},
		{"git", "-C", dir, "commit", "-m", "Initial commit"},
	} {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateSessionName
// ---------------------------------------------------------------------------

func TestValidateSessionName(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"simple", "test", false},
		{"hyphen", "test-session", false},
		{"underscore", "test_session", false},
		{"numbers", "test123", false},
		{"uppercase", "MySession", false},
		{"mixed", "My-Session_2", false},
		{"empty", "", true},
		{"spaces", "test session", true},
		{"slash", "test/session", true},
		{"at sign", "test@session", true},
		{"dot", "test.session", true},
		{"leading hyphen", "-bad", false}, // hyphens allowed anywhere
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSessionName(tc.input)
			if (err != nil) != tc.wantError {
				t.Errorf("ValidateSessionName(%q) error=%v, wantError=%v", tc.input, err, tc.wantError)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetRepoBasename
// ---------------------------------------------------------------------------

func TestGetRepoBasename(t *testing.T) {
	cases := []struct {
		path, want string
	}{
		{"/home/user/repos/myrepo", "myrepo"},
		{"/home/user/repos/myrepo/", "myrepo"},
		{"myrepo", "myrepo"},
		{"/a/b/c", "c"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := GetRepoBasename(tc.path)
			if got != tc.want {
				t.Errorf("GetRepoBasename(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsGitRepo
// ---------------------------------------------------------------------------

func TestIsGitRepo(t *testing.T) {
	tmp := t.TempDir()

	t.Run("git repo", func(t *testing.T) {
		dir := filepath.Join(tmp, "repo")
		setupTestGitRepo(t, dir)
		if !IsGitRepo(dir) {
			t.Error("expected true for git repo")
		}
	})

	t.Run("plain dir", func(t *testing.T) {
		dir := filepath.Join(tmp, "plain")
		os.MkdirAll(dir, 0755)
		if IsGitRepo(dir) {
			t.Error("expected false for plain dir")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		if IsGitRepo(filepath.Join(tmp, "does-not-exist")) {
			t.Error("expected false for nonexistent path")
		}
	})
}

// ---------------------------------------------------------------------------
// CreateSymlink
// ---------------------------------------------------------------------------

func TestCreateSymlink(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "target")
	os.MkdirAll(target, 0755)
	sessionDir := filepath.Join(tmp, "session")
	os.MkdirAll(sessionDir, 0755)

	linkPath, err := CreateSymlink(target, sessionDir)
	if err != nil {
		t.Fatalf("CreateSymlink: %v", err)
	}

	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink, got regular entry")
	}

	resolved, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if resolved != target {
		t.Errorf("symlink target = %q, want %q", resolved, target)
	}
}

func TestCreateSymlinkNamingConvention(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "my-cool-repo")
	os.MkdirAll(target, 0755)
	sessionDir := filepath.Join(tmp, "sess")
	os.MkdirAll(sessionDir, 0755)

	linkPath, err := CreateSymlink(target, sessionDir)
	if err != nil {
		t.Fatalf("CreateSymlink: %v", err)
	}
	if filepath.Base(linkPath) != "my-cool-repo" {
		t.Errorf("expected link name 'my-cool-repo', got %q", filepath.Base(linkPath))
	}
}

func TestCreateSymlinkDuplicate(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "target")
	os.MkdirAll(target, 0755)
	sessionDir := filepath.Join(tmp, "sess")
	os.MkdirAll(sessionDir, 0755)

	if _, err := CreateSymlink(target, sessionDir); err != nil {
		t.Fatalf("first CreateSymlink: %v", err)
	}
	_, err := CreateSymlink(target, sessionDir)
	if err == nil {
		t.Error("expected error on duplicate symlink, got nil")
	}
}

// ---------------------------------------------------------------------------
// CreateWorktree
// ---------------------------------------------------------------------------

func TestCreateWorktree(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "testrepo")
	setupTestGitRepo(t, repoDir)

	sessionDir := filepath.Join(tmp, "sessions", "my-session")
	os.MkdirAll(sessionDir, 0755)

	worktreePath, err := CreateWorktree(repoDir, sessionDir, "my-session")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}
	if !IsGitRepo(worktreePath) {
		t.Error("worktree is not a git repo")
	}
	if filepath.Base(worktreePath) != "testrepo-my-session" {
		t.Errorf("expected worktree name 'testrepo-my-session', got %q", filepath.Base(worktreePath))
	}
}

func TestCreateWorktreeNonGitRepo(t *testing.T) {
	tmp := t.TempDir()
	plain := filepath.Join(tmp, "plain")
	os.MkdirAll(plain, 0755)
	sessionDir := filepath.Join(tmp, "sess")
	os.MkdirAll(sessionDir, 0755)

	_, err := CreateWorktree(plain, sessionDir, "sess")
	if err == nil {
		t.Error("expected error when source is not a git repo")
	}
}

// ---------------------------------------------------------------------------
// CleanupWorktrees
// ---------------------------------------------------------------------------

func TestCleanupWorktrees(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "testrepo")
	setupTestGitRepo(t, repoDir)

	sessionDir := filepath.Join(tmp, "session")
	os.MkdirAll(sessionDir, 0755)

	worktreePath, err := CreateWorktree(repoDir, sessionDir, "test")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Fatal("worktree was not created")
	}

	if err := CleanupWorktrees(sessionDir); err != nil {
		t.Fatalf("CleanupWorktrees: %v", err)
	}

	exec.Command("git", "-C", repoDir, "worktree", "prune").Run()
	out, _ := exec.Command("git", "-C", repoDir, "worktree", "list").Output()
	if strings.Contains(string(out), worktreePath) {
		t.Errorf("worktree still registered after cleanup + prune:\n%s", out)
	}
}

func TestCleanupWorktreesEmptyDir(t *testing.T) {
	tmp := t.TempDir()
	empty := filepath.Join(tmp, "empty")
	os.MkdirAll(empty, 0755)
	// Should not error on a directory with no worktrees
	if err := CleanupWorktrees(empty); err != nil {
		t.Errorf("CleanupWorktrees on empty dir: %v", err)
	}
}

func TestCleanupWorktreesNonexistentDir(t *testing.T) {
	err := CleanupWorktrees("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// ---------------------------------------------------------------------------
// GetWorktreeMainRepo
// ---------------------------------------------------------------------------

func TestGetWorktreeMainRepo(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "testrepo")
	setupTestGitRepo(t, repoDir)

	sessionDir := filepath.Join(tmp, "session")
	os.MkdirAll(sessionDir, 0755)

	worktreePath, err := CreateWorktree(repoDir, sessionDir, "test")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	mainRepo, err := GetWorktreeMainRepo(worktreePath)
	if err != nil {
		t.Fatalf("GetWorktreeMainRepo: %v", err)
	}
	// The resolved main repo should point somewhere inside or equal to repoDir
	if mainRepo == "" {
		t.Error("expected non-empty main repo path")
	}
}

func TestGetWorktreeMainRepoNonGit(t *testing.T) {
	tmp := t.TempDir()
	_, err := GetWorktreeMainRepo(tmp)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

// ---------------------------------------------------------------------------
// Exists / GetPath
// ---------------------------------------------------------------------------

func TestExistsAndGetPath(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "r")
	setupTestGitRepo(t, repoDir)

	if err := Create("exist-test", []string{repoDir}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !Exists("exist-test") {
		t.Error("expected Exists to return true")
	}

	path, err := GetPath("exist-test")
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestExistsFalse(t *testing.T) {
	isolatedRoot(t)
	if Exists("no-such-session") {
		t.Error("expected Exists to return false for missing session")
	}
}

func TestGetPathNotFound(t *testing.T) {
	isolatedRoot(t)
	_, err := GetPath("no-such-session")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreateWithGitRepo(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "myrepo")
	setupTestGitRepo(t, repoDir)

	if err := Create("create-git", []string{repoDir}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	sessionPath, _ := GetPath("create-git")
	worktree := filepath.Join(sessionPath, "myrepo-create-git")
	if _, err := os.Stat(worktree); os.IsNotExist(err) {
		t.Error("expected worktree to be created")
	}
}

func TestCreateWithNonGitDir(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	plainDir := filepath.Join(tmp, "plain")
	os.MkdirAll(plainDir, 0755)

	if err := Create("create-plain", []string{plainDir}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	sessionPath, _ := GetPath("create-plain")
	link := filepath.Join(sessionPath, "plain")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("Lstat link: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink for non-git dir")
	}
}

func TestCreateDuplicateErrors(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "r")
	setupTestGitRepo(t, repoDir)

	if err := Create("dup-session", []string{repoDir}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if err := Create("dup-session", []string{repoDir}); err == nil {
		t.Error("expected error on duplicate Create")
	}
}

func TestCreateInvalidName(t *testing.T) {
	isolatedRoot(t)
	if err := Create("bad name!", []string{}); err == nil {
		t.Error("expected error for invalid session name")
	}
}

func TestCreateMultipleRepos(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repo1 := filepath.Join(tmp, "repo1")
	repo2 := filepath.Join(tmp, "repo2")
	setupTestGitRepo(t, repo1)
	setupTestGitRepo(t, repo2)

	if err := Create("multi", []string{repo1, repo2}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	sessionPath, _ := GetPath("multi")
	for _, name := range []string{"repo1-multi", "repo2-multi"} {
		if _, err := os.Stat(filepath.Join(sessionPath, name)); os.IsNotExist(err) {
			t.Errorf("expected %s worktree to exist", name)
		}
	}
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestListEmpty(t *testing.T) {
	isolatedRoot(t)
	sessions, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestListMultiple(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "r")
	setupTestGitRepo(t, repoDir)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := Create(name, []string{repoDir}); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	sessions, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}

	names := map[string]bool{}
	for _, s := range sessions {
		names[s.Name] = true
		if s.Path == "" {
			t.Errorf("session %q has empty path", s.Name)
		}
		if s.LastModified.IsZero() {
			t.Errorf("session %q has zero LastModified", s.Name)
		}
	}
	for _, n := range []string{"alpha", "beta", "gamma"} {
		if !names[n] {
			t.Errorf("session %q missing from list", n)
		}
	}
}

func TestListRepoCount(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repo1 := filepath.Join(tmp, "r1")
	repo2 := filepath.Join(tmp, "r2")
	setupTestGitRepo(t, repo1)
	setupTestGitRepo(t, repo2)

	if err := Create("count-test", []string{repo1, repo2}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	sessions, _ := List()
	for _, s := range sessions {
		if s.Name == "count-test" && s.RepoCount != 2 {
			t.Errorf("expected RepoCount=2, got %d", s.RepoCount)
		}
	}
}

// ---------------------------------------------------------------------------
// AddRepos
// ---------------------------------------------------------------------------

func TestAddRepos(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repo1 := filepath.Join(tmp, "repo1")
	repo2 := filepath.Join(tmp, "repo2")
	setupTestGitRepo(t, repo1)
	setupTestGitRepo(t, repo2)

	if err := Create("add-test", []string{repo1}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := AddRepos("add-test", []string{repo2}); err != nil {
		t.Fatalf("AddRepos: %v", err)
	}

	sessionPath, _ := GetPath("add-test")
	for _, wt := range []string{"repo1-add-test", "repo2-add-test"} {
		if _, err := os.Stat(filepath.Join(sessionPath, wt)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist after AddRepos", wt)
		}
	}
}

func TestAddReposSessionNotFound(t *testing.T) {
	isolatedRoot(t)
	if err := AddRepos("no-such", []string{"/tmp"}); err == nil {
		t.Error("expected error adding to nonexistent session")
	}
}

func TestAddReposNonGitDir(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "r")
	setupTestGitRepo(t, repoDir)
	plainDir := filepath.Join(tmp, "plain")
	os.MkdirAll(plainDir, 0755)

	if err := Create("add-plain", []string{repoDir}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := AddRepos("add-plain", []string{plainDir}); err != nil {
		t.Fatalf("AddRepos with plain dir: %v", err)
	}

	sessionPath, _ := GetPath("add-plain")
	link := filepath.Join(sessionPath, "plain")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("Lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink for non-git dir in AddRepos")
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "r")
	setupTestGitRepo(t, repoDir)

	if err := Create("del-test", []string{repoDir}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := Delete("del-test"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if Exists("del-test") {
		t.Error("expected session to be gone after Delete")
	}
}

func TestDeleteNotFound(t *testing.T) {
	isolatedRoot(t)
	if err := Delete("ghost"); err == nil {
		t.Error("expected error deleting nonexistent session")
	}
}

func TestDeleteCleansUpWorktrees(t *testing.T) {
	isolatedRoot(t)
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "r")
	setupTestGitRepo(t, repoDir)

	if err := Create("wt-del", []string{repoDir}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	sessionPath, _ := GetPath("wt-del")
	worktreePath := filepath.Join(sessionPath, "r-wt-del")

	if err := Delete("wt-del"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// worktree directory should be gone
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree directory still exists after Delete")
	}

	// git worktree prune removes stale entries; after that the path should not appear
	exec.Command("git", "-C", repoDir, "worktree", "prune").Run()
	out, _ := exec.Command("git", "-C", repoDir, "worktree", "list").Output()
	if strings.Contains(string(out), worktreePath) {
		t.Errorf("worktree still registered after Delete + prune:\n%s", out)
	}
}
