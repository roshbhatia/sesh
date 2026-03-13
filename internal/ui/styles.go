package ui

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Styles holds all lipgloss styles for consistent CLI output.
type Styles struct {
	Title   lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Dim     lipgloss.Style
	Bold    lipgloss.Style
	Accent  lipgloss.Style
}

// NewStyles builds a style-set for the given writer. It honours:
//   - NO_COLOR / CLICOLOR / CLICOLOR_FORCE environment variables
//   - COLORTERM=truecolor / 24bit
//   - Terminal capability detection via TERM / TERM_PROGRAM
func NewStyles(w io.Writer) Styles {
	profile := termenv.EnvColorProfile()
	if termenv.EnvNoColor() {
		profile = termenv.Ascii
	}

	out := termenv.NewOutput(w, termenv.WithColorCache(true), termenv.WithProfile(profile))
	r := lipgloss.NewRenderer(w, termenv.WithProfile(profile))

	s := Styles{
		Title: r.NewStyle().Bold(true).MarginLeft(2),
		Bold:  r.NewStyle().Bold(true),
	}

	if out.HasDarkBackground() {
		s.Success = r.NewStyle().Foreground(lipgloss.Color("40"))
		s.Error = r.NewStyle().Foreground(lipgloss.Color("196"))
		s.Warning = r.NewStyle().Foreground(lipgloss.Color("214"))
		s.Info = r.NewStyle().Foreground(lipgloss.Color("75"))
		s.Dim = r.NewStyle().Foreground(lipgloss.Color("240"))
		s.Accent = r.NewStyle().Foreground(lipgloss.Color("170"))
	} else {
		s.Success = r.NewStyle().Foreground(lipgloss.Color("28"))
		s.Error = r.NewStyle().Foreground(lipgloss.Color("124"))
		s.Warning = r.NewStyle().Foreground(lipgloss.Color("130"))
		s.Info = r.NewStyle().Foreground(lipgloss.Color("33"))
		s.Dim = r.NewStyle().Foreground(lipgloss.Color("245"))
		s.Accent = r.NewStyle().Foreground(lipgloss.Color("125"))
	}

	return s
}

// DefaultStyles is initialized from stderr for immediate use in CLI commands.
var DefaultStyles = NewStyles(os.Stderr)
