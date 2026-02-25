package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/picker"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

const version = "3.0.0"

var rootCmd = &cobra.Command{
	Use:   "sesh",
	Short: "Multi-repo session manager with git worktree support",
	Long:  `sesh - A streamlined session manager for developers.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			sessions, err := session.List()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found. Create one with: sesh new <name>")
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
