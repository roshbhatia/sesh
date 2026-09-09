package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// compileSchema compiles one generated schema file by name.
func compileSchema(t *testing.T, file string) *jsonschema.Schema {
	t.Helper()
	data, err := generatedSchemaBytes(t, file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource(file, document); err != nil {
		t.Fatalf("register %s: %v", file, err)
	}
	schema, err := compiler.Compile(file)
	if err != nil {
		t.Fatalf("compile %s: %v", file, err)
	}
	return schema
}

// generatedSchemaBytes reads the committed schema file, so the test proves the
// generated file, not a fresh render, accepts the output.
func generatedSchemaBytes(t *testing.T, file string) ([]byte, error) {
	t.Helper()
	return readRepoFile(filepath.Join("..", file))
}

func validate(t *testing.T, schema *jsonschema.Schema, data string) {
	t.Helper()
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatalf("parse instance: %v", err)
	}
	if err := schema.Validate(instance); err != nil {
		t.Errorf("output fails its schema: %v\n%s", err, data)
	}
}

func TestListOutputMatchesGeneratedSchema(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "sch-a", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	stdout, _, err := runCmd("list", "--format", "json")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	validate(t, compileSchema(t, listSchemaFile), stdout)
}

func TestOpenOutputMatchesGeneratedSchema(t *testing.T) {
	isolatedRoot(t)
	if _, _, err := runCmd("new", "sch-o", "--empty"); err != nil {
		t.Fatalf("new: %v", err)
	}
	stdout, _, err := runCmd("open", "sch-o", "--format", "json")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	validate(t, compileSchema(t, openSchemaFile), stdout)
}

func TestStatusOutputMatchesGeneratedSchema(t *testing.T) {
	isolatedRoot(t)
	repo := filepath.Join(t.TempDir(), "api")
	setupGitRepo(t, repo)
	if _, err := session.Create("sch-s", []string{repo}, session.CreateOpts{BranchFormat: "sy/{{.Session}}/{{.Repo}}"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	stdout, _, err := runCmd("status", "sch-s", "--format", "json")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	validate(t, compileSchema(t, statusSchemaFile), stdout)
}

// TestGeneratedSchemasAreCurrent fails if a struct changed without a regen,
// the same guard "sy generate --check" is.
func TestGeneratedSchemasAreCurrent(t *testing.T) {
	schemas, err := outputSchemas()
	if err != nil {
		t.Fatalf("render schemas: %v", err)
	}
	for file, want := range schemas {
		got, err := readRepoFile(filepath.Join("..", file))
		if err != nil {
			t.Errorf("read %s: %v", file, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is stale; run sy generate", file)
		}
	}
}
