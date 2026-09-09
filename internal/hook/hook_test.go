package hook

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	texttemplate "text/template"

	"github.com/roshbhatia/seshy/internal/tmpl"
)

func renderShellFixture(t *testing.T, source string, data any) string {
	t.Helper()
	parsed := texttemplate.Must(texttemplate.New("shell-fixture").Parse(source))
	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, data); err != nil {
		t.Fatalf("render shell fixture: %v", err)
	}
	return rendered.String()
}

func testData(dir string) tmpl.TemplateData {
	return tmpl.TemplateData{
		Session:     "test-sess",
		SessionPath: dir,
		User:        "testuser",
		Repos: []tmpl.RepoData{
			{Name: "api", Path: filepath.Join(dir, "api")},
			{Name: "web", Path: filepath.Join(dir, "web")},
		},
	}
}

func TestRunSimple(t *testing.T) {
	dir := t.TempDir()
	errs := Run("post-create", []string{"echo hello"}, testData(dir), dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestRunWithTemplate(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	cmd := renderShellFixture(t, `echo {{"{{.Session}}"}} > {{.}}`, marker)
	errs := Run("post-create", []string{cmd}, testData(dir), dir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	got, _ := os.ReadFile(marker)
	if s := string(got); s != "test-sess\n" {
		t.Errorf("expected 'test-sess', got %q", s)
	}
}

func TestRunSetsEnv(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "env")
	cmd := renderShellFixture(t, `echo $SESHY_SESSION,$SESHY_REPO_COUNT,$SESHY_EVENT > {{.}}`, marker)
	errs := Run("post-create", []string{cmd}, testData(dir), dir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	got, _ := os.ReadFile(marker)
	if s := string(got); s != "test-sess,2,post-create\n" {
		t.Errorf("unexpected env: %q", s)
	}
}

func TestRunNonFatal(t *testing.T) {
	dir := t.TempDir()
	errs := Run("post-create", []string{"false"}, testData(dir), dir)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestRunMultiple(t *testing.T) {
	dir := t.TempDir()
	m1 := filepath.Join(dir, "m1")
	m2 := filepath.Join(dir, "m2")
	cmds := []string{
		renderShellFixture(t, `touch {{.}}`, m1),
		renderShellFixture(t, `touch {{.}}`, m2),
	}
	errs := Run("post-create", cmds, testData(dir), dir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if _, err := os.Stat(m1); err != nil {
		t.Error("m1 not created")
	}
	if _, err := os.Stat(m2); err != nil {
		t.Error("m2 not created")
	}
}

func TestRunEmpty(t *testing.T) {
	dir := t.TempDir()
	errs := Run("post-create", nil, testData(dir), dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty list, got %v", errs)
	}
}

func TestRunHookScript(t *testing.T) {
	dir := t.TempDir()
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	hooksDir := filepath.Join(cfgDir, "seshy", "hooks")
	os.MkdirAll(hooksDir, 0755)
	marker := filepath.Join(dir, "script-ran")
	script := renderShellFixture(t, `#!/bin/sh
touch {{.}}
`, marker)
	scriptPath := filepath.Join(hooksDir, "post-create")
	os.WriteFile(scriptPath, []byte(script), 0755)

	errs := Run("post-create", nil, testData(dir), dir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Error("hook script did not run")
	}
}

// TestRunDropsInheritedGitDir keeps a caller's GIT_DIR out of the hook. With
// it inherited, every git command in the hook would act on the invoking
// repository instead of the session worktree it runs in.
func TestRunDropsInheritedGitDir(t *testing.T) {
	t.Setenv("GIT_DIR", "/elsewhere/.git")
	t.Setenv("GIT_WORK_TREE", "/elsewhere")
	t.Setenv("SESHY_KEEP_ME", "yes")
	dir := t.TempDir()
	marker := filepath.Join(dir, "env")
	cmd := renderShellFixture(t, `env | grep -E '^(GIT_DIR|GIT_WORK_TREE|GIT_INDEX_FILE|SESHY_KEEP_ME)=' > {{.}}; true`, marker)
	errs := Run("post-create", []string{cmd}, testData(dir), dir)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	got, _ := os.ReadFile(marker)
	if s := string(got); s != "SESHY_KEEP_ME=yes\n" {
		t.Errorf("expected only SESHY_KEEP_ME in the hook env, got %q", s)
	}
}
