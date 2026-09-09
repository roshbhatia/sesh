package cmd

import (
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

var rootCmd = &cobra.Command{
	Use:     "sy",
	Short:   "Session manager for multi-repo development",
	Version: version,
	// Runs after argument and flag validation, so usage still prints for a
	// malformed invocation but not for a runtime failure like "session not found".
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
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
				return fmt.Errorf("no session matches '%s': %w", greedyQuery, exitcode.ErrNotFound)
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

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("sy version %s\n", version))
	rootCmd.Flags().StringVar(&greedyQuery, "greedy", "", "Fuzzy-match a session name and print its path")
}
