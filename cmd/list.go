package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	Long:  `List all sessions with their repository counts and last modified times.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found. Create one with: sesh new <name>")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tREPOS\tLAST MODIFIED")
		fmt.Fprintln(w, "────\t─────\t─────────────")

		for _, s := range sessions {
			// Format time as relative
			timeStr := formatRelativeTime(s.LastModified)
			fmt.Fprintf(w, "%s\t%d\t%s\n", s.Name, s.RepoCount, timeStr)
		}

		w.Flush()
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
