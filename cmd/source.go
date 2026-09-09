package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

// The roster contract seshy speaks as a session source. roster is the
// display-side aggregator; it asks each source for a catalog and writes one
// file per source for a launcher to read.
const (
	catalogVersion = "roster.catalog/v1"
	catalogSource  = "seshy"
	catalogTTL     = "10s"
	rowKind        = "session"
	hopLocal       = "local"
)

// catalogDisplay is how a launcher labels seshy's rows.
var catalogDisplay = catalogDisplayJSON{Label: "sessions", Glyph: "cod_briefcase", Order: 10}

// catalogJSON is one roster.catalog/v1 document. generated_at is stamped so
// the printed document is complete under roster's schema; roster restamps it
// when it writes the file.
type catalogJSON struct {
	Version     string             `json:"version"`
	Source      string             `json:"source"`
	GeneratedAt string             `json:"generated_at"`
	TTL         string             `json:"ttl"`
	Display     catalogDisplayJSON `json:"display"`
	Groups      []catalogGroupJSON `json:"groups"`
	Rows        []catalogRowJSON   `json:"rows"`
}

type catalogDisplayJSON struct {
	Label string `json:"label"`
	Glyph string `json:"glyph"`
	Order int    `json:"order"`
}

// catalogGroupJSON is a row group. seshy groups nothing, so the slice is
// always empty, but the shape is roster's.
type catalogGroupJSON struct {
	ID     string         `json:"id"`
	Label  string         `json:"label"`
	Glyph  string         `json:"glyph"`
	OK     bool           `json:"ok"`
	Stale  bool           `json:"stale"`
	Reason *string        `json:"reason"`
	Meta   map[string]any `json:"meta"`
}

// catalogRowJSON is one session as roster sees it. Nullable fields are
// pointers so they print as null rather than vanish.
type catalogRowJSON struct {
	ID        string           `json:"id"`
	Workspace string           `json:"workspace"`
	Label     string           `json:"label"`
	Group     *string          `json:"group"`
	Kind      string           `json:"kind"`
	Host      *string          `json:"host"`
	Cwd       string           `json:"cwd"`
	Status    *statusStateJSON `json:"status"`
	Pane      *string          `json:"pane"`
	Spawn     spawnJSON        `json:"spawn"`
	Meta      map[string]any   `json:"meta"`
}

// statusStateJSON is a row's live status. seshy records none, so every row
// prints null, but the shape is roster's.
type statusStateJSON struct {
	State  string  `json:"state"`
	Reason *string `json:"reason"`
	Since  string  `json:"since"`
}

// spawnJSON is the plan a launcher runs and the hop it runs it through.
type spawnJSON struct {
	Plan planJSON `json:"plan"`
	Hop  hopJSON  `json:"hop"`
}

type planJSON struct {
	Command      []string          `json:"command"`
	Cwd          string            `json:"cwd"`
	Environment  map[string]string `json:"environment"`
	SuccessCodes []int             `json:"successCodes"`
}

type hopJSON struct {
	Kind string `json:"kind"`
}

// catalogRow renders one session as a roster row. The spawn plan is the
// same one "sy open --format json" prints, so both routes agree.
func catalogRow(s session.Session) catalogRowJSON {
	plan := openDocument(s.Name, s.Path)
	return catalogRowJSON{
		ID:        plan.ID,
		Workspace: s.Name,
		Label:     s.Name,
		Kind:      rowKind,
		Cwd:       s.Path,
		Spawn: spawnJSON{
			Plan: planJSON{
				Command:      plan.Command,
				Cwd:          plan.Cwd,
				Environment:  plan.Environment,
				SuccessCodes: plan.SuccessCodes,
			},
			Hop: hopJSON{Kind: hopLocal},
		},
		Meta: map[string]any{
			"repoCount":    s.RepoCount,
			"lastModified": s.LastModified.Format(time.RFC3339),
		},
	}
}

// catalogDocument lists every active session as a roster.catalog/v1 document.
// Archived sessions are not rows: a launcher offers what can be entered.
func catalogDocument(now time.Time) (catalogJSON, error) {
	sessions, err := session.List()
	if err != nil {
		return catalogJSON{}, fmt.Errorf("failed to list sessions: %w", err)
	}
	rows := make([]catalogRowJSON, len(sessions))
	for i, s := range sessions {
		rows[i] = catalogRow(s)
	}
	return catalogJSON{
		Version:     catalogVersion,
		Source:      catalogSource,
		GeneratedAt: now.UTC().Format(time.RFC3339),
		TTL:         catalogTTL,
		Display:     catalogDisplay,
		Groups:      []catalogGroupJSON{},
		Rows:        rows,
	}, nil
}

// openRow resolves one row by its seshy:<name> id, for source.open.
func openRow(id string) (catalogRowJSON, error) {
	name := sessionNameOf(id)
	if name == id {
		return catalogRowJSON{}, exitcode.NotFoundf("row '%s' is not a seshy id", id)
	}
	path, err := session.Resolve(name)
	if err != nil {
		return catalogRowJSON{}, err
	}
	sessions, err := session.List()
	if err != nil {
		return catalogRowJSON{}, fmt.Errorf("failed to list sessions: %w", err)
	}
	for _, s := range sessions {
		if s.Name == name {
			return catalogRow(s), nil
		}
	}
	return catalogRow(session.Session{Name: name, Path: path}), nil
}

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Serve sessions to roster as a session source",
	Long: `Serve sessions to roster, the session-source aggregator a launcher reads.

"sy source list" prints the roster.catalog/v1 document directly, one row per
active session, for inspection. roster itself talks to "sy provider".`,
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print the roster.catalog/v1 document",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		document, err := catalogDocument(time.Now())
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(document)
	},
}

func init() {
	sourceCmd.AddCommand(sourceListCmd)
	rootCmd.AddCommand(sourceCmd)
}
