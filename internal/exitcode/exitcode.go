// Package exitcode names the exit statuses sy reports and the sentinel errors
// that select them. A command returns an error wrapping one of the sentinels;
// main maps it to the status and prints the message once.
package exitcode

import (
	"errors"
	"fmt"

	"github.com/roshbhatia/go-utils/git"
)

// Exit statuses. A git failure exits with git's own status, which is 128 for
// git's errors and 129 for its usage errors.
const (
	Success  = 0
	Failure  = 1
	Usage    = 2
	NotFound = 3
	Refused  = 4
)

// ErrUsage marks an invocation sy could not act on: an unknown flag, the
// wrong number of arguments, flags that exclude each other. Exit status 2.
var ErrUsage = errors.New("usage")

// ErrNotFound marks a name that resolves to no session, repo, or archive
// entry. Exit status 3.
var ErrNotFound = errors.New("not found")

// ErrRefused marks an action sy declined to take: a confirmation it could
// not ask for, or one the user answered no to. Exit status 4.
var ErrRefused = errors.New("refused")

// Of maps err to the status sy exits with.
func Of(err error) int {
	switch {
	case err == nil:
		return Success
	case errors.Is(err, ErrUsage):
		return Usage
	case errors.Is(err, ErrNotFound):
		return NotFound
	case errors.Is(err, ErrRefused):
		return Refused
	}
	if status := git.ExitStatus(err); status > 0 {
		return status
	}
	return Failure
}

// quiet wraps an error that selects an exit status but must print nothing,
// such as "sy current --quiet" outside every session.
type quiet struct{ err error }

func (q quiet) Error() string { return q.err.Error() }
func (q quiet) Unwrap() error { return q.err }

// Quiet returns err marked as one main exits on without a message.
func Quiet(err error) error { return quiet{err} }

// IsQuiet reports whether err was marked by Quiet.
func IsQuiet(err error) bool {
	var q quiet
	return errors.As(err, &q)
}

// marked pairs an error with the sentinel that selects its status. Its
// message is the error's own, so "session 'x' not found" does not grow a
// ": not found" tail the way fmt.Errorf("%w") would give it.
type marked struct{ sentinel, err error }

func (m marked) Error() string   { return m.err.Error() }
func (m marked) Unwrap() []error { return []error{m.sentinel, m.err} }

// Mark returns err answering errors.Is for sentinel while keeping its message.
func Mark(sentinel, err error) error { return marked{sentinel, err} }

// Usagef formats a usage error.
func Usagef(format string, a ...any) error { return Mark(ErrUsage, fmt.Errorf(format, a...)) }

// NotFoundf formats a not-found error.
func NotFoundf(format string, a ...any) error { return Mark(ErrNotFound, fmt.Errorf(format, a...)) }

// Refusedf formats a refusal.
func Refusedf(format string, a ...any) error { return Mark(ErrRefused, fmt.Errorf(format, a...)) }
