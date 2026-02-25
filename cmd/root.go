package cmd

import (
	"fmt"
	"strings"

	"github.com/roshbhatia/sesh/internal/picker"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

const version = "3.0.0"

var greedyQuery string

var rootCmd = &cobra.Command{
	Use:     "sesh",
	Short:   "Multi-repo session manager with git worktree support",
	Long:    `sesh - A streamlined session manager for developers.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found. Create one with: sesh new <name>")
			return nil
		}

		if greedyQuery != "" {
			match := greedyMatch(greedyQuery, sessions)
			if match == nil {
				return fmt.Errorf("no session matches %q", greedyQuery)
			}
			fmt.Println(match.Path)
			return nil
		}

		names := make([]string, len(sessions))
		for i, s := range sessions {
			names[i] = s.Name
		}

		selected, err := picker.SelectOne("Select session", names)
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}

		path, err := session.GetPath(selected)
		if err != nil {
			return fmt.Errorf("failed to get session path: %w", err)
		}

		fmt.Println(path)
		return nil
	},
}

// greedyMatch returns the best session matching query: exact match first,
// then prefix match, then substring match (case-insensitive).
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
	rootCmd.SetVersionTemplate(fmt.Sprintf("sesh version %s\n", version))
	rootCmd.Flags().StringVar(&greedyQuery, "greedy", "", "fuzzy match a session by name and print its path")
}
