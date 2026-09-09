package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roshbhatia/go-utils/git"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
)

func TestExitStatusOfSentinels(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, 0},
		{errors.New("plain"), 1},
		{exitcode.Usagef("bad"), 2},
		{fmt.Errorf("wrapped: %w", exitcode.NotFoundf("gone")), 3},
		{exitcode.Refusedf("no"), 4},
	}
	for _, tc := range cases {
		if got := exitcode.Of(tc.err); got != tc.want {
			t.Errorf("Of(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}

// TestMessageGitFailureIsGitsStderr pins the recovery of git's own words from
// the error go-utils/git builds. A change to that format shows up here, not as
// a garbled fatal line for a user.
func TestMessageGitFailureIsGitsStderr(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "r")
	setupGitRepo(t, repo)
	_, err := git.Output(repo, "rev-parse", "--verify", "refs/heads/nope")
	if err == nil {
		t.Fatal("expected rev-parse to fail")
	}
	wrapped := fmt.Errorf("adding repo: %w", err)

	if got := exitcode.Of(wrapped); got != 128 {
		t.Errorf("exit status = %d, want 128", got)
	}
	msg := Message(wrapped)
	if strings.Contains(msg, "failed in") || strings.HasPrefix(msg, "fatal: ") || strings.Contains(msg, "exit status") {
		t.Errorf("message is not git's stderr: %q", msg)
	}
	if msg == "" {
		t.Error("message is empty")
	}
}

func TestMessageBranchNameError(t *testing.T) {
	err := session.CheckBranchName("HEAD")
	if got := Message(fmt.Errorf("creating session: %w", err)); got != "'HEAD' is not a valid branch name" {
		t.Errorf("message = %q", got)
	}
	if got := exitcode.Of(err); got != 128 {
		t.Errorf("exit status = %d, want 128", got)
	}
}

func TestFatalMessageShape(t *testing.T) {
	cases := map[string]string{
		"Something happened.":      "something happened",
		"fatal: already lowercase": "already lowercase",
		"SESHY_CONFIG is required": "SESHY_CONFIG is required",
		"'x' is odd":               "'x' is odd",
	}
	for in, want := range cases {
		if got := fatalMessage(in); got != want {
			t.Errorf("fatalMessage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestReportWritesOneFatalLine(t *testing.T) {
	var out bytes.Buffer
	Report(&out, exitcode.NotFoundf("session 'x' not found"))
	if out.String() != "fatal: session 'x' not found\n" {
		t.Errorf("report = %q", out.String())
	}

	out.Reset()
	Report(&out, exitcode.Quiet(exitcode.NotFoundf("silent")))
	if out.Len() != 0 {
		t.Errorf("quiet error printed %q", out.String())
	}

	out.Reset()
	Report(&out, exitcode.Usagef("unknown flag: --x"))
	if got := out.String(); got != "fatal: unknown flag: --x\nhint: run 'sy --help' for usage\n" {
		t.Errorf("usage report = %q", got)
	}
}

func TestCobraErrorsBeforePreRunAreUsage(t *testing.T) {
	isolatedRoot(t)
	for _, args := range [][]string{{"list", "--bogus"}, {"path"}, {"frob"}, {"new", "x", "--empty", "/tmp"}} {
		_, _, err := runCmd(args...)
		if !errors.Is(err, exitcode.ErrUsage) {
			t.Errorf("%v: expected ErrUsage, got %v", args, err)
		}
	}
	if _, _, err := runCmd("path", "nope"); errors.Is(err, exitcode.ErrUsage) || !errors.Is(err, exitcode.ErrNotFound) {
		t.Errorf("a runtime not-found must not read as usage: %v", err)
	}
}
