package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:               "status [name]",
	Short:             "Show session details",
	Aliases:           []string{"info"},
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeSessionNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string

		if len(args) > 0 {
			name = args[0]
		} else {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			sessions, err := session.List()
			if err != nil {
				return fmt.Errorf("listing sessions: %w", err)
			}
			if len(sessions) == 0 {
				fmt.Fprintln(os.Stderr, ui.Info("No sessions yet. Create one with "+ui.AccentBold("sy new <name>")))
				return nil
			}
			names := make([]string, len(sessions))
			for i, s := range sessions {
				names[i] = s.Name
			}
			selected, err := runPicker(cfg.SessionPicker, names)
			if err != nil {
				if errors.Is(err, errCancelled) {
					return nil
				}
				return err
			}
			if len(selected) == 0 {
				return fmt.Errorf("no session selected")
			}
			name = selected[0]
		}

		sessionPath, err := session.Resolve(name)
		if err != nil {
			return err
		}

		repos := session.GetSessionRepoInfos(sessionPath)

		summary := [][]string{
			{ui.StdoutFaint("session"), ui.StdoutColor(ui.ColorPurple, name)},
			{ui.StdoutFaint("path"), contractHome(sessionPath)},
			{ui.StdoutFaint("repos"), fmt.Sprintf("%d", len(repos))},
		}
		if err := table(os.Stdout, nil, summary); err != nil {
			return err
		}

		if len(repos) == 0 {
			return nil
		}

		fmt.Println()

		rows := make([][]string, len(repos))
		for i, r := range repos {
			branch := ui.StdoutColor(ui.ColorPurple, r.Branch)
			if r.Branch == "" || r.Detached {
				branch = ui.StdoutFaint(branchLabel(r))
			}
			rows[i] = []string{"  " + r.Name, branch, ui.StdoutFaint(contractHome(r.SourcePath))}
		}
		return table(os.Stdout, []string{"  NAME", "BRANCH", "SOURCE"}, rows)
	},
}

// branchLabel names what a repo entry is checked out on, or what it is when
// it has no branch.
func branchLabel(r session.RepoInfo) string {
	switch {
	case r.Branch == "":
		return "(symlink)"
	case r.Detached:
		return "(detached)"
	}
	return r.Branch
}

func contractHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
