package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/hook"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var (
	forceDelete    bool
	yesDelete      bool
	deleteArchived bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Short:   "Delete a session",
	Aliases: []string{"rm"},
	Args:    cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if deleteArchived {
			return completeArchivedNames(cmd, args, toComplete)
		}
		return completeSessionNames(cmd, args, toComplete)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		list, resolve := session.List, session.Resolve
		if deleteArchived {
			list, resolve = session.ListArchived, session.ResolveArchived
		}

		sessions, err := list()
		if err != nil {
			return fmt.Errorf("listing sessions: %w", err)
		}
		if len(args) == 0 && len(sessions) == 0 {
			fmt.Fprintln(os.Stderr, ui.Info("No sessions to delete."))
			return nil
		}

		name, err := selectSessionName(args, cfg.SessionPicker, sessionNames(sessions))
		if err != nil {
			return err
		}
		if name == "" {
			return nil
		}

		sessionPath, err := resolve(name)
		if err != nil {
			return err
		}

		ok, err := confirm(fmt.Sprintf("Delete session %s and its worktrees? Branches will remain.", ui.AccentBold(name)), skipConfirm(yesDelete, forceDelete))
		if err != nil || !ok {
			return err
		}

		if deleteArchived {
			if err := session.DeleteArchived(name, forceDelete); err != nil {
				return reportDeleteResult(name, err)
			}
		} else {
			// Run pre-delete hooks with full repo info
			repoInfos := session.GetSessionRepoInfos(sessionPath)
			data := session.BuildTemplateData(name, sessionPath, repoInfos)
			hook.Run("pre-delete", cfg.Hooks.PreDelete, data, sessionPath)

			if err := session.Delete(name, forceDelete); err != nil {
				return reportDeleteResult(name, err)
			}
		}

		fmt.Fprintln(os.Stderr, ui.Successf("Deleted session %s", ui.AccentBold(name)))
		return nil
	},
}

// reportDeleteResult turns a delete error into either a hard failure or, when
// --force already removed the session, a warning plus success.
func reportDeleteResult(name string, err error) error {
	if !errors.Is(err, session.ErrCleanupIncomplete) {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	fmt.Fprintln(os.Stderr, ui.Warningf("Deleted session %s, but some worktrees or branches were left behind:", ui.AccentBold(name)))
	fmt.Fprintf(os.Stderr, "  %v\n", err)
	return nil
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation and delete even if worktree cleanup fails")
	deleteCmd.Flags().BoolVarP(&yesDelete, "yes", "y", false, "Skip the confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteArchived, "archived", false, "Delete an archived session instead of an active one")
	rootCmd.AddCommand(deleteCmd)
}
