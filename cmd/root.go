package cmd

import (
	"fmt"
	"strings"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/roshbhatia/seshy/internal/ui"
	"github.com/spf13/cobra"
)

const version = "4.0.0"

var greedyQuery string

var rootCmd = &cobra.Command{
	Use:     "sy",
	Short:   "Session manager for multi-repo development",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if greedyQuery != "" {
			match := greedyMatch(greedyQuery, sessions)
			if match == nil {
				return fmt.Errorf("no session matches %q", greedyQuery)
			}
			fmt.Println(match.Path)
			return nil
		}

		// Default: show list (same as `sy list`)
		return printSessionList(sessions, "")
	},
}

func printSessionList(sessions []session.Session, format string) error {
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

	// Default: human-readable table
	if len(sessions) == 0 {
		fmt.Println(ui.Info("No sessions yet. Create one with " + ui.AccentBold("sy new <name>")))
		return nil
	}

	// Calculate column widths
	nameW, reposW := len("SESSION"), len("REPOS")
	rows := make([]struct{ name, repos, modified string }, len(sessions))
	for i, s := range sessions {
		rows[i].name = s.Name
		rows[i].repos = fmt.Sprintf("%d", s.RepoCount)
		rows[i].modified = formatRelativeTime(s.LastModified)
		if len(rows[i].name) > nameW {
			nameW = len(rows[i].name)
		}
		if len(rows[i].repos) > reposW {
			reposW = len(rows[i].repos)
		}
	}

	fmtStr := fmt.Sprintf("%%-%ds  %%-%ds  %%s\n", nameW, reposW)
	fmt.Printf(fmtStr, ui.Color(ui.ColorPurple, "SESSION"), ui.Color(ui.ColorPurple, "REPOS"), ui.Color(ui.ColorPurple, "MODIFIED"))
	for _, r := range rows {
		fmt.Printf(fmtStr, r.name, r.repos, ui.Faint(r.modified))
	}
	return nil
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
