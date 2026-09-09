package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/roshbhatia/seshy/internal/tmpl"
	"github.com/spf13/cobra"
)

var (
	forceRemove bool
	yesRemove   bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <session> <repo>",
	Short: "Remove a repository from a session",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			sessions, _ := session.List()
			names := make([]string, len(sessions))
			for i, s := range sessions {
				names[i] = s.Name
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		case 1:
			sessionPath, err := session.Resolve(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			repos := session.GetSessionRepoInfos(sessionPath)
			names := make([]string, len(repos))
			for i, r := range repos {
				names[i] = r.Name
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name, repoName := args[0], args[1]

		sessionPath, err := session.Resolve(name)
		if err != nil {
			return err
		}

		ok, err := confirm(fmt.Sprintf("Remove repo %s from session %s?", ui.AccentBold(repoName), ui.AccentBold(name)), skipConfirm(yesRemove, forceRemove))
		if err != nil || !ok {
			return err
		}

		if err := session.RemoveRepoEntry(sessionPath, repoName, forceRemove); err != nil {
			return fmt.Errorf("failed to remove repo: %w", err)
		}

		// Re-render session templates with updated repo list (non-fatal)
		allRepos := session.GetSessionRepoInfos(sessionPath)
		data := session.BuildTemplateData(name, sessionPath, allRepos)
		sessionTmplDir := filepath.Join(config.ConfigDir(), "templates", "session")
		_ = tmpl.RenderSessionDir(sessionTmplDir, sessionPath, data)

		fmt.Fprintln(os.Stderr, ui.Successf("Removed %s from session %s", ui.AccentBold(repoName), ui.AccentBold(name)))
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Skip confirmation prompt and remove even if worktree cleanup fails")
	removeCmd.Flags().BoolVarP(&yesRemove, "yes", "y", false, "Skip the confirmation prompt")
	rootCmd.AddCommand(removeCmd)
}
