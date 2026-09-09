package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/roshbhatia/go-utils/provider"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// rosterCatalogSchema compiles the pinned copy of roster's catalog schema.
// The flake check keeps the copy identical to the roster release it names.
func rosterCatalogSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	const name = "roster.catalog.v1.schema.json"
	data, err := os.ReadFile(filepath.Join("..", "schema", name))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource(name, document); err != nil {
		t.Fatalf("register schema: %v", err)
	}
	schema, err := compiler.Compile(name)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return schema
}

func assertValidCatalog(t *testing.T, schema *jsonschema.Schema, data []byte) {
	t.Helper()
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse instance: %v", err)
	}
	if err := schema.Validate(instance); err != nil {
		t.Errorf("catalog fails roster's schema: %v\n%s", err, data)
	}
}

// TestSourceListValidatesAgainstRosterSchema is the contract with roster: the
// document sy prints is accepted by the schema roster validates with.
func TestSourceListValidatesAgainstRosterSchema(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "api")
	setupGitRepo(t, repo)
	if _, err := session.Create("cat-a", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, _, err := runCmd("new", "cat-b", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	stdout, _, err := runCmd("source", "list")
	if err != nil {
		t.Fatalf("source list: %v", err)
	}
	assertValidCatalog(t, rosterCatalogSchema(t), []byte(stdout))

	var got catalogJSON
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Version != "roster.catalog/v1" || got.Source != "seshy" || got.TTL != "10s" {
		t.Errorf("header = %+v", got)
	}
	if got.Display != (catalogDisplayJSON{Label: "sessions", Glyph: "cod_briefcase", Order: 10}) {
		t.Errorf("display = %+v", got.Display)
	}
	if len(got.Groups) != 0 || len(got.Rows) != 2 {
		t.Fatalf("want 0 groups and 2 rows, got %d and %d", len(got.Groups), len(got.Rows))
	}
	pathA, _ := session.Resolve("cat-a")
	row := got.Rows[0]
	if row.ID != "seshy:cat-a" || row.Workspace != "cat-a" || row.Label != "cat-a" || row.Kind != "session" || row.Cwd != pathA {
		t.Errorf("row = %+v", row)
	}
	if row.Group != nil || row.Host != nil || row.Status != nil || row.Pane != nil {
		t.Errorf("nullable fields must be null: %+v", row)
	}
	wantPlan := planJSON{Command: []string{}, Cwd: pathA, Environment: map[string]string{"SESHY_SESSION": "cat-a"}, SuccessCodes: []int{0}}
	if !reflect.DeepEqual(row.Spawn.Plan, wantPlan) || row.Spawn.Hop.Kind != "local" {
		t.Errorf("spawn = %+v, want plan %+v over a local hop", row.Spawn, wantPlan)
	}
	if row.Meta["repoCount"] != float64(1) {
		t.Errorf("meta.repoCount = %v, want 1", row.Meta["repoCount"])
	}
	if _, err := time.Parse(time.RFC3339, row.Meta["lastModified"].(string)); err != nil {
		t.Errorf("meta.lastModified: %v", err)
	}
}

func TestSourceListEmptyIsStillValid(t *testing.T) {
	isolatedRoot(t)
	stdout, _, err := runCmd("source", "list")
	if err != nil {
		t.Fatalf("source list: %v", err)
	}
	assertValidCatalog(t, rosterCatalogSchema(t), []byte(stdout))
	if !strings.Contains(stdout, `"rows": []`) {
		t.Errorf("empty catalog must print rows as [], got:\n%s", stdout)
	}
}

func TestSourceListExcludesArchived(t *testing.T) {
	isolatedRoot(t)
	for _, name := range []string{"live", "shelved"} {
		if _, _, err := runCmd("new", name, "--empty"); err != nil {
			t.Fatalf("new %s: %v", name, err)
		}
	}
	if _, _, err := runCmd("archive", "shelved"); err != nil {
		t.Fatalf("archive: %v", err)
	}
	stdout, _, err := runCmd("source", "list")
	if err != nil {
		t.Fatalf("source list: %v", err)
	}
	var got catalogJSON
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Rows) != 1 || got.Rows[0].ID != "seshy:live" {
		t.Errorf("rows = %+v; want only seshy:live", got.Rows)
	}
}

// requestFrame builds one provider/v1 request line.
func requestFrame(t *testing.T, capability string, input any) string {
	t.Helper()
	request := provider.Request{Version: provider.Version, Kind: provider.FrameRequest, RequestID: "req-1", Capability: capability}
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		request.Input = data
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return string(data) + "\n"
}

// serve runs the provider on one frame and decodes its single result line.
func serve(t *testing.T, frame string) provider.Result {
	t.Helper()
	var out bytes.Buffer
	if err := serveProvider(strings.NewReader(frame), &out, time.Now()); err != nil {
		t.Fatalf("serveProvider: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("want exactly one result line, got %d:\n%s", len(lines), out.String())
	}
	var result provider.Result
	decoder := json.NewDecoder(strings.NewReader(lines[0]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode result: %v\n%s", err, lines[0])
	}
	if result.Version != provider.Version || result.Kind != provider.FrameResult || result.RequestID != "req-1" {
		t.Errorf("result frame header = %+v", result)
	}
	return result
}

func TestProviderValidateIsOK(t *testing.T) {
	isolatedRoot(t)
	result := serve(t, requestFrame(t, "provider.validate", nil))
	if result.Status != provider.ResultOK || string(result.Output) != `{"ok":true}` {
		t.Errorf("provider.validate = %+v", result)
	}
}

func TestProviderSourceListMatchesSourceListCommand(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "prov", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	result := serve(t, requestFrame(t, "source.list", nil))
	if result.Status != provider.ResultOK {
		t.Fatalf("source.list = %+v", result)
	}
	assertValidCatalog(t, rosterCatalogSchema(t), result.Output)
	var got catalogJSON
	if err := json.Unmarshal(result.Output, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Rows) != 1 || got.Rows[0].ID != "seshy:prov" {
		t.Errorf("rows = %+v", got.Rows)
	}
}

func TestProviderSourceOpenReturnsTheRow(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "row", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	result := serve(t, requestFrame(t, "source.open", map[string]string{"id": "seshy:row"}))
	if result.Status != provider.ResultOK {
		t.Fatalf("source.open = %+v", result)
	}
	var row catalogRowJSON
	if err := json.Unmarshal(result.Output, &row); err != nil {
		t.Fatalf("decode: %v", err)
	}
	path, _ := session.Resolve("row")
	if row.ID != "seshy:row" || row.Cwd != path || row.Spawn.Plan.Cwd != path || row.Spawn.Plan.Environment["SESHY_SESSION"] != "row" {
		t.Errorf("row = %+v", row)
	}
	listed := serve(t, requestFrame(t, "source.list", nil))
	var catalog catalogJSON
	if err := json.Unmarshal(listed.Output, &catalog); err != nil {
		t.Fatalf("decode catalog: %v", err)
	}
	if !reflect.DeepEqual(catalog.Rows[0], row) {
		t.Errorf("source.open row differs from the listed row:\n%+v\n%+v", row, catalog.Rows[0])
	}
}

func TestProviderErrorsAreResultFrames(t *testing.T) {
	isolatedRoot(t)
	cases := []struct {
		name, capability, message string
		input                     any
	}{
		{"unknown capability", "session.open", `unsupported capability "session.open"`, nil},
		{"missing id", "source.open", "input.id is required", map[string]string{}},
		{"foreign id", "source.open", "row 'tether:x' is not a seshy id", map[string]string{"id": "tether:x"}},
		{"gone session", "source.open", "session 'gone' not found", map[string]string{"id": "seshy:gone"}},
	}
	for _, tc := range cases {
		result := serve(t, requestFrame(t, tc.capability, tc.input))
		if result.Status != provider.ResultError || result.Message != tc.message || len(result.Output) != 0 {
			t.Errorf("%s: result = %+v; want error %q", tc.name, result, tc.message)
		}
	}
}

func TestProviderRejectsNonRequestFrames(t *testing.T) {
	isolatedRoot(t)
	cases := map[string]string{
		"not json":    "nope\n",
		"wrong kind":  `{"version":"provider/v1","kind":"event","requestId":"r","capability":"source.list"}`,
		"no id":       `{"version":"provider/v1","kind":"request","capability":"source.list"}`,
		"old version": `{"version":"provider/v0","kind":"request","requestId":"r","capability":"source.list"}`,
		"unknown key": `{"version":"provider/v1","kind":"request","requestId":"r","capability":"source.list","extra":1}`,
	}
	for name, frame := range cases {
		var out bytes.Buffer
		err := serveProvider(strings.NewReader(frame), &out, time.Now())
		if !errors.Is(err, exitcode.ErrUsage) {
			t.Errorf("%s: err = %v; want a usage error", name, err)
		}
		if out.Len() != 0 {
			t.Errorf("%s: wrote %q; a frame that cannot be answered gets no result", name, out.String())
		}
	}
}

// TestManifestMatchesTheProvider: the manifest roster discovers names the
// binary and exactly the capabilities dispatch answers.
func TestManifestMatchesTheProvider(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "share", "seshy", "providers", "seshy.yaml"))
	if err != nil {
		t.Fatalf("open manifest: %v", err)
	}
	defer file.Close()
	manifest, err := provider.Decode(file, ".yaml")
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("manifest is not a valid provider/v1 manifest: %v", err)
	}
	if manifest.Kind != "source" || manifest.Name != catalogSource {
		t.Errorf("kind/name = %q/%q; want source/seshy", manifest.Kind, manifest.Name)
	}
	if !reflect.DeepEqual(manifest.Command, []string{"sy", "provider"}) {
		t.Errorf("command = %v; want [sy provider]", manifest.Command)
	}
	if manifest.Defaults.Timeout.Duration() != 5*time.Second {
		t.Errorf("defaults.timeout = %s; want 5s", manifest.Defaults.Timeout.Duration())
	}
	want := map[string]bool{capabilityValidate: true, capabilityList: true, capabilityOpen: true}
	for name := range manifest.Actions {
		if !want[name] {
			t.Errorf("manifest declares %q, which dispatch does not answer", name)
		}
		delete(want, name)
	}
	for name := range want {
		t.Errorf("manifest lacks %q, which dispatch answers", name)
	}
}
