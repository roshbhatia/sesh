// Package exitcode names the exit statuses sy reports and the sentinel errors
// that select them. A command returns an error wrapping one of the sentinels;
// main maps it to the status and prints the message once.
package exitcode

import "errors"

// ErrNotFound marks a name that resolves to no session, repo, or archive
// entry. Exit status 3.
var ErrNotFound = errors.New("not found")

// ErrRefused marks an action sy declined to take: a confirmation it could
// not ask for, or one the user answered no to. Exit status 4.
var ErrRefused = errors.New("refused")
