package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Output shell integration code",
	Long:  `Output shell integration code to be eval'd in your shell configuration.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := "zsh"
		if len(args) > 0 {
			shell = args[0]
		}

		if shell != "zsh" {
			return fmt.Errorf("only zsh is currently supported")
		}

		fmt.Print(getZshIntegration())
		return nil
	},
}

func getZshIntegration() string {
	return `# sesh shell integration for zsh

# s - Quick navigate to session
s() {
    local session_path
    session_path=$(sesh path "$1" 2>/dev/null)
    if [[ -n "$session_path" && -d "$session_path" ]]; then
        cd "$session_path" || return 1
    else
        echo "Session '$1' not found" >&2
        return 1
    fi
}

# si - Interactive session selector
si() {
    local session_path
    session_path=$(sesh select 2>/dev/null)
    if [[ -n "$session_path" && -d "$session_path" ]]; then
        cd "$session_path" || return 1
    fi
}

# Completion for sesh command
_sesh() {
    local -a sessions
    local sessions_root="${XDG_STATE_HOME:-$HOME/.local/state}/sesh/sessions"
    
    if [[ -d "$sessions_root" ]]; then
        sessions=(${(f)"$(ls -1 "$sessions_root" 2>/dev/null)"})
    fi

    local -a commands
    commands=(
        'new:Create a new session'
        'list:List all sessions'
        'delete:Delete a session'
        'path:Print session path'
        'select:Interactively select a session'
        'init:Output shell integration code'
        'version:Show version'
        'help:Show help'
    )

    if (( CURRENT == 2 )); then
        _describe -t commands 'sesh commands' commands
        _describe -t sessions 'sessions' sessions
    elif (( CURRENT == 3 )); then
        case "$words[2]" in
            delete|rm|remove|path)
                _describe -t sessions 'sessions' sessions
                ;;
        esac
    fi
}

compdef _sesh sesh

# Completion for s function
_s() {
    local sessions_root="${XDG_STATE_HOME:-$HOME/.local/state}/sesh/sessions"
    if [[ -d "$sessions_root" ]]; then
        _values 'sessions' ${(f)"$(ls -1 "$sessions_root" 2>/dev/null)"}
    fi
}

compdef _s s
`
}

func init() {
	rootCmd.AddCommand(initCmd)
}
