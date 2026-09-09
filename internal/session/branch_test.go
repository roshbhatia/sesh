package session

import (
	"errors"
	"fmt"
	"os/user"
	"strings"
	"testing"

	"github.com/roshbhatia/go-utils/git"
)

func TestRenderBranchNameDefault(t *testing.T) {
	name, err := RenderBranchName("sy/{{.Session}}/{{.Repo}}", "my-feat", "my-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "sy/my-feat/my-api" {
		t.Errorf("expected 'sy/my-feat/my-api', got %q", name)
	}
}

func TestRenderBranchNameCustomTemplate(t *testing.T) {
	name, err := RenderBranchName("feature/{{.Repo}}-{{.Session}}", "v2", "core")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "feature/core-v2" {
		t.Errorf("expected 'feature/core-v2', got %q", name)
	}
}

func TestRenderBranchNameUserVariable(t *testing.T) {
	name, err := RenderBranchName("dev/{{.User}}/{{.Repo}}", "sess", "api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	u, _ := user.Current()
	if !strings.HasPrefix(name, "dev/"+u.Username+"/") {
		t.Errorf("expected user prefix, got %q", name)
	}
}

func TestRenderBranchNameInvalidTemplate(t *testing.T) {
	_, err := RenderBranchName("sy/{{.Bad", "s", "r")
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestRenderBranchNameRendersInvalidName(t *testing.T) {
	_, err := RenderBranchName("sy/{{.Session}} {{.Repo}}", "my feat", "api")
	if err == nil {
		t.Error("expected error for branch name with spaces")
	}
}

func TestCheckBranchNameValid(t *testing.T) {
	for _, name := range []string{"main", "feature/x", "sy/sess/repo", "a-b_c.d"} {
		if err := CheckBranchName(name); err != nil {
			t.Errorf("expected valid for %q, got error: %v", name, err)
		}
	}
}

// TestCheckBranchNameRejectsWhatGitRejects covers the rules the hand-rolled
// validator used to carry, plus the ones it missed (HEAD, a leading dash).
func TestCheckBranchNameRejectsWhatGitRejects(t *testing.T) {
	for _, name := range []string{"", "my branch", "a..b", "a~b", "a^b", "a:b", "a?b", "a*b", "a[b", "a\\b", "branch.lock", "trailing.", "HEAD", "-x", "@{1}"} {
		err := CheckBranchName(name)
		if err == nil {
			t.Errorf("expected %q to be rejected", name)
			continue
		}
		var bad *BranchNameError
		if !errors.As(err, &bad) {
			t.Errorf("%q: expected *BranchNameError, got %T", name, err)
		}
		if want := fmt.Sprintf("'%s' is not a valid branch name", name); err.Error() != want {
			t.Errorf("%q: message = %q, want %q", name, err.Error(), want)
		}
		if git.ExitStatus(err) != 128 {
			t.Errorf("%q: exit status = %d, want 128", name, git.ExitStatus(err))
		}
	}
}
