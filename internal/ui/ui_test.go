package ui

import (
	"strings"
	"testing"
)

func TestSuccessContainsCheckmark(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := Success("done")
	if !strings.Contains(out, "✓") || !strings.Contains(out, "done") {
		t.Errorf("got %q", out)
	}
}

func TestErrorContainsCross(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := Error("fail")
	if !strings.Contains(out, "✗") || !strings.Contains(out, "fail") {
		t.Errorf("got %q", out)
	}
}

func TestWarningContainsIcon(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := Warning("watch out")
	if !strings.Contains(out, "⚠") || !strings.Contains(out, "watch out") {
		t.Errorf("got %q", out)
	}
}

func TestInfoContainsMessage(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := Info("note")
	if !strings.Contains(out, "note") {
		t.Errorf("got %q", out)
	}
}

func TestFaintContainsMessage(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := Faint("faded")
	if !strings.Contains(out, "faded") {
		t.Errorf("got %q", out)
	}
}

func TestTableRendersHeadersAndRows(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := NewTable(
		[]string{"NAME", "VALUE"},
		[][]string{{"a", "1"}, {"b", "2"}},
	)
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "a") {
		t.Errorf("got %q", out)
	}
}

func TestSuccessf(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := Successf("added %d items", 3)
	if !strings.Contains(out, "added 3 items") {
		t.Errorf("got %q", out)
	}
}

func TestAccentBoldRenders(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := AccentBold.Render("hello")
	if !strings.Contains(out, "hello") {
		t.Errorf("got %q", out)
	}
}
