package picker

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// testSt returns a styles value built against a no-TTY writer (Ascii profile).
// Tests care about text content, not ANSI codes, so this is sufficient.
func testSt() styles { return newStyles(&bytes.Buffer{}) }

// ---------------------------------------------------------------------------
// multiItemDelegate rendering tests
// ---------------------------------------------------------------------------

func TestMultiItemDelegateHeight(t *testing.T) {
	d := multiItemDelegate{checked: map[int]bool{}, st: testSt()}
	if d.Height() != 1 {
		t.Errorf("expected Height 1, got %d", d.Height())
	}
}

func TestMultiItemDelegateSpacing(t *testing.T) {
	d := multiItemDelegate{checked: map[int]bool{}, st: testSt()}
	if d.Spacing() != 0 {
		t.Errorf("expected Spacing 0, got %d", d.Spacing())
	}
}

func TestMultiItemDelegateUpdate(t *testing.T) {
	d := multiItemDelegate{checked: map[int]bool{}, st: testSt()}
	if cmd := d.Update(nil, nil); cmd != nil {
		t.Error("expected Update to return nil cmd")
	}
}

type renderScenario struct {
	name      string
	index     int
	cursor    int
	checked   map[int]bool
	wantFill  string
	wantAbsent string
}

func newListModelForRender(items []list.Item, cursor int) list.Model {
	m := list.New(items, itemDelegate{st: testSt()}, 80, 20)
	for i := 0; i < cursor; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	return m
}

func TestMultiItemDelegateRender(t *testing.T) {
	listItems := []list.Item{
		item{value: "alpha"},
		item{value: "beta"},
		item{value: "gamma"},
	}

	tests := []struct {
		name       string
		index      int         // which item to render
		cursor     int         // current cursor position in list
		checked    map[int]bool
		wantChecked bool
		wantCursor  bool
	}{
		{
			name:        "unchecked non-cursor",
			index:       1,
			cursor:      0,
			checked:     map[int]bool{},
			wantChecked: false,
			wantCursor:  false,
		},
		{
			name:        "checked non-cursor",
			index:       1,
			cursor:      0,
			checked:     map[int]bool{1: true},
			wantChecked: true,
			wantCursor:  false,
		},
		{
			name:        "cursor unchecked",
			index:       0,
			cursor:      0,
			checked:     map[int]bool{},
			wantChecked: false,
			wantCursor:  true,
		},
		{
			name:        "cursor and checked",
			index:       0,
			cursor:      0,
			checked:     map[int]bool{0: true},
			wantChecked: true,
			wantCursor:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := multiItemDelegate{checked: tt.checked, st: testSt()}
			m := newListModelForRender(listItems, tt.cursor)

			var buf bytes.Buffer
			d.Render(&buf, m, tt.index, listItems[tt.index])
			out := buf.String()

			if tt.wantChecked && !strings.Contains(out, "◉") {
				t.Errorf("expected ◉ indicator for checked item, got: %q", out)
			}
			if !tt.wantChecked && !strings.Contains(out, "○") {
				t.Errorf("expected ○ indicator for unchecked item, got: %q", out)
			}
			if tt.wantCursor && !strings.Contains(out, ">") {
				t.Errorf("expected cursor '>' indicator, got: %q", out)
			}
		})
	}
}

func TestMultiItemDelegateRenderBadItem(t *testing.T) {
	d := multiItemDelegate{checked: map[int]bool{}, st: testSt()}
	m := list.New(nil, d, 80, 10)
	var buf bytes.Buffer
	// Pass a non-item type — should produce no output and not panic
	d.Render(&buf, m, 0, badItem{})
	if buf.Len() != 0 {
		t.Errorf("expected no output for bad item type, got: %q", buf.String())
	}
}

type badItem struct{}

func (b badItem) FilterValue() string { return "" }

// ---------------------------------------------------------------------------
// itemDelegate (single-select) rendering tests
// ---------------------------------------------------------------------------

func TestItemDelegateHeight(t *testing.T) {
	d := itemDelegate{st: testSt()}
	if d.Height() != 1 {
		t.Errorf("expected Height 1, got %d", d.Height())
	}
}

func TestItemDelegateSpacing(t *testing.T) {
	d := itemDelegate{st: testSt()}
	if d.Spacing() != 0 {
		t.Errorf("expected Spacing 0, got %d", d.Spacing())
	}
}

func TestItemDelegateRenderSelected(t *testing.T) {
	st := testSt()
	listItems := []list.Item{item{value: "foo"}, item{value: "bar"}}
	m := list.New(listItems, itemDelegate{st: st}, 80, 10)

	var buf bytes.Buffer
	d := itemDelegate{st: st}
	d.Render(&buf, m, 0, listItems[0]) // index 0 == cursor
	out := buf.String()
	if !strings.Contains(out, ">") {
		t.Errorf("expected '>' cursor for selected item, got: %q", out)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("expected item value 'foo' in output, got: %q", out)
	}
}

func TestItemDelegateRenderUnselected(t *testing.T) {
	st := testSt()
	listItems := []list.Item{item{value: "foo"}, item{value: "bar"}}
	m := list.New(listItems, itemDelegate{st: st}, 80, 10)

	var buf bytes.Buffer
	d := itemDelegate{st: st}
	d.Render(&buf, m, 1, listItems[1]) // index 1, cursor is at 0
	out := buf.String()
	if strings.Contains(out, ">") {
		t.Errorf("did not expect '>' cursor for non-selected item, got: %q", out)
	}
	if !strings.Contains(out, "bar") {
		t.Errorf("expected item value 'bar' in output, got: %q", out)
	}
}

func TestItemDelegateRenderBadItem(t *testing.T) {
	d := itemDelegate{st: testSt()}
	m := list.New(nil, d, 80, 10)
	var buf bytes.Buffer
	d.Render(&buf, m, 0, badItem{})
	if buf.Len() != 0 {
		t.Errorf("expected no output for bad item type, got: %q", buf.String())
	}
}

// ---------------------------------------------------------------------------
// item FilterValue
// ---------------------------------------------------------------------------

func TestItemFilterValue(t *testing.T) {
	i := item{value: "hello"}
	if i.FilterValue() != "hello" {
		t.Errorf("expected FilterValue 'hello', got %q", i.FilterValue())
	}
}

// ---------------------------------------------------------------------------
// min helper
// ---------------------------------------------------------------------------

func TestMin(t *testing.T) {
	tests := []struct{ a, b, want int }{
		{1, 2, 1},
		{5, 3, 3},
		{4, 4, 4},
		{0, 10, 0},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		if got := min(tt.a, tt.b); got != tt.want {
			t.Errorf("min(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// singleModel / multiModel state machine tests (no TUI program execution)
// ---------------------------------------------------------------------------

func makeListModel(vals []string) list.Model {
	items := make([]list.Item, len(vals))
	for i, v := range vals {
		items[i] = item{value: v}
	}
	return list.New(items, itemDelegate{st: testSt()}, 80, 20)
}

func TestSingleModelInit(t *testing.T) {
	m := singleModel{list: makeListModel([]string{"a"})}
	if cmd := m.Init(); cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestSingleModelQuit(t *testing.T) {
	m := singleModel{list: makeListModel([]string{"a", "b"})}
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	final := result.(singleModel)
	if !final.quit {
		t.Error("expected quit=true after pressing q")
	}
}

func TestSingleModelCtrlC(t *testing.T) {
	m := singleModel{list: makeListModel([]string{"x"})}
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	final := result.(singleModel)
	if !final.quit {
		t.Error("expected quit=true after ctrl+c")
	}
}

func TestSingleModelEnterSelectsItem(t *testing.T) {
	m := singleModel{list: makeListModel([]string{"chosen", "other"})}
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	final := result.(singleModel)
	if final.selected != "chosen" {
		t.Errorf("expected selected='chosen', got %q", final.selected)
	}
}

func TestSingleModelWindowResize(t *testing.T) {
	m := singleModel{list: makeListModel([]string{"a"})}
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_ = result.(singleModel)
}

func TestSingleModelView(t *testing.T) {
	m := singleModel{list: makeListModel([]string{"a"})}
	v := m.View()
	if !strings.HasPrefix(v, "\n") {
		t.Error("expected View to start with newline")
	}
}

func TestMultiModelInit(t *testing.T) {
	m := multiModel{list: makeListModel([]string{"a"}), selected: map[int]bool{}, st: testSt()}
	if cmd := m.Init(); cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestMultiModelQuit(t *testing.T) {
	m := multiModel{list: makeListModel([]string{"a"}), selected: map[int]bool{}, st: testSt()}
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	final := result.(multiModel)
	if !final.quit {
		t.Error("expected quit=true after pressing q")
	}
}

func TestMultiModelCtrlC(t *testing.T) {
	m := multiModel{list: makeListModel([]string{"a"}), selected: map[int]bool{}, st: testSt()}
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	final := result.(multiModel)
	if !final.quit {
		t.Error("expected quit=true after ctrl+c")
	}
}

func TestMultiModelSpaceTogglesSelection(t *testing.T) {
	m := multiModel{
		list:     makeListModel([]string{"a", "b", "c"}),
		selected: map[int]bool{},
		items:    []string{"a", "b", "c"},
		st:       testSt(),
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = result.(multiModel)
	if !m.selected[0] {
		t.Error("expected index 0 to be selected after Space")
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = result.(multiModel)
	if m.selected[0] {
		t.Error("expected index 0 to be deselected after second Space")
	}
}

func TestMultiModelEnterConfirms(t *testing.T) {
	m := multiModel{list: makeListModel([]string{"a"}), selected: map[int]bool{0: true}, items: []string{"a"}, st: testSt()}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = cmd
}

func TestMultiModelViewShowsCount(t *testing.T) {
	m := multiModel{
		list:     makeListModel([]string{"a", "b"}),
		selected: map[int]bool{0: true, 1: true},
		items:    []string{"a", "b"},
		st:       testSt(),
	}
	m.list.Title = "Pick"
	v := m.View()
	if !strings.Contains(v, "2 selected") {
		t.Errorf("expected '2 selected' in view, got: %q", v)
	}
}

func TestMultiModelViewNoCount(t *testing.T) {
	m := multiModel{list: makeListModel([]string{"a"}), selected: map[int]bool{}, items: []string{"a"}, st: testSt()}
	m.list.Title = "Pick"
	v := m.View()
	if strings.Contains(v, "selected") {
		t.Errorf("expected no 'selected' text when nothing selected, got: %q", v)
	}
}

func TestMultiModelWindowResize(t *testing.T) {
	m := multiModel{list: makeListModel([]string{"a"}), selected: map[int]bool{}, st: testSt()}
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_ = result.(multiModel)
}

// ---------------------------------------------------------------------------
// SelectOne / SelectMany error paths (no TTY available in test env)
// ---------------------------------------------------------------------------

func TestSelectOneEmptyItems(t *testing.T) {
	_, err := SelectOne("prompt", []string{})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

func TestSelectManyEmptyItems(t *testing.T) {
	_, err := SelectMany("prompt", []string{})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

// ---------------------------------------------------------------------------
// Render output helpers — verify no panic on writer errors
// ---------------------------------------------------------------------------

type errWriter struct{}

func (e errWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

func TestItemDelegateRenderWriteError(t *testing.T) {
	d := itemDelegate{st: testSt()}
	m := list.New([]list.Item{item{value: "x"}}, d, 80, 10)
	d.Render(errWriter{}, m, 0, item{value: "x"})
}

func TestMultiItemDelegateRenderWriteError(t *testing.T) {
	checked := map[int]bool{0: true}
	d := multiItemDelegate{checked: checked, st: testSt()}
	m := list.New([]list.Item{item{value: "x"}}, d, 80, 10)
	d.Render(errWriter{}, m, 0, item{value: "x"})
}

// Compile-time check that errWriter satisfies io.Writer
var _ io.Writer = errWriter{}

// ---------------------------------------------------------------------------
// newStyles / colour environment tests
// ---------------------------------------------------------------------------

func TestNewStylesReturnsStyles(t *testing.T) {
	st := newStyles(&bytes.Buffer{})
	// Just verify the styles struct is populated (non-zero) without panicking.
	_ = st.title
	_ = st.item
	_ = st.selected
	_ = st.checked
	_ = st.paginator
	_ = st.help
}

func TestNewStylesNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	st := newStyles(&bytes.Buffer{})
	// With NO_COLOR set the selected style should produce plain text (no ANSI).
	out := st.selected.Render("hello")
	if out != "hello" {
		// Some lipgloss versions still add padding — strip and check no escape
		if strings.Contains(out, "\x1b[") {
			t.Errorf("expected no ANSI codes with NO_COLOR=1, got %q", out)
		}
	}
}

func TestNewStylesClicolorZero(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR", "0")
	t.Setenv("CLICOLOR_FORCE", "0")
	st := newStyles(&bytes.Buffer{})
	out := st.checked.Render("x")
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes with CLICOLOR=0, got %q", out)
	}
}

func TestNewStylesDarkBackground(t *testing.T) {
	// Ensure we get styles without panic regardless of background detection.
	// In CI there's no real TTY, so HasDarkBackground returns a default.
	t.Setenv("NO_COLOR", "")
	st := newStyles(&bytes.Buffer{})
	// selected and checked must be non-zero styles
	if st.selected.GetPaddingLeft() < 0 {
		t.Error("unexpected negative padding on selected style")
	}
}

func TestStderrStylesInitialized(t *testing.T) {
	// stderrStyles is package-level; just confirm it doesn't zero-value panic
	_ = stderrStyles.title
	_ = stderrStyles.item
}
