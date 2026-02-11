# sesh

Multi-repo session manager for developers who work across multiple repositories.

## What is sesh?

`sesh` creates isolated workspaces (sessions) with symlinks to your frequently-used repositories. Each session is a directory containing:
- Symlinks to git repos organized by `org/repo-name`
- A basic `shell.nix` for Nix environments
- A `.envrc` file for direnv integration
- Session metadata (description)

Perfect for when you need to work across multiple repos but don't want to navigate through your entire `~/github` directory.

## Installation

### Prerequisites

- [zoxide](https://github.com/ajeetdsouza/zoxide) - For intelligent directory tracking
- [gum](https://github.com/charmbracelet/gum) - For beautiful terminal UI
- [direnv](https://direnv.net/) - (Optional) For automatic environment loading

### Install sesh

```bash
# Using task
task install

# Or manually
cp sesh ~/bin/sesh
chmod +x ~/bin/sesh

# Make sure ~/bin is in your PATH
export PATH="$HOME/bin:$PATH"
```

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

# Delete a session
sesh delete my-project
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

# Open in editor with all repos accessible
code .

# Add more repos later
sesh add platform

# Clean up when done
sesh delete platform
```

### Shell integration

Add to your `.zshrc` or `.bashrc` for quick access:

```bash
# Quick session switcher
s() {
    local session_path=$(sesh "$1")
    if [[ -n "$session_path" ]]; then
        cd "$session_path"
    fi
}

# Usage: s my-project
```

Or use with tmux/zellij:

```bash
# Open tmux session in sesh workspace
tmux new-session -s my-project -c "$(sesh my-project)"
```

## Development

```bash
# Check dependencies
task check

# Run tests
task test

# Install locally
task install

# Clean up test sessions
task clean
```

## Why sesh?

- **Focused workspaces**: Group related repos without cluttering your main directories
- **Fast navigation**: Jump to multi-repo workspaces instantly
- **Nix integration**: Each session has its own shell.nix for project-specific tools
- **Organized**: Repos are symlinked with `org/repo` structure for clarity
- **zoxide-powered**: Leverages your existing directory history
- **Beautiful UI**: Uses gum for polished terminal interactions

## Tips

1. **Use with direnv**: Run `direnv allow` in your session directory to auto-load the Nix environment
2. **Customize shell.nix**: Edit the generated shell.nix to add project-specific tools
3. **Session naming**: Use descriptive names like `platform-v2`, `experiments`, `client-work`
4. **Regular cleanup**: Delete sessions when projects are done to keep things tidy

## License

MIT
