package cmd

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
)

func TestOpenPrintsPathLikePath(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "opn", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	viaOpen, _, err := runCmd("open", "opn")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	viaPath, _, err := runCmd("path", "opn")
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if viaOpen != viaPath {
		t.Errorf("open printed %q, path printed %q", viaOpen, viaPath)
	}
}

func TestOpenFormatJSONIsThePlan(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "opn-json", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	stdout, _, err := runCmd("open", "opn-json", "--format", "json")
	if err != nil {
		t.Fatalf("open --format json: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode: %v\n%s", err, stdout)
	}
	path, _ := session.Resolve("opn-json")
	want := map[string]any{
		"version":      "seshy.open/v1",
		"id":           "seshy:opn-json",
		"cwd":          path,
		"command":      []any{},
		"environment":  map[string]any{"SESHY_SESSION": "opn-json"},
		"successCodes": []any{float64(0)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("open --format json =\n%v\nwant\n%v", got, want)
	}
}

func TestOpenAcceptsSessionID(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "opn-id", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	stdout, _, err := runCmd("open", "seshy:opn-id")
	if err != nil {
		t.Fatalf("open seshy:opn-id: %v", err)
	}
	path, _ := session.Resolve("opn-id")
	if stdout != path+"\n" {
		t.Errorf("open printed %q, want %q", stdout, path)
	}
}

func TestOpenMissingSessionIsNotFound(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("open", "gone", "--format", "json")
	if !errors.Is(err, exitcode.ErrNotFound) {
		t.Errorf("open gone = %v; want ErrNotFound", err)
	}
}

func TestOpenRejectsUnknownFormat(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("open", "x", "--format", "names")
	if !errors.Is(err, exitcode.ErrUsage) {
		t.Errorf("open --format names = %v; want a usage error", err)
	}
}
