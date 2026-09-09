package cmd

import (
	"os"

	goconfig "github.com/roshbhatia/go-utils/config"
)

// Plumbing output schemas, generated from the Go structs that render each
// document and held by "sy generate --check". A consumer validates the output
// of the matching command against the file of the same version.
const (
	listSchemaFile   = "schema/list.v1.schema.json"
	statusSchemaFile = "schema/status.v1.schema.json"
	openSchemaFile   = "schema/open.v1.schema.json"
)

// outputSchemas maps each plumbing schema file to its rendered bytes. list is
// a bare array of session entries; status and open are single documents.
func outputSchemas() (map[string][]byte, error) {
	list, err := goconfig.Schema[sessionListJSON]("seshy.list/v1")
	if err != nil {
		return nil, err
	}
	status, err := goconfig.Schema[statusJSON]("seshy.status/v1")
	if err != nil {
		return nil, err
	}
	open, err := goconfig.Schema[openJSON]("seshy.open/v1")
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		listSchemaFile:   list,
		statusSchemaFile: status,
		openSchemaFile:   open,
	}, nil
}

// readRepoFile reads a repository file by path, for tests that check a
// committed artifact.
func readRepoFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
