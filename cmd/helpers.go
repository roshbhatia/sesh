package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/roshbhatia/go-utils/paths"
	"github.com/roshbhatia/go-utils/terminal"
	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var (
	errCancelled       = errors.New("selection cancelled")
	errNothingSelected = errors.New("nothing selected")
)

// runSource executes a command and returns stdout lines.
func runSource(command string) ([]string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("repo source command failed: %v\n  command: %s", err, command)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}

// runPicker executes a picker command, piping input to stdin and reading selections from stdout.
func runPicker(command string, input []string) ([]string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = strings.NewReader(strings.Join(input, "\n") + "\n")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		// fzf exits 130 on ctrl-c, 1 on no match
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code := exitErr.ExitCode()
			if code == 130 || code == 1 {
				return nil, errCancelled
			}
		}
		return nil, fmt.Errorf("picker command failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	if len(result) == 0 {
		return nil, errNothingSelected
	}
	return result, nil
}

// skipConfirm reports whether a destructive command may skip its prompt. Both
// --yes and --force skip it; --force also overrides a cleanup failure, so when
// a caller passed --force it is told once that --yes alone would skip the
// prompt without forcing.
func skipConfirm(yes, force bool) bool {
	if force {
		ui.Diagnostic(os.Stderr, "warning", "--force implies --yes; pass --yes to skip the prompt without forcing")
	}
	return yes || force
}

// confirm asks question on stderr and reads one line from stdin. force skips
// the question. When stdin is not a terminal there is no one to answer, so
// confirm refuses instead of reading EOF as "no" and exiting 0 with nothing
// done, which a script would take for success. An explicit "no" is reported
// as ok == false with a nil error, and the caller exits quietly.
func confirm(question string, force bool) (ok bool, err error) {
	if force {
		return true, nil
	}
	if !terminal.IsTTY(os.Stdin) {
		return false, exitcode.Refusedf("refusing to prompt; stdin is not a terminal (pass --force)")
	}
	fmt.Fprintf(os.Stderr, "%s [y/N] ", question)
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Fprintln(os.Stderr, ui.Info("Cancelled."))
		return false, nil
	}
	return true, nil
}

// readRepoArgs returns the repo paths named on the command line, plus one per
// stdin line when --stdin is set or an argument is "-". fromStdin reports that
// stdin was consulted, so an empty list from it means "no repos" rather than
// "ask the picker".
func readRepoArgs(args []string, useStdin bool, stdin io.Reader) (repos []string, fromStdin bool) {
	for _, arg := range args {
		if arg == "-" {
			useStdin = true
			continue
		}
		repos = append(repos, arg)
	}
	if !useStdin {
		return repos, false
	}
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			repos = append(repos, line)
		}
	}
	return repos, true
}

// prependDefaults adds default repos to the front of candidates, deduplicating.
func prependDefaults(defaults, candidates []string) []string {
	if len(defaults) == 0 {
		return candidates
	}
	seen := make(map[string]bool, len(defaults))
	result := make([]string, 0, len(defaults)+len(candidates))
	for _, d := range defaults {
		d = paths.ExpandHome(d)
		if !seen[d] {
			seen[d] = true
			result = append(result, d)
		}
	}
	for _, c := range candidates {
		if !seen[c] {
			seen[c] = true
			result = append(result, c)
		}
	}
	return result
}

// warnReusedBranches surfaces git's DWIM: a worktree that landed on a branch
// which already existed did not get the fresh branch the caller asked for.
func warnReusedBranches(repos []session.RepoInfo) {
	for _, repo := range repos {
		if repo.Reused {
			ui.Diagnostic(os.Stderr, "warning", fmt.Sprintf("reusing branch '%s'", repo.Branch))
		}
	}
}

// sessionNames extracts the names from a session list.
func sessionNames(sessions []session.Session) []string {
	names := make([]string, len(sessions))
	for i, s := range sessions {
		names[i] = s.Name
	}
	return names
}

// selectSessionName resolves the session a command should act on: args[0] when
// given, otherwise a pick from names via the configured session picker.
// An empty name with a nil error means there was nothing to pick or the user
// backed out, and the caller should exit quietly.
func selectSessionName(args []string, picker string, names []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if len(names) == 0 {
		return "", nil
	}
	selected, err := runPicker(picker, names)
	if err != nil {
		if errors.Is(err, errCancelled) || errors.Is(err, errNothingSelected) {
			return "", nil
		}
		return "", err
	}
	return selected[0], nil
}

// completeSessionNames completes the first argument with live session names.
func completeSessionNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	sessions, _ := session.List()
	return sessionNames(sessions), cobra.ShellCompDirectiveNoFileComp
}

// completeArchivedNames completes the first argument with archived session names.
func completeArchivedNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	sessions, _ := session.ListArchived()
	return sessionNames(sessions), cobra.ShellCompDirectiveNoFileComp
}
