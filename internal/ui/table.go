package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// NewTable renders a styled table with the given headers and rows.
func NewTable(headers []string, rows [][]string) string {
	s := DefaultStyles

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(s.Accent.GetForeground()).Padding(0, 1)
	cellStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	return t.Render()
}
