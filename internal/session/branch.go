package session

import (
	"bytes"
	"fmt"
	"os/user"
	"text/template"

	"github.com/roshbhatia/go-utils/git"
	"github.com/roshbhatia/seshy/internal/tmpl"
)

// RenderBranchName evaluates a Go template string with the given session and
// repo names, then checks the result against git's own branch-name rules.
func RenderBranchName(tmplStr string, sessionName string, repo string) (string, error) {
	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	data := tmpl.TemplateData{
		Session: sessionName,
		Repo:    repo,
		User:    username,
	}

	parsed, err := template.New("branch").Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("invalid branch template %q: %w", tmplStr, err)
	}
	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("invalid branch template %q: %w", tmplStr, err)
	}
	name := rendered.String()

	if err := CheckBranchName(name); err != nil {
		return "", err
	}
	return name, nil
}

// CheckBranchName asks git whether name is a valid branch name. A rejection
// is a *BranchNameError wrapping git's *RefFormatError, so git.ExitStatus
// still reads git's own status from it.
func CheckBranchName(name string) error {
	err := git.CheckRefFormat(name)
	if err == nil {
		return nil
	}
	if git.ExitStatus(err) < 0 {
		return err
	}
	return &BranchNameError{Name: name, Err: err}
}

// BranchNameError reports a branch name that git rejects. Its message is
// git's, minus the "fatal:" label that the caller adds when it prints.
type BranchNameError struct {
	Name string
	Err  error
}

func (e *BranchNameError) Error() string {
	return fmt.Sprintf("'%s' is not a valid branch name", e.Name)
}

func (e *BranchNameError) Unwrap() error { return e.Err }
