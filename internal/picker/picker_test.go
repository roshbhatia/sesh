package picker

import "testing"

func TestItemFilterValue(t *testing.T) {
	i := item{value: "hello"}
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
