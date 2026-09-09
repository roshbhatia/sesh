package cmd

import (
	"fmt"
	"strings"

	"github.com/roshbhatia/seshy/internal/exitcode"
)

// Output formats a command can render. Table is the human default; the rest
// are plumbing a script parses.
const (
	formatTable = "table"
	formatJSON  = "json"
	formatNames = "names"
	formatPaths = "paths"
)

// resolveFormat picks the output format from --format and the boolean
// aliases that predate it. An alias and a conflicting --format, or two
// aliases, is a usage error. format is the --format value, or empty when the
// flag was not given; aliases maps each alias flag's format to whether it was
// set. accepted is the list of formats the command renders.
func resolveFormat(format string, aliases map[string]bool, accepted ...string) (string, error) {
	chosen := format
	for _, name := range accepted {
		if !aliases[name] {
			continue
		}
		if chosen != "" && chosen != name {
			return "", exitcode.Usagef("--%s conflicts with --format %s", name, chosen)
		}
		chosen = name
	}
	if chosen == "" {
		return formatTable, nil
	}
	for _, name := range accepted {
		if chosen == name {
			return chosen, nil
		}
	}
	return "", exitcode.Usagef("unknown format %q (expected %s)", chosen, strings.Join(accepted, ", "))
}

// formatUsage is the help text for a --format flag over accepted.
func formatUsage(accepted ...string) string {
	return fmt.Sprintf("Output format: %s", strings.Join(accepted, ", "))
}
