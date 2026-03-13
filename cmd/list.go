package cmd

import (
	"fmt"
	"time"

	"github.com/roshbhatia/sesh/internal/session"
	"github.com/roshbhatia/sesh/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all sessions",
	Long:    `List all sessions with their repository counts and last modified times.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println(ui.Info("No sessions found. Create one with: sesh new <name>"))
			return nil
		}

		headers := []string{"NAME", "REPOS", "LAST MODIFIED"}
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
	},
}

func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
