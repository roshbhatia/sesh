package session

import (
	"bytes"
	"fmt"
	"os/user"
	"strings"
	"text/template"
)

// BranchVars holds variables available in branch name templates.
type BranchVars struct {
	Session string
	Repo    string
	User    string
}

// RenderBranchName evaluates a Go template string with the given session and repo names.
func RenderBranchName(tmpl string, session string, repo string) (string, error) {
	t, err := template.New("branch").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("invalid branch template %q: %w", tmpl, err)
	}

	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	vars := BranchVars{
		Session: session,
		Repo:    repo,
		User:    username,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("executing branch template: %w", err)
	}

	name := buf.String()
	if err := ValidateBranchName(name); err != nil {
		return "", err
	}

	return name, nil
}

// ValidateBranchName checks that a branch name is valid for git.
func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("branch name %q contains spaces", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name %q contains '..'", name)
	}
	for _, c := range name {
		if c < 32 || c == 127 { // control chars
			return fmt.Errorf("branch name %q contains control characters", name)
		}
	}
	for _, bad := range []string{"~", "^", ":", "\\", "?", "*", "["} {
		if strings.Contains(name, bad) {
			return fmt.Errorf("branch name %q contains invalid character %q", name, bad)
		}
	}
	if strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch name %q ends with .lock", name)
	}
	if strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name %q ends with '.'", name)
	}
	return nil
}
