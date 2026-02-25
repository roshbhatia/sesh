package picker

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	checkedItemStyle  = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("40"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type item struct {
	value string
}

func (i item) FilterValue() string { return i.value }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(i.value))
}

// multiItemDelegate renders items with checked/unchecked indicators.
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

	var prefix string
	if isChecked {
		prefix = "◉ "
	} else {
		prefix = "○ "
	}

	var rendered string
	if isCursor {
		rendered = selectedItemStyle.Render("> " + prefix + i.value)
	} else if isChecked {
		rendered = checkedItemStyle.Render(prefix + i.value)
	} else {
		rendered = itemStyle.Render(prefix + i.value)
	}
	fmt.Fprint(w, rendered)
}

// single-select model
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
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selected = i.value
			}
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m singleModel) View() string {
	return "\n" + m.list.View()
}

// multi-select model
type multiModel struct {
	list     list.Model
	selected map[int]bool
	items    []string
	quit     bool
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
	count := len(m.selected)
	title := m.list.Title
	if count > 0 {
		title = fmt.Sprintf("%s (%d selected)", m.list.Title, count)
	}
	m.list.Title = title
	return "\n" + m.list.View()
}

// SelectOne presents an interactive list and returns the selected item.
func SelectOne(prompt string, items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	listItems := make([]list.Item, len(items))
	for i, s := range items {
		listItems[i] = item{value: s}
	}

	l := list.New(listItems, itemDelegate{}, 0, min(len(items)+4, 20))
	l.Title = prompt
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := singleModel{list: l}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("picker failed: %w", err)
	}

	final := result.(singleModel)
	if final.quit || final.selected == "" {
		return "", fmt.Errorf("selection cancelled")
	}
	return final.selected, nil
}

// SelectMany presents an interactive list with multi-select (Space to toggle, Enter to confirm).
func SelectMany(prompt string, items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select")
	}

	listItems := make([]list.Item, len(items))
	for i, s := range items {
		listItems[i] = item{value: s}
	}

	checked := make(map[int]bool)
	l := list.New(listItems, multiItemDelegate{checked: checked}, 0, min(len(items)+4, 20))
	l.Title = prompt + " (Space to toggle, Enter to confirm)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := multiModel{
		list:     l,
		selected: checked,
		items:    items,
	}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("picker failed: %w", err)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
