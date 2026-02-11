# sesh

Multi-repo session manager with powerful git integration for developers who work across multiple repositories.

## What is sesh?

`sesh` creates isolated workspaces (sessions) with symlinks to your frequently-used repositories. Each session is a directory containing:
- Symlinks to git repos organized by `org/repo-name`
- A pre-configured `shell.nix` with common development stacks
- A `.envrc` file for direnv integration
- Session metadata (description)
- **Git operations across all repos simultaneously**
- **Transparent worktree support for parallel feature work**

Perfect for when you need to work across multiple repos with synchronized branches and unified git operations.

## Installation

### Prerequisites

- [zoxide](https://github.com/ajeetdsouza/zoxide) - For intelligent directory tracking
- [fzf](https://github.com/junegunn/fzf) - For fuzzy finding and multi-select
- [gum](https://github.com/charmbracelet/gum) - For beautiful terminal UI
- [direnv](https://direnv.net/) - (Optional) For automatic environment loading
- [git](https://git-scm.com/) - For git integration features
- [bats-core](https://github.com/bats-core/bats-core) - (Optional) For running tests

### Install sesh

```bash
# Using task (recommended - installs binary + zsh completion)
task install

# Or manually
cp sesh ~/bin/sesh
chmod +x ~/bin/sesh

# Install zsh completion (optional)
mkdir -p ~/.local/share/zsh/site-functions
cp _sesh ~/.local/share/zsh/site-functions/_sesh

# Make sure ~/bin is in your PATH
export PATH="$HOME/bin:$PATH"

# For completions, ensure your .zshrc has (if not already present):
fpath=(~/.local/share/zsh/site-functions $fpath)
autoload -Uz compinit && compinit
```

After installation, restart your shell or run `exec zsh` for completions to take effect.

## Shell Integration

Add this to your `.zshrc` for the best experience:

```bash
eval "$(sesh init zsh)"
```

This provides:
- `s <session>` - Quick navigate to any session
- `si` - Interactive session selector with preview
- Tab completion for all commands and session names

## Usage

### Interactive Mode

Just run `sesh` to get an interactive menu:

```bash
sesh
```

### Create a new session

```bash
# Interactive creation
sesh new my-project

# With description
sesh new my-project "Platform composition v2 work"
```

This will:
1. Prompt for a description (if not provided)
2. Show a filterable list of directories from zoxide
3. Let you select multiple repos (Space to select, Enter to confirm)
4. Create symlinks organized by `org-name/repo-name`

### List sessions

```bash
sesh list
```

Shows all sessions with descriptions and repo counts.

### Access a session

```bash
# Print session path
sesh my-project

# Change into session directory
cd "$(sesh my-project)"

# Or use in scripts
SESSION_PATH=$(sesh my-project)
```

### Manage repos in a session

```bash
# Add more repos to existing session
sesh add my-project

# Remove repos from session
sesh remove my-project
```

### Manage sessions

```bash
# Rename a session
sesh rename old-name new-name

# Update session description
sesh describe my-project "Updated description here"

# Show current description
sesh describe my-project

# Delete a session
sesh delete my-project
```

## Git Integration

Work with git across all repos in a session simultaneously.

### Check status across all repos

```bash
sesh status my-project

# Shows:
# - Current branch for each repo
# - Modified/staged/untracked files
# - Commits ahead/behind remote
```

### Sync all repos

```bash
# Pull all repos (skips repos with uncommitted changes)
sesh sync my-project

# Fetch only (no pull)
sesh sync my-project --fetch
```

### Branch management

```bash
# Create/switch to branch in all repos
sesh branch my-project feature-x

# Show current branches
sesh branch my-project --status

# Create isolated worktree session (for parallel feature work)
sesh branch my-project feature-y --worktree
# This creates "my-project-feature-y" session with git worktrees
# Work on feature-y without affecting your main workspace
```

### Execute commands across all repos

```bash
# Run any command in all repos
sesh exec my-project git fetch
sesh exec my-project git log -1
sesh exec my-project npm install

# Commands run in each repo directory
# Exit codes are aggregated
```

## Directory Structure

Sessions are stored in `$XDG_STATE_HOME/sesh/` (defaults to `~/.local/state/sesh/`):

```
~/.local/state/sesh/
├── my-project/
│   ├── .sesh-desc              # Session description
│   ├── .envrc                  # Direnv config (use nix)
│   ├── shell.nix               # Nix shell environment
│   ├── nike-runtime-foundation/
│   │   ├── composition-runtime@ -> /path/to/repo
│   │   └── provider-metadata@ -> /path/to/repo
│   └── rbha18_nike/
│       └── sysinit@ -> /path/to/repo
```

## Examples

### Multi-repo development workflow

```bash
# Create a session for platform work
sesh new platform "NRF platform composition work"
# Select: composition-runtime, provider-metadata, nrf-core-components

# Access the session
cd "$(sesh platform)"
# Or with shell integration: s platform

# Open in editor with all repos accessible
code .

# Check git status across all repos
sesh status platform

# Create feature branch in all repos
sesh branch platform feature-auth-improvements

# Work on your changes...

# Sync all repos
sesh sync platform

# Need to experiment? Create a worktree session
sesh branch platform experimental-refactor --worktree
s platform-experimental-refactor
# Work independently, delete when done

# Clean up
sesh delete platform-experimental-refactor  # Cleans up worktrees automatically
```

### Parallel feature development

```bash
# Main work in original session
cd "$(sesh platform)"
# On feature-a branch

# Need to work on feature-b without stashing?
sesh branch platform feature-b --worktree

# Now you have two independent sessions:
s platform            # Working on feature-a
s platform-feature-b  # Working on feature-b

# Both sessions share git history but have independent working trees
```

### Shell integration

Add to your `.zshrc` or `.bashrc` for quick access:

```bash
# Already provided by: eval "$(sesh init zsh)"
# This gives you:
# - s <session>  : Quick cd to session
# - si           : Interactive selector
# - Tab completion everywhere
```

Example usage with shell integration:

```bash
# Navigate to session
s my-project

# Interactive picker
si

# All sesh commands have completion
sesh branch <TAB>     # Shows session names
sesh status <TAB>     # Shows session names
```

### Use with tmux/zellij

```bash
# Open tmux session in sesh workspace
tmux new-session -s my-project -c "$(sesh my-project)"

# Or with shell integration
tmux new-session -s my-project -c "$(s my-project)"
```

## Development

```bash
# Check dependencies
task check

# Run tests (requires bats-core)
task test

# Install bats on macOS
brew install bats-core

# Install locally
task install

# Clean up test sessions
task clean
```

## Why sesh?

- **Focused workspaces**: Group related repos without cluttering your main directories
- **Fast navigation**: Jump to multi-repo workspaces instantly with `s` command
- **Git integration**: Manage branches, sync repos, check status across all repos at once
- **Worktree support**: Work on multiple branches in parallel without stashing
- **Nix integration**: Each session has pre-configured shell.nix with common stacks
- **Organized**: Repos are symlinked with `org/repo` structure for clarity
- **zoxide-powered**: Leverages your existing directory history
- **Beautiful UI**: Uses gum for polished terminal interactions
- **Shell integration**: Natural `s` and `si` commands with completion

## Tips

1. **Use with direnv**: Run `direnv allow` in your session directory to auto-load the Nix environment
2. **Customize shell.nix**: Edit the generated shell.nix and uncomment the stacks you need (Go, Node, Python, Rust, etc.)
3. **Session naming**: Use descriptive names like `platform-v2`, `experiments`, `client-work`
4. **Worktrees for experiments**: Use `--worktree` flag when you want to experiment without affecting your main work
5. **Regular cleanup**: Delete worktree sessions when done to keep things tidy
6. **Git workflows**: Use `sesh branch` to synchronize branch changes across all repos in a session

## License

MIT
