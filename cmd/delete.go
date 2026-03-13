package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/roshbhatia/sesh/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var forceDelete bool

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

		if !forceDelete {
			// Check if stdin is a TTY
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("non-interactive environment: use --force to delete without confirmation")
			}

			var confirmed bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Delete session '%s'?", name)).
				Affirmative("Yes, delete").
				Negative("Cancel").
				Value(&confirmed).
				Run()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				fmt.Fprintln(os.Stderr, ui.Info("Deletion cancelled."))
				return nil
			}
		}

		// Delete the session
		if err := session.Delete(name); err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}

		fmt.Fprintln(os.Stderr, ui.Successf("Deleted session '%s'", name))
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}
