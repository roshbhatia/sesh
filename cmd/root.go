package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/fzf"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

const version = "3.0.0"

var rootCmd = &cobra.Command{
	Use:   "sesh",
	Short: "Multi-repo session manager with git worktree support",
	Long: `sesh - A streamlined session manager for developers.

Create isolated workspaces with git worktrees and shell.nix environments.
Mirrors zoxide's elegant UX with 's' and 'si' commands.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no arguments provided, run interactive selector
		if len(args) == 0 {
			sessions, err := session.List()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found. Create one with: sesh new <name>")
				return nil
			}

			// Extract session names
			names := make([]string, len(sessions))
			for i, s := range sessions {
				names[i] = s.Name
			}

			// Show interactive selector
			selected, err := fzf.SelectSession(names)
			if err != nil {
				return fmt.Errorf("selection failed: %w", err)
			}

			if selected == "" {
				return nil
			}

			// Print path to selected session
			path, err := session.GetPath(selected)
			if err != nil {
				return fmt.Errorf("failed to get session path: %w", err)
			}

			fmt.Println(path)
			return nil
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("sesh version %s\n", version))
}
