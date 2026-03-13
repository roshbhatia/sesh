package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewStylesReturnsNonZero(t *testing.T) {
	st := NewStyles(&bytes.Buffer{})
	_ = st.Title
	_ = st.Success
	_ = st.Error
	_ = st.Warning
	_ = st.Info
	_ = st.Dim
	_ = st.Bold
	_ = st.Accent
}

func TestNewStylesNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	st := NewStyles(&bytes.Buffer{})
	out := st.Success.Render("hello")
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes with NO_COLOR=1, got %q", out)
	}
}

func TestDefaultStylesInitialized(t *testing.T) {
	// Accessing DefaultStyles should not panic
	_ = DefaultStyles.Title
	_ = DefaultStyles.Success
}

func TestSuccessOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Success("done")
	if !strings.Contains(out, "✓") {
		t.Errorf("expected ✓ prefix in Success output, got %q", out)
	}
	if !strings.Contains(out, "done") {
		t.Errorf("expected 'done' in Success output, got %q", out)
	}
}

func TestErrorOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Error("fail")
	if !strings.Contains(out, "✗") {
		t.Errorf("expected ✗ prefix in Error output, got %q", out)
	}
}

func TestWarningOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Warning("watch out")
	if !strings.Contains(out, "⚠") {
		t.Errorf("expected ⚠ prefix in Warning output, got %q", out)
	}
}

func TestInfoOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Info("note")
	if !strings.Contains(out, "note") {
		t.Errorf("expected 'note' in Info output, got %q", out)
	}
}

func TestDimOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Dim("faded")
	if !strings.Contains(out, "faded") {
		t.Errorf("expected 'faded' in Dim output, got %q", out)
	}
}

func TestBoldOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Bold("strong")
	if !strings.Contains(out, "strong") {
		t.Errorf("expected 'strong' in Bold output, got %q", out)
	}
}

func TestNewTableReturnsOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := NewTable(
		[]string{"NAME", "VALUE"},
		[][]string{{"a", "1"}, {"b", "2"}},
	)
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected 'NAME' header in table output, got %q", out)
	}
	if !strings.Contains(out, "a") {
		t.Errorf("expected row data 'a' in table output, got %q", out)
	}
}

func TestFormatf(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	DefaultStyles = NewStyles(&bytes.Buffer{})
	out := Successf("added %d items", 3)
	if !strings.Contains(out, "added 3 items") {
		t.Errorf("expected formatted string in Successf output, got %q", out)
	}
}
