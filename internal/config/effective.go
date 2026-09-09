package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	sharedconfig "github.com/roshbhatia/go-utils/config"
	"github.com/roshbhatia/go-utils/git"
	"go.yaml.in/yaml/v3"
)

// Origin names where an effective config value came from, in precedence order.
type Origin string

const (
	OriginFlag      Origin = "flag"       // a command-line flag
	OriginEnv       Origin = "env"        // a SESHY_<FIELD> environment override
	OriginGitConfig Origin = "git config" // git config seshy.<key> in the source repo
	OriginFile      Origin = "file"       // the config.yaml
	OriginDefault   Origin = "default"    // seshy's built-in default
)

// Setting is one effective config value and where it came from.
type Setting struct {
	Name   string
	Value  string
	Origin Origin
}

// envVar is the environment override for a config field, matching how the
// shared loader derives it: SESHY_ then each json name upper-snaked.
func envVar(path ...string) string {
	parts := make([]string, len(path)+1)
	parts[0] = "SESHY"
	for i, name := range path {
		parts[i+1] = snakeUpper(name)
	}
	return strings.Join(parts, "_")
}

// snakeUpper uppercases name and inserts an underscore before each interior
// capital, so "branchFormat" becomes "BRANCH_FORMAT".
func snakeUpper(name string) string {
	var out strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' && i > 0 {
			out.WriteByte('_')
		}
		out.WriteRune(r)
	}
	return strings.ToUpper(out.String())
}

// fileKeys returns the top-level keys the config file sets, so an effective
// value can be attributed to the file rather than a default.
func fileKeys() map[string]map[string]bool {
	present := map[string]map[string]bool{"": {}}
	path, err := sharedconfig.Path(configOptions())
	if err != nil {
		return present
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return present
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return present
	}
	for key, value := range raw {
		present[""][key] = true
		if nested, ok := value.(map[string]any); ok {
			present[key] = map[string]bool{}
			for inner := range nested {
				present[key][inner] = true
			}
		}
	}
	return present
}

// Effective returns every config field with its value and origin, in a stable
// order. Origin is env over file over default; a blank value carries the
// default origin. branchFormat is additionally overridable per source repo
// through git config seshy.branchFormat, which this global view does not read.
func Effective() ([]Setting, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	keys := fileKeys()
	value := reflect.ValueOf(*cfg)
	typ := value.Type()
	var settings []Setting
	for i := range value.NumField() {
		field := typ.Field(i)
		name := jsonName(field)
		if name == "" {
			continue
		}
		fieldValue := value.Field(i)
		if fieldValue.Kind() == reflect.Struct {
			nested := fieldValue.Type()
			for j := range fieldValue.NumField() {
				inner := nested.Field(j)
				innerName := jsonName(inner)
				if innerName == "" {
					continue
				}
				settings = append(settings, setting(
					name+"."+innerName,
					fieldValue.Field(j),
					envVar(name, innerName),
					keys[name][innerName],
				))
			}
			continue
		}
		settings = append(settings, setting(name, fieldValue, envVar(name), keys[""][name]))
	}
	return settings, nil
}

// setting builds one Setting, choosing the origin from the environment, the
// file, then the default.
func setting(name string, value reflect.Value, env string, inFile bool) Setting {
	origin := OriginDefault
	if inFile {
		origin = OriginFile
	}
	if _, ok := os.LookupEnv(env); ok {
		origin = OriginEnv
	}
	return Setting{Name: name, Value: renderValue(value), Origin: origin}
}

// renderValue prints a field value on one line: a slice as comma-joined, a
// scalar as itself.
func renderValue(value reflect.Value) string {
	if value.Kind() == reflect.Slice {
		parts := make([]string, value.Len())
		for i := range value.Len() {
			parts[i] = fmt.Sprintf("%v", value.Index(i).Interface())
		}
		return strings.Join(parts, ",")
	}
	return fmt.Sprintf("%v", value.Interface())
}

// jsonName returns a struct field's json name, or "" when it has none.
func jsonName(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "-" {
		return ""
	}
	return name
}

// ResolveBranchFormat returns the branch-name template for a repo and where it
// came from. The command-line branch override is handled by the caller and
// sits above every layer here: SESHY_BRANCH_FORMAT, then git config
// seshy.branchFormat in the source repo, then the config file, then the
// default. repoPath may be empty when no source repo is in play.
func ResolveBranchFormat(repoPath string) (string, Origin) {
	if value, ok := os.LookupEnv(envVar("branchFormat")); ok {
		return value, OriginEnv
	}
	if repoPath != "" {
		if value, set, err := git.ConfigGet(repoPath, "seshy.branchFormat"); err == nil && set {
			return value, OriginGitConfig
		}
	}
	if keys := fileKeys()[""]; keys["branchFormat"] {
		if cfg, err := Load(); err == nil {
			return cfg.BranchFormat, OriginFile
		}
	}
	return defaults().BranchFormat, OriginDefault
}
