package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var (
	listFormat   string
	listJSON     bool
	listNames    bool
	listPaths    bool
	listArchived bool
)

// listFormats are the formats "sy list" renders.
var listFormats = []string{formatTable, formatJSON, formatNames, formatPaths}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all sessions",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := resolveFormat(listFormat, map[string]bool{
			formatJSON:  listJSON,
			formatNames: listNames,
			formatPaths: listPaths,
		}, listFormats...)
		if err != nil {
			return err
		}

		list, empty := session.List, noSessionsMessage()
		if listArchived {
			list, empty = session.ListArchived, "No archived sessions. Archive one with "+ui.AccentBold("sy archive <name>")
		}

		sessions, err := list()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		return printSessionList(sessions, format, empty, listArchived)
	},
}

// sessionID is the identifier other tools address a session by. The prefix
// keeps it distinct from ids other session sources hand the same display.
func sessionID(name string) string { return "seshy:" + name }

// sessionJSON is one entry of "sy list --format json". The first four keys
// predate id and archived and keep their names and shapes.
type sessionJSON struct {
	Name         string `json:"name" jsonschema:"description=Session name, the directory basename"`
	Path         string `json:"path" jsonschema:"description=Absolute session directory"`
	RepoCount    int    `json:"repoCount" jsonschema:"description=Repo entries in the session directory"`
	LastModified string `json:"lastModified" jsonschema:"format=date-time,description=Directory modification time (RFC 3339)"`
	ID           string `json:"id" jsonschema:"description=Row identifier, seshy:<name>"`
	Archived     bool   `json:"archived" jsonschema:"description=True when listed from the archive"`
}

// sessionListJSON is the document "sy list --format json" prints: a bare
// array, one element per session.
type sessionListJSON []sessionJSON

func printSessionsJSON(sessions []session.Session, archived bool) error {
	out := make(sessionListJSON, len(sessions))
	for i, s := range sessions {
		out[i] = sessionJSON{
			Name:         s.Name,
			Path:         s.Path,
			RepoCount:    s.RepoCount,
			LastModified: s.LastModified.Format(time.RFC3339),
			ID:           sessionID(s.Name),
			Archived:     archived,
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
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
	listCmd.Flags().StringVar(&listFormat, "format", "", formatUsage(listFormats...))
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output JSON")
	listCmd.Flags().BoolVar(&listNames, "names", false, "Output session names only")
	listCmd.Flags().BoolVar(&listPaths, "paths", false, "Output session paths only")
	listCmd.Flags().BoolVar(&listArchived, "archived", false, "List archived sessions instead of active ones")
	rootCmd.AddCommand(listCmd)
}
