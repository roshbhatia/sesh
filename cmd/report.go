package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/roshbhatia/go-utils/git"
	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
)

// Report writes the terminating error to w the way git does: one "fatal:"
// line, lowercase, without a trailing period. A git failure is reported with
// git's own stderr, so the user reads what git said rather than a paraphrase;
// any lines git printed before its fatal line are written first, verbatim.
// A usage error also gets a hint at the help text.
func Report(w io.Writer, err error) {
	if err == nil || exitcode.IsQuiet(err) {
		return
	}
	for _, line := range leadingLines(err) {
		fmt.Fprintln(w, line)
	}
	ui.Fatal(w, Message(err))
	if exitcode.Of(err) == exitcode.Usage {
		ui.Hint(w, "run 'sy --help' for usage")
	}
}

// Message is the one line that describes err after a "fatal:" or "error:"
// label: git's own last diagnostic for a git failure, otherwise err's text
// lowercased and without a trailing period.
func Message(err error) string {
	var branch *session.BranchNameError
	if errors.As(err, &branch) {
		return branch.Error()
	}
	if stderr, ok := gitStderr(err); ok {
		lines := strings.Split(stderr, "\n")
		if i := lastDiagnostic(lines); i >= 0 {
			_, rest, _ := strings.Cut(lines[i], ": ")
			return rest
		}
		return stderr
	}
	return fatalMessage(err.Error())
}

// leadingLines is what git printed before its diagnostic, such as
// "Preparing worktree", which is worth reading but is not the message.
func leadingLines(err error) []string {
	stderr, ok := gitStderr(err)
	if !ok {
		return nil
	}
	lines := strings.Split(stderr, "\n")
	if i := lastDiagnostic(lines); i > 0 {
		return lines[:i]
	}
	return nil
}

// lastDiagnostic finds the last line git labelled "fatal:" or "error:".
func lastDiagnostic(lines []string) int {
	for i := len(lines) - 1; i >= 0; i-- {
		if label, _, found := strings.Cut(lines[i], ": "); found && (label == "fatal" || label == "error") {
			return i
		}
	}
	return -1
}

// gitStderr recovers git's stderr from the error go-utils/git builds, which
// reads "git <verb> failed in <dir>: <stderr>: <exit status>". Only an error
// from a git exit qualifies.
func gitStderr(err error) (string, bool) {
	status := git.ExitStatus(err)
	if status < 0 {
		return "", false
	}
	msg := err.Error()
	start := strings.Index(msg, " failed in ")
	if start < 0 {
		return "", false
	}
	rest := msg[start+len(" failed in "):]
	_, stderr, ok := strings.Cut(rest, ": ")
	if !ok {
		return "", false
	}
	stderr = strings.TrimSuffix(stderr, fmt.Sprintf(": exit status %d", status))
	return strings.TrimSpace(stderr), stderr != ""
}

// fatalMessage lowercases the first letter and drops a trailing period, and
// strips a "fatal:" or "error:" label the message already carries so the line
// does not read "fatal: fatal: ...".
func fatalMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	for _, label := range []string{"fatal: ", "error: ", "Error: "} {
		msg = strings.TrimPrefix(msg, label)
	}
	msg = strings.TrimSuffix(msg, ".")
	if r, size := utf8.DecodeRuneInString(msg); size > 0 && unicode.IsUpper(r) {
		next, _ := utf8.DecodeRuneInString(msg[size:])
		// Leave an acronym or a flag name such as "SESHY_CONFIG" alone.
		if !unicode.IsUpper(next) {
			msg = string(unicode.ToLower(r)) + msg[size:]
		}
	}
	return msg
}
