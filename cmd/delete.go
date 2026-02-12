package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete a session",
	Long:    `Delete a session and clean up all associated git worktrees.`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		sessions, err := session.List()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, s := range sessions {
			completions = append(completions, s.Name)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Check if session exists
		if !session.Exists(name) {
			return fmt.Errorf("session '%s' not found", name)
		}

		// Delete the session
		if err := session.Delete(name); err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}

		fmt.Printf("✓ Deleted session '%s'\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
