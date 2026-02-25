package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/fzf"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add repositories to an existing session",
	Long:  `Add additional repositories to an existing session.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Check if session exists
		if !session.Exists(name) {
			return fmt.Errorf("session '%s' not found", name)
		}

		// Check dependencies
		if err := fzf.CheckDependencies(); err != nil {
			return err
		}

		// Select repos using fzf
		fmt.Println("Select repositories to add (Space to select, Enter to confirm)...")
		repos, err := fzf.SelectRepos()
		if err != nil {
			return fmt.Errorf("failed to select repositories: %w", err)
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repositories selected")
		}

		// Add repos to session
		if err := session.AddRepos(name, repos); err != nil {
			return fmt.Errorf("failed to add repositories: %w", err)
		}

		// Get session path to display
		sessionPath, _ := session.GetPath(name)

		fmt.Printf("\n✓ Added %d repo(s) to session '%s'\n", len(repos), name)
		fmt.Printf("  Path: %s\n", sessionPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
