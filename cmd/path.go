package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/fzf"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Print the path to a session",
	Long:  `Print the absolute path to a session. Used by shell integration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		path, err := session.GetPath(name)
		if err != nil {
			return err
		}

		fmt.Println(path)
		return nil
	},
}

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively select a session",
	Long:  `Launch fzf to interactively select a session. Used by 'si' shell function.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check dependencies
		if err := fzf.CheckDependencies(); err != nil {
			return err
		}

		// Get all sessions
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			return fmt.Errorf("no sessions found. Create one with: sesh new <name>")
		}

		// Extract session names
		names := make([]string, len(sessions))
		for i, s := range sessions {
			names[i] = s.Name
		}

		// Select session with fzf
		selected, err := fzf.SelectSession(names)
		if err != nil {
			return err
		}

		// Print path to selected session
		path, err := session.GetPath(selected)
		if err != nil {
			return err
		}

		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(selectCmd)
}
