# sesh

A streamlined session manager for developers who work across multiple repositories. Built in Go with focus on simplicity and git worktree integration.

## What is sesh?

`sesh` creates isolated development workspaces (sessions) with:
- **Git worktrees** for parallel feature development
- **shell.nix** for reproducible environments
- **Flat structure** - no complex nesting
- **Fast navigation** - mirrors zoxide's `z`/`zi` with `s`/`si`

Perfect for when you need clean, isolated workspaces without the complexity.

## Installation

### Prerequisites

- [Go](https://golang.org/) 1.21+ (for building from source)
- [git](https://git-scm.com/) - Version control
- [direnv](https://direnv.net/) - (Optional) Auto-load environments

### Install from source

```bash
git clone https://github.com/roshbhatia/sesh
cd sesh

# Build and install
task install

# Or manually
go build -o ~/bin/sesh .
chmod +x ~/bin/sesh

# Make sure ~/bin is in your PATH
export PATH="$HOME/bin:$PATH"
```

## Shell Integration

Add these functions to your `.zshrc` (or `.bashrc`):

```bash
# s <session> — jump directly to a session by name (greedy match)
function s() {
  local path
  path=$(sesh --greedy "$1") || return 1
  cd "$path"
}

# si — interactive session picker
function si() {
  local path
  path=$(sesh) || return 1
  cd "$path"
}
```

After adding, restart your shell:

```bash
exec zsh
```

### Usage

```bash
# Jump to a session (greedy: exact, prefix, or substring match)
s platform

# Interactive picker
si
```

## Quick Start

```bash
# Create a new session
sesh new platform-work

# Navigate to it (interactive picker)
si

# Or jump directly by name
s platform-work

# Later, add more repos to the session
sesh add platform-work

# List all sessions
sesh list

# Delete when done
sesh delete platform-work
```

## Usage

### Create a session

```bash
sesh new my-project
```

Select repositories from your zoxide history using the interactive picker (Space to toggle, Enter to confirm).

**What gets created:**
```
~/.local/state/sesh/sessions/my-project/
├── shell.nix
├── composition-runtime-my-project/  # Git worktree
└── provider-metadata-my-project/    # Git worktree
```

### Navigate to sessions

```bash
# Interactive selection
si

# Direct navigation (greedy match)
s my-project

# Manual with path
cd $(sesh path my-project)
```

### List sessions

```bash
sesh list
# or
sesh ls
```

Shows all sessions with repo counts and last modified times.

### Delete a session

```bash
sesh delete my-project
# or
sesh rm my-project
```

Automatically cleans up all git worktrees.

## Session Structure

Sessions use a simplified flat structure:

```
~/.local/state/sesh/sessions/
└── platform-v2/
    ├── shell.nix                          # Nix environment
    ├── composition-runtime-platform-v2/   # Git worktree
    └── provider-metadata-platform-v2/     # Git worktree
```

**Key design choices:**
- Session name is the description (no separate metadata)
- Worktrees named `<repo-basename>-<session-name>`
- Only `shell.nix` for configuration
- No org/repo nesting - flat and simple

## Git Worktrees

For git repositories, sesh creates worktrees instead of symlinks. This means:

✓ **Parallel development** - Work on multiple branches simultaneously  
✓ **Independent state** - Each session has its own working directory  
✓ **Shared history** - All worktrees share the same git history  
✓ **Clean isolation** - No stashing or branch switching needed

**Example workflow:**

```bash
# Main work
sesh new platform-main
s platform-main
# Work on main/master branch

# Experiment in parallel
sesh new platform-experiment
s platform-experiment
# Work on experimental features

# Both sessions are independent
# Delete experiment when done
sesh rm platform-experiment
```

## Environment Variables

### `XDG_STATE_HOME`

Override default state directory:

```bash
export XDG_STATE_HOME="$HOME/.local/state"  # Default
```

Sessions stored in: `$XDG_STATE_HOME/sesh/sessions/`

## Nix Integration

Each session includes a `shell.nix` template:

```nix
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    git
    # Add your development tools:
    # nodejs
    # go
    # python3
  ];
  
  shellHook = ''
    echo "Session: $(basename $PWD)"
  '';
}
```

**Usage with direnv:**

```bash
cd $(sesh path my-session)
echo "use nix" > .envrc
direnv allow
# Environment auto-loads when you cd into session
```

## Commands

| Command | Description |
|---------|-------------|
| `sesh new <name>` | Create a new session |
| `sesh add <name>` | Add repositories to existing session |
| `sesh list` or `sesh ls` | List all sessions |
| `sesh delete <name>` | Delete a session |
| `sesh path <name>` | Print session path |
| `sesh select` | Interactive session picker (outputs path) |
| `sesh --greedy <query>` | Fuzzy match session and print its path |
| `sesh --version` | Show version |
| `sesh --help` | Show help |

## Development

### Build from source

```bash
# Run tests
task test

# Build binary
task build

# Install locally
task install

# Development cycle
task dev

# Clean build artifacts
task clean

# Check dependencies
task check
```

### Run tests

```bash
# All tests with verbose output
task test

# Quick test run
task test-short

# Coverage report
task test-coverage
```

## Migration from v2.x (bash version)

sesh v3.0 is a complete rewrite with breaking changes:

**What changed:**
- Written in Go (was bash)
- Simplified session structure (no org/repo nesting)
- Removed: `describe`, `rename`, `add`, `remove`, `branch`, `status`, `sync`, `exec`
- Worktrees named `<repo>-<session>` (not nested under org/)
- Session name is the description (no `.sesh-desc` file)

**Migration steps:**
1. Sessions are not compatible - you'll need to recreate them
2. Update any scripts/aliases that depended on v2 behavior
3. Install the new binary: `task install`

**Philosophy:**
v3 focuses on the core workflow: create sessions, manage worktrees, jump between them. Everything else is left to standard git commands.

## Examples

### Multi-repo project

```bash
# Create session for platform work
sesh new platform-v2
# Select: composition-runtime, provider-metadata

# Navigate
s platform-v2

# Later, add another repo
sesh add platform-v2
# Select: nrf-core

# Each repo is a worktree named:
# - composition-runtime-platform-v2/
# - provider-metadata-platform-v2/
# - nrf-core-platform-v2/
```

### Parallel feature development

```bash
# Main feature work
sesh new auth-v1
s auth-v1

# Need to prototype something else?
sesh new auth-v2-experiment
s auth-v2-experiment

# Both sessions are independent
# Delete experiment when done
sesh delete auth-v2-experiment
```

### With tmux

```bash
# Create tmux session in sesh workspace
tmux new-session -s platform -c "$(sesh path platform-v2)"
```

## Why the rewrite?

**v2.x problems:**
- 1300 lines of bash
- Complex org/repo nesting
- Too many commands (describe, rename, add, remove, etc.)
- Multi-repo git operations were overkill

**v3.0 benefits:**
- ~500 lines of Go (more maintainable)
- Flat, simple structure
- Focused on core workflow
- Fast binary (no bash subshells)
- Better error handling
- Comprehensive test suite

## Comparison with other tools

| Feature | sesh | tmux | zoxide | tmux-sessionizer |
|---------|------|------|--------|------------------|
| Git worktrees | ✓ | - | - | - |
| Session management | ✓ | ✓ | - | ✓ |
| Nix integration | ✓ | - | - | - |
| Fast directory jumping | ✓ | - | ✓ | - |
| Multi-repo workflows | ✓ | - | - | - |

sesh complements tmux/zellij (not a replacement) - use them together!

## Tips

1. **Use descriptive session names**: `platform-auth-v2` better than `temp-work`
2. **Leverage direnv**: Auto-load Nix environments with `direnv allow`
3. **Clean up experiments**: Delete experimental sessions when done
4. **Combine with tmux**: Create tmux sessions inside sesh sessions
5. **One session per feature**: Keep sessions focused on specific work

## License

MIT

## Credits

Inspired by:
- [zoxide](https://github.com/ajeetdsouza/zoxide) - For the elegant UX patterns
- [tmux-sessionizer](https://github.com/joshmedeski/t-smart-tmux-session-manager) - For session management ideas
