package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// NewTable renders a borderless styled table with the given headers and rows.
func NewTable(headers []string, rows [][]string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorPurple).
		Bold(true).
		PaddingRight(2)

	cellStyle := lipgloss.NewStyle().PaddingRight(2)
	dimCellStyle := cellStyle.Foreground(ColorGray)

	t := table.New().
		Headers(headers...).
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			// Dim the last column (typically timestamps)
			if col == len(headers)-1 {
				return dimCellStyle
			}
			return cellStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	return t.Render()
}
