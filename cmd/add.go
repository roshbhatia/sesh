package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/picker"
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

		if !session.Exists(name) {
			return fmt.Errorf("session '%s' not found", name)
		}

		dirs, err := zoxideDirs()
		if err != nil {
			return err
		}

		repos, err := picker.SelectMany("Select repositories to add", dirs)
		if err != nil {
			return fmt.Errorf("failed to select repositories: %w", err)
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repositories selected")
		}

		if err := session.AddRepos(name, repos); err != nil {
			return fmt.Errorf("failed to add repositories: %w", err)
		}

		sessionPath, _ := session.GetPath(name)
		fmt.Printf("Added %d repo(s) to session '%s'\n", len(repos), name)
		fmt.Printf("  Path: %s\n", sessionPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
