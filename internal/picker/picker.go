package picker

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/roshbhatia/sesh/internal/ui"
)

// ── styles ──────────────────────────────────────────────────────────────

var (
	titleStyle     = lipgloss.NewStyle().MarginLeft(2).Foreground(ui.ColorPurple).Bold(true)
	itemStyle      = lipgloss.NewStyle().PaddingLeft(4)
	selectedStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(ui.ColorPurple).Bold(true)
	checkedStyle   = lipgloss.NewStyle().PaddingLeft(4).Foreground(ui.ColorGreen)
	dimStyle       = lipgloss.NewStyle().PaddingLeft(4).Foreground(ui.ColorGray)
	paginatorStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle      = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

// ── items ───────────────────────────────────────────────────────────────

type item struct{ value string }

func (i item) FilterValue() string { return i.value }

type descItem struct {
	title string
	desc  string
}

func (i descItem) Title() string       { return i.title }
func (i descItem) Description() string { return i.desc }
func (i descItem) FilterValue() string { return i.title }

// ── delegates ───────────────────────────────────────────────────────────

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("▸ "+i.value))
	} else {
		fmt.Fprint(w, itemStyle.Render(i.value))
	}
}

type multiItemDelegate struct {
	checked map[int]bool
}

func (d multiItemDelegate) Height() int                             { return 1 }
func (d multiItemDelegate) Spacing() int                            { return 0 }
func (d multiItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d multiItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	isChecked := d.checked[index]
	isCursor := index == m.Index()

	prefix := "○ "
	if isChecked {
		prefix = "● "
	}

	switch {
	case isCursor:
		fmt.Fprint(w, selectedStyle.Render("▸ "+prefix+i.value))
	case isChecked:
		fmt.Fprint(w, checkedStyle.Render("  "+prefix+i.value))
	default:
		fmt.Fprint(w, itemStyle.Render("  "+prefix+i.value))
	}
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
		return m, nil
	case tea.KeyMsg:
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
	selected  map[int]bool
	items     []string
	quit      bool
}

func (m multiModel) Init() tea.Cmd { return nil }

func (m multiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit
		case " ":
			idx := m.list.Index()
			if m.selected[idx] {
				delete(m.selected, idx)
			} else {
				m.selected[idx] = true
			}
			m.list.SetDelegate(multiItemDelegate{checked: m.selected})
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
		listItems[i] = item{value: s}
	}
	l := newList(listItems, itemDelegate{}, min(len(items)+4, 20))
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
		listItems[i] = item{value: s}
	}
	checked := make(map[int]bool)
	baseTitle := prompt + " (space=toggle, enter=confirm)"
	l := newList(listItems, multiItemDelegate{checked: checked}, min(len(items)+4, 20))
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
	for idx := range final.selected {
		out = append(out, items[idx])
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

	l := newList(listItems, delegate, min(len(items)*3+4, 24))
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
