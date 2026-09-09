package cmd

import (
	"fmt"
	"os"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var pruneDryRun bool

var pruneCmd = &cobra.Command{
	Use:   "prune [repo...]",
	Short: "Reclaim worktrees and branches a removed session left behind",
	Long: `Reclaim what a removed session left in git.

For each repo, prune worktree registrations whose directories are gone, then
delete the seshy branches whose session directory no longer exists. Also
remove symlinks under the sessions root whose targets are gone. With no repo
arguments, prune visits every source repo of every session. A "-" argument
reads repo paths from stdin.

Each action prints one line to stderr. --dry-run prints the actions without
taking them.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repos, _ := readRepoArgs(args, false, os.Stdin)
		actions, err := session.Prune(repos, pruneDryRun)
		for _, action := range actions {
			if pruneDryRun {
				fmt.Fprintln(os.Stderr, ui.Info("would "+action.String()))
			} else {
				fmt.Fprintln(os.Stderr, action.String())
			}
		}
		if err != nil {
			return fmt.Errorf("prune incomplete: %w", err)
		}
		if len(actions) == 0 {
			fmt.Fprintln(os.Stderr, ui.Info("Nothing to prune."))
		}
		return nil
	},
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "Print the actions without taking them")
	rootCmd.AddCommand(pruneCmd)
}
