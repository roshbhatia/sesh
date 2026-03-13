package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/roshbhatia/sesh/internal/picker"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/roshbhatia/sesh/internal/ui"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new session",
	Long:  `Create a new session with the given name and select repositories from zoxide.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := session.ValidateSessionName(name); err != nil {
			return err
		}

		if session.Exists(name) {
			return fmt.Errorf("session '%s' already exists", name)
		}

		dirs, err := zoxideDirs()
		if err != nil {
			return err
		}

		repos, err := picker.SelectMany("Select repositories", dirs)
		if err != nil {
			return fmt.Errorf("failed to select repositories: %w", err)
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repositories selected")
		}

		err = ui.RunWithSpinner("Creating worktrees", func() error {
			return session.Create(name, repos)
		})
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		sessionPath, _ := session.GetPath(name)
		fmt.Fprintln(os.Stderr, ui.Successf("Created session '%s'", name))
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.Dim("Path:"), sessionPath)
		fmt.Fprintf(os.Stderr, "  %s %d\n", ui.Dim("Repos:"), len(repos))

		return nil
	},
}

func zoxideDirs() ([]string, error) {
	out, err := exec.Command("zoxide", "query", "--list").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query zoxide: %w", err)
	}
	dirs := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(dirs) == 0 || (len(dirs) == 1 && dirs[0] == "") {
		return nil, fmt.Errorf("no directories found in zoxide database")
	}
	return dirs, nil
}

func init() {
	rootCmd.AddCommand(newCmd)
}
