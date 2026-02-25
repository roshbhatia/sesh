package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/fzf"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new session",
	Long:  `Create a new session with the given name and select repositories from zoxide.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Check dependencies first
		if err := fzf.CheckDependencies(); err != nil {
			return err
		}

		// Validate session name
		if err := session.ValidateSessionName(name); err != nil {
			return err
		}

		// Check if session already exists
		if session.Exists(name) {
			return fmt.Errorf("session '%s' already exists", name)
		}

		// Select repos using fzf
		fmt.Println("Select repositories (Space to select, Enter to confirm)...")
		repos, err := fzf.SelectRepos()
		if err != nil {
			return fmt.Errorf("failed to select repositories: %w", err)
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repositories selected")
		}

		// Create the session
		if err := session.Create(name, repos); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Get session path to display
		sessionPath, _ := session.GetPath(name)

		fmt.Printf("\n✓ Created session '%s'\n", name)
		fmt.Printf("  Path: %s\n", sessionPath)
		fmt.Printf("  Repos: %d\n", len(repos))
		fmt.Printf("\nNavigate with: cd %s\n", sessionPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
