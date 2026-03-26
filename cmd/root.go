package cmd

import (
	"fmt"
	"strings"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/roshbhatia/seshy/internal/ui"
	"github.com/spf13/cobra"
)

const version = "3.0.0"

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
		return printSessionList(sessions)
	},
}

func printSessionList(sessions []session.Session) error {
	if len(sessions) == 0 {
		fmt.Println(ui.Info("No sessions yet. Create one with " + ui.AccentBold.Render("sy new <name>")))
		return nil
	}

	headers := []string{"SESSION", "REPOS", "MODIFIED"}
	rows := make([][]string, len(sessions))
	for i, s := range sessions {
		rows[i] = []string{
			s.Name,
			fmt.Sprintf("%d", s.RepoCount),
			formatRelativeTime(s.LastModified),
		}
	}
	fmt.Println(ui.NewTable(headers, rows))
	return nil
}

// greedyMatch returns the best session matching query: exact → prefix → substring (case-insensitive).
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
