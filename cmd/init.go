package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const bashInit = `# sesh shell integration
s() {
  local dir
  dir="$(sesh path "$1" 2>/dev/null)" || dir="$(sesh --greedy "$1" 2>/dev/null)"
  if [ -n "$dir" ]; then
    cd "$dir" || return 1
  else
    echo "sesh: session '$1' not found" >&2
    return 1
  fi
}

si() {
  local dir
  dir="$(sesh select 2>/dev/tty)" || return 1
  cd "$dir" || return 1
}
`

const zshInit = `# sesh shell integration
s() {
  local dir
  dir="$(sesh path "$1" 2>/dev/null)" || dir="$(sesh --greedy "$1" 2>/dev/null)"
  if [[ -n "$dir" ]]; then
    cd "$dir" || return 1
  else
    echo "sesh: session '$1' not found" >&2
    return 1
  fi
}

si() {
  local dir
  dir="$(sesh select 2>/dev/tty)" || return 1
  cd "$dir" || return 1
}
`

const fishInit = `# sesh shell integration
function s
  set -l dir (sesh path $argv[1] 2>/dev/null; or sesh --greedy $argv[1] 2>/dev/null)
  if test -n "$dir"
    cd $dir
  else
    echo "sesh: session '$argv[1]' not found" >&2
    return 1
  end
end

function si
  set -l dir (sesh select 2>/dev/tty)
  or return 1
  cd $dir
end
`

var initCmd = &cobra.Command{
	Use:   "init <bash|zsh|fish>",
	Short: "Print shell integration snippets",
	Long: `Output shell function definitions for sesh integration.

Add to your shell config:
  # bash (~/.bashrc)
  eval "$(sesh init bash)"

  # zsh (~/.zshrc)
  eval "$(sesh init zsh)"

  # fish (~/.config/fish/config.fish)
  sesh init fish | source`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			fmt.Print(bashInit)
		case "zsh":
			fmt.Print(zshInit)
		case "fish":
			fmt.Print(fishInit)
		default:
			return fmt.Errorf("unsupported shell %q: supported shells are bash, zsh, fish", args[0])
		}
		return nil
	},
	ValidArgs: []string{"bash", "zsh", "fish"},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
