package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/roshbhatia/go-utils/completion"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var generateCheck bool

var completionCmd = &cobra.Command{
	Use:       "completion <bash|zsh|fish|nu>",
	Short:     "Generate shell completion",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "nu"},
	RunE: func(_ *cobra.Command, args []string) error {
		generated, err := generatedCompletion(args[0])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, generated)
		return err
	},
}

func generatedCompletion(shell string) (string, error) {
	generated, err := completion.Generate(shell, completionMetadata())
	if err != nil || shell != "nu" {
		return generated, err
	}
	const rootStart = `export extern "sy" [`
	const rootEnd = "\n]\n\n"
	if !strings.HasPrefix(generated, rootStart) {
		return "", errors.New("generated Nushell completion has no sy root signature")
	}
	end := strings.Index(generated, rootEnd)
	if end < 0 {
		return "", errors.New("generated Nushell completion has an incomplete sy root signature")
	}
	signature := generated[:end+2]
	signature = strings.Replace(signature, `export extern "sy"`, "export def --env sy", 1)
	body, err := renderShellIntegration("nu", nushellBody)
	if err != nil {
		return "", err
	}
	return signature + body + "\n" + generated[end+len(rootEnd):], nil
}

var completionValuesCmd = &cobra.Command{
	Use:    "__values <active|archived|repositories|sessions|shells|add|new|remove> [context]",
	Hidden: true,
	Args:   cobra.RangeArgs(1, 2),
	RunE: func(_ *cobra.Command, args []string) error {
		context := ""
		if len(args) == 2 {
			context = args[1]
		}
		values, err := completionValues(args[0], context)
		if err != nil {
			return err
		}
		for _, value := range values {
			if !strings.ContainsRune(value, '\n') {
				if _, err := fmt.Fprintln(os.Stdout, value); err != nil {
					return err
				}
			}
		}
		return nil
	},
}

var generateCmd = &cobra.Command{
	Use:    "generate",
	Short:  "Generate README docs, schema, and completions",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return generateRepository(generateCheck)
	},
}

func init() {
	generateCmd.Flags().BoolVar(&generateCheck, "check", false, "Fail when generated files are stale")
	rootCmd.AddCommand(completionCmd, completionValuesCmd, generateCmd)
}

func completionValues(kind, context string) ([]string, error) {
	switch kind {
	case "active":
		items, err := session.List()
		return sessionNames(items), err
	case "archived":
		items, err := session.ListArchived()
		return sessionNames(items), err
	case "repositories":
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		items, err := runSource(cfg.RepoSource)
		if err != nil {
			return nil, err
		}
		return prependDefaults(cfg.DefaultRepos, items), nil
	case "sessions":
		if strings.Contains(context, "--archived") {
			return completionValues("archived", context)
		}
		return completionValues("active", context)
	case "shells":
		return []string{"bash", "zsh", "fish", "nu"}, nil
	case "add":
		args := commandArguments(context, "add", map[string]bool{"--branch": true, "-b": true})
		if len(args) == 0 {
			return completionValues("active", context)
		}
		return completionValues("repositories", context)
	case "new":
		args := commandArguments(context, "new", map[string]bool{"--branch": true, "-b": true})
		if len(args) == 0 {
			return nil, nil
		}
		return completionValues("repositories", context)
	case "remove":
		args := commandArguments(context, "remove", nil)
		if len(args) == 0 {
			return completionValues("active", context)
		}
		found, err := session.Resolve(args[0])
		if err != nil {
			return nil, nil
		}
		repos := session.GetSessionRepoInfos(found)
		values := make([]string, len(repos))
		for index, repo := range repos {
			values[index] = repo.Name
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unknown completion value set %q", kind)
	}
}

func commandArguments(context, name string, valueFlags map[string]bool) []string {
	words := strings.Fields(context)
	if len(words) > 0 && !unicode.IsSpace(rune(context[len(context)-1])) {
		// The last token is the value being completed. Only earlier tokens can
		// decide which positional argument comes next.
		words = words[:len(words)-1]
	}
	start := -1
	for index, word := range words {
		if strings.Trim(word, `"'`) == name {
			start = index + 1
			break
		}
	}
	if start < 0 {
		return nil
	}
	var args []string
	for index := start; index < len(words); index++ {
		word := strings.Trim(words[index], `"'`)
		if valueFlags[word] {
			index++
			continue
		}
		if strings.HasPrefix(word, "-") {
			continue
		}
		args = append(args, word)
	}
	return args
}

func completionMetadata() completion.Command {
	metadata := commandMetadata()
	children := make([]completion.Command, 0, len(metadata.Subcommands)-1)
	for _, child := range metadata.Subcommands {
		if child.Name != "completion" {
			children = append(children, child)
		}
	}
	metadata.Subcommands = children
	return metadata
}

func commandMetadata() completion.Command {
	metadata := cobraMetadata(rootCmd)
	metadata.CompletionCommand = valuesInvocation("active")
	decorateCompletion(&metadata, nil)
	return metadata
}

func decorateCompletion(command *completion.Command, parent []string) {
	path := append(append([]string{}, parent...), command.Name)
	joined := strings.Join(path[1:], " ")
	switch joined {
	case "add":
		command.CompletionCommand = valuesInvocation("add")
	case "archive", "attach", "open", "path", "rename", "status", "switch":
		command.CompletionCommand = valuesInvocation("active")
	case "delete":
		command.CompletionCommand = valuesInvocation("sessions")
	case "init":
		command.CompletionCommand = valuesInvocation("shells")
	case "new":
		command.CompletionCommand = valuesInvocation("new")
	case "remove":
		command.CompletionCommand = valuesInvocation("remove")
	case "unarchive":
		command.CompletionCommand = valuesInvocation("archived")
	}
	for index := range command.Subcommands {
		decorateCompletion(&command.Subcommands[index], path)
	}
}

func valuesInvocation(kind string) []string {
	return []string{"sy", "__values", kind, completion.ContextPlaceholder}
}

func cobraMetadata(command *cobra.Command) completion.Command {
	metadata := completion.Command{
		Name:            command.Name(),
		Description:     command.Short,
		Synopsis:        command.UseLine(),
		LongDescription: command.Long,
	}
	command.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		metadata.Flags = append(metadata.Flags, completion.Flag{
			Name: flag.Name, Short: flag.Shorthand, Description: flag.Usage,
			Value: flag.NoOptDefVal == "",
		})
	})
	for _, child := range command.Commands() {
		if child.Hidden {
			continue
		}
		metadata.Subcommands = append(metadata.Subcommands, cobraMetadata(child))
	}
	return metadata
}

func generateRepository(check bool) error {
	schema, err := config.Schema()
	if err != nil {
		return err
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		return fmt.Errorf("read README.md: %w", err)
	}
	updated, err := completion.ReplaceSection(string(readme), "commands", completion.Markdown(commandMetadata()))
	if err != nil {
		return err
	}
	outputs := map[string][]byte{
		"README.md":                 []byte(updated),
		"schema/config.schema.json": schema,
	}
	for _, shell := range []string{"bash", "fish", "nu", "zsh"} {
		generated, generateErr := generatedCompletion(shell)
		if generateErr != nil {
			return generateErr
		}
		outputs[filepath.Join("completions", "sy."+shell)] = []byte(generated + "\n")
	}
	paths := make([]string, 0, len(outputs))
	for path := range outputs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := writeGenerated(path, outputs[path], check); err != nil {
			return err
		}
	}
	return nil
}

func writeGenerated(path string, content []byte, check bool) error {
	existing, err := os.ReadFile(path)
	if check {
		if err != nil || !bytes.Equal(existing, content) {
			return fmt.Errorf("%s is stale; run sy generate", path)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
