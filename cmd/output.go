package cmd

import (
	"io"
	"strings"

	"github.com/roshbhatia/go-utils/cell"
	"github.com/roshbhatia/go-utils/ui"
)

// table writes header and rows to w through cell.Table, which pads by display
// width and strips every ANSI sequence when w is not a color-capable
// terminal, so a piped table is plain bytes. Cells may carry their own color.
//
// The header is padded before it is colored. That keeps the bytes a terminal
// receives identical to what sy has always printed, where the padding sits
// inside the escape sequence rather than after it.
func table(w io.Writer, header []string, rows [][]string) error {
	all := make([][]string, 0, len(rows)+1)
	if header != nil {
		widths := make([]int, len(header))
		for column, cellText := range header {
			widths[column] = cell.Width(cellText)
		}
		for _, row := range rows {
			for column, cellText := range row {
				if column < len(widths) {
					widths[column] = max(widths[column], cell.Width(cellText))
				}
			}
		}
		colored := make([]string, len(header))
		for column, cellText := range header {
			if column < len(header)-1 {
				cellText += strings.Repeat(" ", widths[column]-cell.Width(cellText))
			}
			colored[column] = ui.StdoutColor(ui.ColorPurple, cellText)
		}
		all = append(all, colored)
	}
	return cell.Table(w, append(all, rows...))
}
