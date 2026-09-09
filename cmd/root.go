package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

const version = "4.1.0"

var greedyQuery string

// preRunReached records that cobra finished parsing flags and validating
// arguments for the invoked command. An error returned before that point is a
// usage error; one returned after it is a runtime failure.
var preRunReached bool

var rootCmd = &cobra.Command{
	Use:     "sy",
	Short:   "Session manager for multi-repo development",
	Version: version,
	// main prints the terminating error once, as "fatal: ...", and picks the
	// exit status from it. cobra's own "Error:" line and usage dump would
	// duplicate that.
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		preRunReached = true
		if cmd == configEditCmd || cmd == configInitCmd {
			return nil
		}
		if _, err := config.Load(); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if greedyQuery != "" {
			match := greedyMatch(greedyQuery, sessions)
			if match == nil {
				return exitcode.NotFoundf("no session matches '%s'", greedyQuery)
			}
			fmt.Println(match.Path)
			return nil
		}

		// Default: show list (same as `sy list`)
		return printSessionList(sessions, "", noSessionsMessage())
	},
}

// noSessionsMessage is shown when the sessions table is empty.
func noSessionsMessage() string {
	return "No sessions yet. Create one with " + ui.AccentBold("sy new <name>")
}

// printSessionList renders sessions in the requested format. empty is the
// message shown when the list has no entries and the format is the human table.
func printSessionList(sessions []session.Session, format, empty string) error {
	switch format {
	case "json":
		return printSessionsJSON(sessions)
	case "names":
		for _, s := range sessions {
			fmt.Println(s.Name)
		}
		return nil
	case "paths":
		for _, s := range sessions {
			fmt.Println(s.Path)
		}
		return nil
	}

	// Default: human-readable table. The prose for an empty list is advice,
	// not data, so it goes to stderr and a piped listing stays empty.
	if len(sessions) == 0 {
		fmt.Fprintln(os.Stderr, ui.Info(empty))
		return nil
	}

	rows := make([][]string, len(sessions))
	for i, s := range sessions {
		rows[i] = []string{s.Name, fmt.Sprintf("%d", s.RepoCount), ui.StdoutFaint(formatRelativeTime(s.LastModified))}
	}
	return table(os.Stdout, []string{"SESSION", "REPOS", "MODIFIED"}, rows)
}

// greedyMatch returns the best session matching query: exact > prefix > substring (case-insensitive).
func greedyMatch(query string, sessions []session.Session) *session.Session {
	q := strings.ToLower(query)
	for i, s := range sessions {
		if strings.ToLower(s.Name) == q {
			return &sessions[i]
		}
	}
	for i, s := range sessions {
		if strings.HasPrefix(strings.ToLower(s.Name), q) {
			return &sessions[i]
		}
	}
	for i, s := range sessions {
		if strings.Contains(strings.ToLower(s.Name), q) {
			return &sessions[i]
		}
	}
	return nil
}

// Execute runs the CLI. An error cobra raised before the command's pre-run
// hook, such as an unknown flag or the wrong number of arguments, is marked as
// a usage error so it exits 2.
func Execute() error {
	preRunReached = false
	err := rootCmd.Execute()
	if err != nil && !preRunReached && !errors.Is(err, exitcode.ErrUsage) {
		return exitcode.Mark(exitcode.ErrUsage, err)
	}
	return err
}

// rootUsageSections documents what the man page of a git-shaped tool would:
// the exit statuses scripts branch on and the environment sy reads. Only the
// root help carries them.
const rootUsageSections = `{{if not .HasParent}}
Exit Status:
  0    success
  1    failure
  2    usage error
  3    session, repo, or archive entry not found
  4    refused: a confirmation sy could not ask for, or was answered no
  128  git failed; git's own message is printed after "fatal:"

Environment:
  SESHY_CONFIG            config file, instead of $XDG_CONFIG_HOME/seshy/config.yaml
  SESHY_<FIELD>           override one config field, e.g. SESHY_SESSIONS_DIR,
                          SESHY_BRANCH_FORMAT, SESHY_HOOKS_POST_CREATE='["cmd"]'
  XDG_CONFIG_HOME         config root (default ~/.config)
  XDG_STATE_HOME          sessions and archive root (default ~/.local/state)
  SYSINIT_PATHS_MANIFEST  paths manifest that names the sessions directory
  NO_COLOR                disable color on every stream
  EDITOR                  editor for "sy config edit" (default vi)

Hooks run with the caller's environment minus GIT_DIR, GIT_WORK_TREE, and
GIT_INDEX_FILE, plus SESHY_SESSION, SESHY_SESSION_PATH, SESHY_REPOS,
SESHY_REPO_COUNT, and SESHY_EVENT.
{{end}}`

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("sy version %s\n", version))
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + rootUsageSections)
	rootCmd.Flags().StringVar(&greedyQuery, "greedy", "", "Fuzzy-match a session name and print its path")
}
