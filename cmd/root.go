package cmd

import (
	"fmt"

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
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("sesh version %s\n", version))
}
