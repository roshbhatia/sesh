# sesh

Minimalist session manager for developers working across multiple repositories, with git worktree integration.

## Installation

```bash
git clone https://github.com/roshbhatia/sesh
cd sesh
task install
```

## Commands

| Command | Description |
|---------|-------------|
| `sesh new <name>` | Create a new session, selecting repos from zoxide history |
| `sesh add <name>` | Add repositories to an existing session |
| `sesh list` / `sesh ls` | List all sessions |
| `sesh delete <name>` / `sesh rm <name>` | Delete a session and clean up worktrees |
| `sesh path <name>` | Print session path |
| `sesh select` | Interactive session picker (outputs path) |
| `sesh --greedy <query>` | Fuzzy match session and print its path |
| `s <name>` | Jump to session (shell function) |
| `si` | Interactive session picker with cd (shell function) |

## Recipies

```bash
function s() {
local path
path=$(sesh --greedy "$1") || return 1
cd "$path"
}

function si() {
local path
path=$(sesh) || return 1
cd "$path"
}
```
