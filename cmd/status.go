package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

// statusVersion names the document "sy status --format json" prints.
const statusVersion = "seshy.status/v1"

// statusFormats are the formats "sy status" renders.
var statusFormats = []string{formatTable, formatJSON}

var statusFormat string

// statusJSON is the seshy.status/v1 document: one session and every repo
// entry in it.
type statusJSON struct {
	Version string           `json:"version" jsonschema:"enum=seshy.status/v1,description=Document version"`
	Name    string           `json:"name" jsonschema:"description=Session name"`
	Path    string           `json:"path" jsonschema:"description=Absolute session directory"`
	Repos   []statusRepoJSON `json:"repos" jsonschema:"description=Repo entries in directory order"`
}

// statusRepoJSON is one repo entry of a session.
type statusRepoJSON struct {
	Name         string `json:"name" jsonschema:"description=Entry basename in the session directory"`
	Path         string `json:"path" jsonschema:"description=Absolute entry path"`
	Source       string `json:"source" jsonschema:"description=Absolute path of the source repo or linked directory"`
	Branch       string `json:"branch" jsonschema:"description=Checked-out branch; empty for a symlink; HEAD when detached"`
	Kind         string `json:"kind" jsonschema:"enum=worktree,enum=symlink,enum=clone,description=What the entry is"`
	Locked       bool   `json:"locked" jsonschema:"description=The worktree is locked in its source repo"`
	Detached     bool   `json:"detached" jsonschema:"description=The worktree is on no branch"`
	BranchReused bool   `json:"branchReused" jsonschema:"description=The branch existed before seshy checked it out"`
}

// statusDocument builds the seshy.status/v1 document for a session.
func statusDocument(name, path string, repos []session.RepoInfo) statusJSON {
	out := statusJSON{Version: statusVersion, Name: name, Path: path, Repos: make([]statusRepoJSON, len(repos))}
	for i, r := range repos {
		out.Repos[i] = statusRepoJSON{
			Name:         r.Name,
			Path:         r.Path,
			Source:       r.SourcePath,
			Branch:       r.Branch,
			Kind:         r.Kind,
			Locked:       r.Locked,
			Detached:     r.Detached,
			BranchReused: r.Reused,
		}
	}
	return out
}

var statusCmd = &cobra.Command{
	Use:               "status [name]",
	Short:             "Show session details",
	Aliases:           []string{"info"},
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeSessionNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := resolveFormat(statusFormat, nil, statusFormats...)
		if err != nil {
			return err
		}

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

		if format == formatJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(statusDocument(name, sessionPath, repos))
		}

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
	statusCmd.Flags().StringVar(&statusFormat, "format", "", formatUsage(statusFormats...))
	rootCmd.AddCommand(statusCmd)
}
