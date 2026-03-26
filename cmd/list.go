package cmd

import (
	"fmt"
	"time"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all sessions",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		return printSessionList(sessions)
	},
}

func formatRelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
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
