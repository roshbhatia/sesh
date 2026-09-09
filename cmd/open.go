package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

// openVersion names the document "sy open --format json" prints.
const openVersion = "seshy.open/v1"

// openFormats are the formats "sy open" renders. The table form is the bare
// path, the same line "sy path" prints.
var openFormats = []string{formatTable, formatJSON}

var openFormat string

// openJSON is the seshy.open/v1 document: the plan a launcher runs to enter a
// session. It carries what seshy knows and nothing a caller layers on, so the
// environment holds SESHY_SESSION and no launcher-specific names.
type openJSON struct {
	Version      string            `json:"version" jsonschema:"enum=seshy.open/v1,description=Document version"`
	ID           string            `json:"id" jsonschema:"description=Row identifier, seshy:<name>"`
	Cwd          string            `json:"cwd" jsonschema:"description=Absolute session directory"`
	Command      []string          `json:"command" jsonschema:"description=Program to run; empty means the caller's default"`
	Environment  map[string]string `json:"environment" jsonschema:"description=Variables to set; SESHY_SESSION names the session"`
	SuccessCodes []int             `json:"successCodes" jsonschema:"description=Exit statuses the caller treats as success"`
}

// openDocument builds the seshy.open/v1 document for a resolved session.
func openDocument(name, path string) openJSON {
	return openJSON{
		Version:      openVersion,
		ID:           sessionID(name),
		Cwd:          path,
		Command:      []string{},
		Environment:  map[string]string{"SESHY_SESSION": name},
		SuccessCodes: []int{0},
	}
}

// sessionNameOf accepts a session name or its seshy:<name> id, so the id a
// listing printed can be handed straight back.
func sessionNameOf(arg string) string {
	return strings.TrimPrefix(arg, "seshy:")
}

var openCmd = &cobra.Command{
	Use:   "open <name>",
	Short: "Print what a launcher needs to enter a session",
	Long: `Print the session directory, or with --format json the seshy.open/v1 plan a
launcher runs to enter the session: its cwd, an empty command for the caller's
default program, and SESHY_SESSION in the environment.

The name is matched exactly, or given as the seshy:<name> id a listing
printed. A session that no longer exists exits 3.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSessionNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := resolveFormat(openFormat, nil, openFormats...)
		if err != nil {
			return err
		}
		name := sessionNameOf(args[0])
		path, err := session.Resolve(name)
		if err != nil {
			return err
		}
		if format == formatJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(openDocument(name, path))
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	openCmd.Flags().StringVar(&openFormat, "format", "", formatUsage(openFormats...))
	rootCmd.AddCommand(openCmd)
}
