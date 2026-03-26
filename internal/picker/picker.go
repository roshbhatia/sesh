package picker

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/roshbhatia/seshy/internal/ui"
)

// ── styles ──────────────────────────────────────────────────────────────

var (
	titleStyle     = lipgloss.NewStyle().MarginLeft(2).Foreground(ui.ColorPurple).Bold(true)
	paginatorStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle      = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

// ── items ───────────────────────────────────────────────────────────────

// item carries a display value and its original index for stable identity.
type item struct {
	value    string
	origIdx  int
}

func (i item) FilterValue() string { return i.value }

type descItem struct {
	title string
	desc  string
}

func (i descItem) Title() string       { return i.title }
func (i descItem) Description() string { return i.desc }
func (i descItem) FilterValue() string { return i.title }

// ── delegates ───────────────────────────────────────────────────────────

// itemDelegate renders single-select items.
// The FilterValue text is rendered as the primary content, with cursor
// indication via a styled prefix column so filter highlights align.
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	cursor := "  "
	if index == m.Index() {
		cursor = lipgloss.NewStyle().Foreground(ui.ColorPurple).Bold(true).Render("▸ ")
	}
	text := lipgloss.NewStyle().PaddingLeft(2).Render(cursor + i.value)
	fmt.Fprint(w, text)
}

// multiItemDelegate renders multi-select items with check state.
// Uses origIdx from the item for stable identity across filter operations.
type multiItemDelegate struct {
	checked map[int]bool // keyed by item.origIdx
}

func (d multiItemDelegate) Height() int                             { return 1 }
func (d multiItemDelegate) Spacing() int                            { return 0 }
func (d multiItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d multiItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	isChecked := d.checked[i.origIdx]
	isCursor := index == m.Index()

	check := "○"
	if isChecked {
		check = "●"
	}

	var prefix string
	if isCursor {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorPurple).Bold(true).Render("▸ " + check + " ")
	} else if isChecked {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorGreen).Render("  " + check + " ")
	} else {
		prefix = lipgloss.NewStyle().Foreground(ui.ColorGray).Render("  " + check + " ")
	}

	text := lipgloss.NewStyle().PaddingLeft(2).Render(prefix + i.value)
	fmt.Fprint(w, text)
}

// ── models ──────────────────────────────────────────────────────────────

type singleModel struct {
	list     list.Model
	selected string
	quit     bool
}

func (m singleModel) Init() tea.Cmd { return nil }

func (m singleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(min(m.list.Height(), msg.Height-4))
		return m, nil
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit
		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case item:
				m.selected = i.value
			case descItem:
				m.selected = i.title
			}
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m singleModel) View() string { return "\n" + m.list.View() }

type multiModel struct {
	list      list.Model
	baseTitle string
	selected  map[int]bool // keyed by item.origIdx
	items     []string
	quit      bool
}

func (m multiModel) Init() tea.Cmd { return nil }

func (m multiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(min(m.list.Height(), msg.Height-4))
		return m, nil
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit
		case " ":
			if sel, ok := m.list.SelectedItem().(item); ok {
				if m.selected[sel.origIdx] {
					delete(m.selected, sel.origIdx)
				} else {
					m.selected[sel.origIdx] = true
				}
				m.list.SetDelegate(multiItemDelegate{checked: m.selected})
			}
			return m, nil
		case "enter":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m multiModel) View() string {
	title := m.baseTitle
	if count := len(m.selected); count > 0 {
		title = fmt.Sprintf("%s (%d selected)", m.baseTitle, count)
	}
	m.list.Title = title
	return "\n" + m.list.View()
}

// ── public API ──────────────────────────────────────────────────────────

const defaultHeight = 20

func listHeight(itemCount, perItem int) int {
	return min(itemCount*perItem+4, defaultHeight)
}

func newList(items []list.Item, delegate list.ItemDelegate, height int) list.Model {
	l := list.New(items, delegate, 0, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginatorStyle
	l.Styles.HelpStyle = helpStyle
	return l
}

// SelectOne presents a single-select list.
func SelectOne(prompt string, items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}
	listItems := make([]list.Item, len(items))
	for i, s := range items {
		listItems[i] = item{value: s, origIdx: i}
	}
	l := newList(listItems, itemDelegate{}, listHeight(len(items), 1))
	l.Title = prompt

	m := singleModel{list: l}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("picker: %w", err)
	}
	final := result.(singleModel)
	if final.quit || final.selected == "" {
		return "", fmt.Errorf("selection cancelled")
	}
	return final.selected, nil
}

// SelectMany presents a multi-select list (Space to toggle, Enter to confirm).
func SelectMany(prompt string, items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select")
	}
	listItems := make([]list.Item, len(items))
	for i, s := range items {
		listItems[i] = item{value: s, origIdx: i}
	}
	checked := make(map[int]bool)
	baseTitle := prompt + " (space=toggle, enter=confirm)"
	l := newList(listItems, multiItemDelegate{checked: checked}, listHeight(len(items), 1))
	l.Title = baseTitle

	m := multiModel{list: l, baseTitle: baseTitle, selected: checked, items: items}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("picker: %w", err)
	}
	final := result.(multiModel)
	if final.quit {
		return nil, fmt.Errorf("selection cancelled")
	}
	out := make([]string, 0, len(final.selected))
	for origIdx := range final.selected {
		out = append(out, items[origIdx])
	}
	return out, nil
}

// SelectOneWithDescription presents a single-select list with description lines.
func SelectOneWithDescription(prompt string, items []string, descriptions []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}
	listItems := make([]list.Item, len(items))
	for i, s := range items {
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		listItems[i] = descItem{title: s, desc: desc}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(ui.ColorPurple).BorderLeftForeground(ui.ColorPurple)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(ui.ColorGray).BorderLeftForeground(ui.ColorPurple)

	l := newList(listItems, delegate, listHeight(len(items), 3))
	l.Title = prompt

	m := singleModel{list: l}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("picker: %w", err)
	}
	final := result.(singleModel)
	if final.quit || final.selected == "" {
		return "", fmt.Errorf("selection cancelled")
	}
	return final.selected, nil
}

// ── testing helpers ─────────────────────────────────────────────────────

// RenderMultiItem renders a multi-select item for testing purposes.
func RenderMultiItem(value string, origIdx int, isChecked bool, isCursor bool, cursorIdx int) string {
	d := multiItemDelegate{checked: map[int]bool{}}
	if isChecked {
		d.checked[origIdx] = true
	}

	var buf strings.Builder
	// Create a minimal list model for rendering
	listItems := []list.Item{item{value: value, origIdx: origIdx}}
	l := list.New(listItems, d, 80, 10)
	if isCursor {
		// Index 0 is the cursor position
	}

	d.Render(&buf, l, 0, listItems[0])
	return buf.String()
}
