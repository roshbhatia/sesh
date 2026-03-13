# sesh

Minimalist session manager for developers working across multiple repositories, with git worktree integration.

## Installation

```bash
git clone https://github.com/roshbhatia/sesh
cd sesh
task install
```

## Shell Integration

Add to your shell config for `s` (jump to session) and `si` (interactive picker) functions:

```bash
# bash (~/.bashrc)
eval "$(sesh init bash)"

# zsh (~/.zshrc)
eval "$(sesh init zsh)"

# fish (~/.config/fish/config.fish)
sesh init fish | source
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
| `sesh init <shell>` | Print shell integration snippets |
| `sesh --greedy <query>` | Fuzzy match session and print its path |
| `s <name>` | Jump to session (shell function, via `sesh init`) |
| `si` | Interactive session picker with cd (shell function, via `sesh init`) |
