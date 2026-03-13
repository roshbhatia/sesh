package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/roshbhatia/sesh/internal/picker"
	"github.com/roshbhatia/sesh/internal/session"
	"github.com/roshbhatia/sesh/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add repositories to an existing session",
	Long:  `Add additional repositories to an existing session.`,
	Args:  cobra.ExactArgs(1),
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

		if !session.Exists(name) {
			return fmt.Errorf("session '%s' not found", name)
		}

		sessionPath, err := session.GetPath(name)
		if err != nil {
			return err
		}

		dirs, err := zoxideDirs()
		if err != nil {
			return err
		}

		// Filter out repos already in the session
		existingSources, _ := session.ListRepoSources(sessionPath)
		existingSet := make(map[string]bool, len(existingSources))
		for _, s := range existingSources {
			resolved, err := filepath.EvalSymlinks(s)
			if err != nil {
				resolved = s
			}
			existingSet[resolved] = true
		}

		var available []string
		for _, d := range dirs {
			resolved, err := filepath.EvalSymlinks(d)
			if err != nil {
				resolved = d
			}
			if !existingSet[resolved] {
				available = append(available, d)
			}
		}

		if len(available) == 0 {
			fmt.Fprintln(os.Stderr, ui.Info("All available repositories are already in the session."))
			return nil
		}

		repos, err := picker.SelectMany("Select repositories to add", available)
		if err != nil {
			return fmt.Errorf("failed to select repositories: %w", err)
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repositories selected")
		}

		var result session.AddResult
		err = ui.RunWithSpinner("Adding repositories", func() error {
			var e error
			result, e = session.AddReposResult(name, repos)
			return e
		})
		if err != nil {
			return fmt.Errorf("failed to add repositories: %w", err)
		}

		total := len(repos)
		added := len(result.Added)

		if len(result.Skipped) > 0 {
			for _, s := range result.Skipped {
				fmt.Fprintln(os.Stderr, ui.Warningf("Skipped %s (already in session)", s))
			}
		}

		if len(result.Errors) > 0 {
			for repo, e := range result.Errors {
				fmt.Fprintln(os.Stderr, ui.Errorf("Failed %s: %v", repo, e))
			}
		}

		fmt.Fprintln(os.Stderr, ui.Successf("Added %d/%d repo(s) to session '%s'", added, total, name))
		fmt.Fprintf(os.Stderr, "  %s %s\n", ui.Dim("Path:"), sessionPath)

		if result.Err() != nil {
			return fmt.Errorf("some repositories failed to add")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
