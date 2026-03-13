package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — uses ANSI 256 colors that work on both dark and light terminals.
// lipgloss handles NO_COLOR / CLICOLOR automatically.
var (
	ColorPurple = lipgloss.Color("99")
	ColorGreen  = lipgloss.Color("76")
	ColorRed    = lipgloss.Color("204")
	ColorYellow = lipgloss.Color("214")
	ColorBlue   = lipgloss.Color("69")
	ColorGray   = lipgloss.Color("245")
	ColorWhite  = lipgloss.Color("255")
)

// Reusable styles
var (
	Bold       = lipgloss.NewStyle().Bold(true)
	Dim        = lipgloss.NewStyle().Foreground(ColorGray)
	Accent     = lipgloss.NewStyle().Foreground(ColorPurple)
	AccentBold = lipgloss.NewStyle().Foreground(ColorPurple).Bold(true)

	StyleSuccess = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleError   = lipgloss.NewStyle().Foreground(ColorRed)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorYellow)
	StyleInfo    = lipgloss.NewStyle().Foreground(ColorBlue)
)
