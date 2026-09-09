package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/hook"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/roshbhatia/seshy/internal/tmpl"
	"github.com/spf13/cobra"
)

var (
	newBranch    string
	newStart     string
	newExisting  bool
	newReference bool
	newStdin     bool
	newEmpty     bool
)

var newCmd = &cobra.Command{
	Use:   "new <name> [repos...]",
	Short: "Create a new session",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := session.ValidateSessionName(name); err != nil {
			return err
		}
		if session.Exists(name) {
			return fmt.Errorf("session '%s' already exists", name)
		}

		if newExisting && newBranch == "" {
			return fmt.Errorf("--existing requires --branch")
		}
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		repos, fromStdin := readRepoArgs(args[1:], newStdin, os.Stdin)

		if newEmpty && len(repos) > 0 {
			return exitcode.Usagef("--empty takes no repositories")
		}

		// The picker is the only interactive path here. --empty and --stdin both
		// say the caller already supplied the repo list, so an empty list means an
		// empty session instead of a prompt no one is there to answer.
		if len(repos) == 0 && !newEmpty && !fromStdin {
			candidates, err := runSource(cfg.RepoSource)
			if err != nil {
				return fmt.Errorf("repo source: %w", err)
			}
			candidates = prependDefaults(cfg.DefaultRepos, candidates)
			if len(candidates) > 0 {
				selected, err := runPicker(cfg.Picker, candidates)
				if err != nil {
					if errors.Is(err, errCancelled) {
						return nil
					}
					return err
				}
				repos = selected
			}
		}

		opts := session.CreateOpts{
			BranchFormat:    cfg.BranchFormat,
			BranchFormatFor: branchFormatResolver(),
			BranchOverride:  newBranch,
			StartPoint:      newStart,
			ExistingBranch:  newExisting,
			Reference:       newReference,
		}

		repoInfos, err := session.Create(name, repos, opts)
		if err != nil {
			return err
		}

		warnReusedBranches(repoInfos)

		sessionPath, _ := session.Resolve(name)
		data := session.BuildTemplateData(name, sessionPath, repoInfos)

		// Render per-repo templates
		repoTmplDir := filepath.Join(config.ConfigDir(), "templates", "repo")
		for _, ri := range repoInfos {
			rd := data.ForRepo(tmpl.RepoData{Name: ri.Name, Path: ri.Path, Source: ri.SourcePath, Branch: ri.Branch})
			if err := tmpl.RenderDir(repoTmplDir, ri.Path, rd); err != nil {
				fmt.Fprintln(os.Stderr, ui.Warningf("template error for %s: %v", ri.Name, err))
			}
		}

		// Render session-level templates
		sessionTmplDir := filepath.Join(config.ConfigDir(), "templates", "session")
		if err := tmpl.RenderDir(sessionTmplDir, sessionPath, data); err != nil {
			fmt.Fprintln(os.Stderr, ui.Warningf("session template error: %v", err))
		}

		// Run post-create hooks
		hook.Run("post-create", cfg.Hooks.PostCreate, data, sessionPath)

		fmt.Fprintln(os.Stderr, ui.Successf("Created session %s", ui.AccentBold(name)))
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.Faint("path:"), sessionPath)
		fmt.Fprintf(os.Stderr, "  %s %d\n", ui.Faint("repos:"), len(repoInfos))
		if len(repoInfos) == 0 {
			fmt.Fprintln(os.Stderr, ui.Info("Empty session. Add repos with "+ui.AccentBold("sy add "+name)))
		}
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newStart, "start-point", "", "Commit to start a new branch from (default HEAD)")
	newCmd.Flags().BoolVar(&newExisting, "existing", false, "Check out the existing branch named by --branch")
	newCmd.Flags().BoolVar(&newReference, "reference", false, "Link existing directories without creating worktrees")
	newCmd.Flags().StringVarP(&newBranch, "branch", "b", "", "Override branch name for all worktrees")
	newCmd.Flags().BoolVar(&newStdin, "stdin", false, "Read repo paths from stdin")
	newCmd.Flags().BoolVar(&newEmpty, "empty", false, "Create the session with no repositories")
	newCmd.MarkFlagsMutuallyExclusive("reference", "branch")
	newCmd.MarkFlagsMutuallyExclusive("reference", "start-point")
	newCmd.MarkFlagsMutuallyExclusive("existing", "start-point")
	rootCmd.AddCommand(newCmd)
}
