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

		fmt.Printf("%-10s %s\n", ui.Faint("session"), ui.AccentBold(name))
		fmt.Printf("%-10s %s\n", ui.Faint("path"), contractHome(sessionPath))
		fmt.Printf("%-10s %d\n", ui.Faint("repos"), len(repos))

		if len(repos) == 0 {
			return nil
		}

		fmt.Println()

		nameW := len("NAME")
		branchW := len("BRANCH")
		for _, r := range repos {
			if len(r.Name) > nameW {
				nameW = len(r.Name)
			}
			label := branchLabel(r)
			if len(label) > branchW {
				branchW = len(label)
			}
		}

		fmtStr := fmt.Sprintf("  %%-%ds  %%-%ds  %%s\n", nameW, branchW)
		fmt.Printf(fmtStr,
			ui.ANSIColor(ui.ColorPurple, "NAME"),
			ui.ANSIColor(ui.ColorPurple, "BRANCH"),
			ui.ANSIColor(ui.ColorPurple, "SOURCE"),
		)
		for _, r := range repos {
			branchStr := ui.ANSIColor(ui.ColorPurple, r.Branch)
			if r.Branch == "" || r.Detached {
				branchStr = ui.Faint(branchLabel(r))
			}
			fmt.Printf(fmtStr, r.Name, branchStr, ui.Faint(contractHome(r.SourcePath)))
		}

		return nil
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
