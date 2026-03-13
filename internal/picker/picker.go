package picker

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// styles holds all lipgloss styles bound to a specific renderer so that
// color decisions (profile, dark/light background) are made against the
// terminal that will actually display the output (stderr).
type styles struct {
	title     lipgloss.Style
	item      lipgloss.Style
	selected  lipgloss.Style
	checked   lipgloss.Style
	dim       lipgloss.Style
	paginator lipgloss.Style
	help      lipgloss.Style
}

// newStyles builds a style-set for the given writer.  It honours:
//   - NO_COLOR / CLICOLOR / CLICOLOR_FORCE environment variables
//   - COLORTERM=truecolor / 24bit
//   - Terminal capability detection via the TERM / TERM_PROGRAM env vars
//   - LS_COLORS is a per-entry file-type convention and does not map
//     cleanly to a TUI picker; we respect the same colour-disable flags
//     that LS_COLORS implementations honour (NO_COLOR, CLICOLOR=0).
func newStyles(w io.Writer) styles {
	// Resolve the colour profile from the environment first so we respect
	// NO_COLOR, CLICOLOR, CLICOLOR_FORCE, and COLORTERM before probing the fd.
	profile := termenv.EnvColorProfile()
	if termenv.EnvNoColor() {
		profile = termenv.Ascii
	}

	// Build a termenv Output against w (stderr) with the resolved profile so
	// HasDarkBackground() queries the right terminal.
	out := termenv.NewOutput(w, termenv.WithColorCache(true), termenv.WithProfile(profile))

	r := lipgloss.NewRenderer(w, termenv.WithProfile(profile))

	s := styles{
		title:     r.NewStyle().MarginLeft(2),
		item:      r.NewStyle().PaddingLeft(4),
		paginator: list.DefaultStyles().PaginationStyle.PaddingLeft(4),
		help:      list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1),
	}

	// Adaptive accent colours: prefer the terminal's own palette entries so
	// they match the user's theme rather than hard-coding specific hues.
	if out.HasDarkBackground() {
		// Bright magenta / bright green are visible on dark backgrounds.
		s.selected = r.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
		s.checked = r.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("40"))
		s.dim = r.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("240"))
	} else {
		// On light backgrounds use darker variants of the same palette slots.
		s.selected = r.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("125"))
		s.checked = r.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("28"))
		s.dim = r.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("245"))
	}

	return s
}

// stderrStyles is initialised once at package init from stderr so that the
// colour profile is detected from the real terminal, not from a pipe.
var stderrStyles = newStyles(os.Stderr)

// keep the _ reference so the import is used even if only stderrStyles is read
var _ = termenv.Ascii

type item struct {
	value string
}

func (i item) FilterValue() string { return i.value }

type itemDelegate struct {
	st styles
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	if index == m.Index() {
		fmt.Fprint(w, d.st.selected.Render("> "+i.value))
	} else {
		fmt.Fprint(w, d.st.item.Render(i.value))
	}
}

// multiItemDelegate renders items with checked/unchecked indicators.
type multiItemDelegate struct {
	checked map[int]bool
	st      styles
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
		prefix = "◉ "
	}

	var rendered string
	switch {
	case isCursor:
		rendered = d.st.selected.Render("> " + prefix + i.value)
	case isChecked:
		rendered = d.st.checked.Render(prefix + i.value)
	default:
		rendered = d.st.item.Render(prefix + i.value)
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

func (m singleModel) View() string {
	return "\n" + m.list.View()
}

// multi-select model
type multiModel struct {
	list      list.Model
	baseTitle string
	selected  map[int]bool
	items     []string
	quit      bool
	st        styles
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
			m.list.SetDelegate(multiItemDelegate{checked: m.selected, st: m.st})
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

// SelectOne presents an interactive list and returns the selected item.
func SelectOne(prompt string, items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	st := newStyles(os.Stderr)

	listItems := make([]list.Item, len(items))
	for i, s := range items {
		listItems[i] = item{value: s}
	}

	l := list.New(listItems, itemDelegate{st: st}, 0, min(len(items)+4, 20))
	l.Title = prompt
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = st.title
	l.Styles.PaginationStyle = st.paginator
	l.Styles.HelpStyle = st.help

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

	st := newStyles(os.Stderr)

	listItems := make([]list.Item, len(items))
	for i, s := range items {
		listItems[i] = item{value: s}
	}

	baseTitle := prompt + " (Space to toggle, Enter to confirm)"

	checked := make(map[int]bool)
	l := list.New(listItems, multiItemDelegate{checked: checked, st: st}, 0, min(len(items)+4, 20))
	l.Title = baseTitle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = st.title
	l.Styles.PaginationStyle = st.paginator
	l.Styles.HelpStyle = st.help

	m := multiModel{
		list:      l,
		baseTitle: baseTitle,
		selected:  checked,
		items:     items,
		st:        st,
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

// descItem supports both title and description for richer display.
type descItem struct {
	title string
	desc  string
}

func (i descItem) Title() string       { return i.title }
func (i descItem) Description() string { return i.desc }
func (i descItem) FilterValue() string { return i.title }

// SelectOneWithDescription presents an interactive list with title+description per item.
// Filtering matches on title only. Returns the selected title string.
func SelectOneWithDescription(prompt string, items []string, descriptions []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	st := newStyles(os.Stderr)

	listItems := make([]list.Item, len(items))
	for i, s := range items {
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		listItems[i] = descItem{title: s, desc: desc}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(listItems, delegate, 0, min(len(items)*3+4, 24))
	l.Title = prompt
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = st.title
	l.Styles.PaginationStyle = st.paginator
	l.Styles.HelpStyle = st.help

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
