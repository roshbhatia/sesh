package cmd

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
)

func TestResolveFormatDefaultsToTable(t *testing.T) {
	got, err := resolveFormat("", nil, listFormats...)
	if err != nil || got != formatTable {
		t.Fatalf("resolveFormat() = %q, %v; want table", got, err)
	}
}

func TestResolveFormatAliasSelectsFormat(t *testing.T) {
	got, err := resolveFormat("", map[string]bool{formatNames: true}, listFormats...)
	if err != nil || got != formatNames {
		t.Fatalf("resolveFormat(--names) = %q, %v; want names", got, err)
	}
}

func TestResolveFormatAliasAgreeingWithFlag(t *testing.T) {
	got, err := resolveFormat(formatJSON, map[string]bool{formatJSON: true}, listFormats...)
	if err != nil || got != formatJSON {
		t.Fatalf("resolveFormat(--json --format json) = %q, %v; want json", got, err)
	}
}

func TestResolveFormatConflictsAreUsageErrors(t *testing.T) {
	cases := []struct {
		format  string
		aliases map[string]bool
	}{
		{"", map[string]bool{formatJSON: true, formatNames: true}},
		{formatPaths, map[string]bool{formatJSON: true}},
		{"yaml", nil},
	}
	for _, tc := range cases {
		_, err := resolveFormat(tc.format, tc.aliases, listFormats...)
		if !errors.Is(err, exitcode.ErrUsage) {
			t.Errorf("resolveFormat(%q, %v) = %v; want a usage error", tc.format, tc.aliases, err)
		}
	}
}

// TestListFormatJSONIsSupersetOfJSONFlag: --format json prints the same bare
// array --json always has, plus id and archived.
func TestListFormatJSONIsSupersetOfJSONFlag(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "fmt-json", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	viaFormat, _, err := runCmd("list", "--format", "json")
	if err != nil {
		t.Fatalf("list --format json: %v", err)
	}
	viaFlag, _, err := runCmd("list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	if viaFormat != viaFlag {
		t.Errorf("--format json and --json differ:\n%s\n%s", viaFormat, viaFlag)
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(viaFormat), &entries); err != nil {
		t.Fatalf("not a bare array: %v\n%s", err, viaFormat)
	}
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	path, _ := session.Resolve("fmt-json")
	want := map[string]any{"name": "fmt-json", "path": path, "repoCount": float64(0), "id": "seshy:fmt-json", "archived": false}
	for key, value := range want {
		if entries[0][key] != value {
			t.Errorf("%s = %v, want %v", key, entries[0][key], value)
		}
	}
	if _, ok := entries[0]["lastModified"].(string); !ok {
		t.Errorf("lastModified missing or not a string: %v", entries[0]["lastModified"])
	}
}

func TestListFormatArchivedFlagsEntries(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "fmt-arch", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	if _, _, err := runCmd("archive", "fmt-arch"); err != nil {
		t.Fatalf("archive: %v", err)
	}
	stdout, _, err := runCmd("list", "--archived", "--format", "json")
	if err != nil {
		t.Fatalf("list --archived --format json: %v", err)
	}
	var entries []sessionJSON
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 || !entries[0].Archived || entries[0].ID != "seshy:fmt-arch" {
		t.Errorf("entries = %+v; want one archived seshy:fmt-arch", entries)
	}
}

func TestListFormatNamesAndPaths(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "fmt-np", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	names, _, err := runCmd("list", "--format", "names")
	if err != nil || names != "fmt-np\n" {
		t.Errorf("list --format names = %q, %v", names, err)
	}
	path, _ := session.Resolve("fmt-np")
	paths, _, err := runCmd("list", "--format", "paths")
	if err != nil || paths != path+"\n" {
		t.Errorf("list --format paths = %q, %v; want %q", paths, err, path)
	}
}

func TestListFormatConflictExitsUsage(t *testing.T) {
	isolatedRoot(t)
	_, _, err := runCmd("list", "--json", "--format", "names")
	if !errors.Is(err, exitcode.ErrUsage) {
		t.Errorf("list --json --format names = %v; want a usage error", err)
	}
}
