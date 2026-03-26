package picker

import (
	"strings"
	"testing"
)

func TestItemFilterValue(t *testing.T) {
	i := item{value: "hello", origIdx: 0}
	if i.FilterValue() != "hello" {
		t.Errorf("expected 'hello', got %q", i.FilterValue())
	}
}

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

func TestDescItemFilterValue(t *testing.T) {
	i := descItem{title: "my-session", desc: "3 repos"}
	if i.FilterValue() != "my-session" {
		t.Errorf("expected 'my-session', got %q", i.FilterValue())
	}
}

func TestDescItemTitle(t *testing.T) {
	i := descItem{title: "session-a", desc: "info"}
	if i.Title() != "session-a" {
		t.Errorf("expected 'session-a', got %q", i.Title())
	}
}

func TestDescItemDescription(t *testing.T) {
	i := descItem{title: "session-a", desc: "2 repos · just now"}
	if i.Description() != "2 repos · just now" {
		t.Errorf("got %q", i.Description())
	}
}

func TestSelectOneWithDescriptionEmptyItems(t *testing.T) {
	_, err := SelectOneWithDescription("prompt", []string{}, []string{})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

func TestMultiItemRenderChecked(t *testing.T) {
	out := RenderMultiItem("my-repo", 0, true, false, -1)
	if !strings.Contains(out, "●") {
		t.Errorf("expected ● for checked item, got %q", out)
	}
}

func TestMultiItemRenderUnchecked(t *testing.T) {
	out := RenderMultiItem("my-repo", 0, false, false, -1)
	if !strings.Contains(out, "○") {
		t.Errorf("expected ○ for unchecked item, got %q", out)
	}
}

func TestItemOrigIdx(t *testing.T) {
	i := item{value: "test", origIdx: 5}
	if i.origIdx != 5 {
		t.Errorf("expected origIdx 5, got %d", i.origIdx)
	}
}
