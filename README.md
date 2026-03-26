# seshy

Minimalist session manager for multi-repo development, with git worktree integration.

## Installation

```bash
git clone https://github.com/roshbhatia/seshy
cd seshy
task install
```

## Commands

| Command | Description |
|---------|-------------|
| `sy new <name>` | Create a new session, selecting repos from zoxide history |
| `sy add <name>` | Add repositories to an existing session |
| `sy list` / `sy ls` | List all sessions |
| `sy delete <name>` / `sy rm <name>` | Delete a session and clean up worktrees + branches |
| `sy path <name>` | Print session path |
| `sy config` | Show effective configuration |
| `sy config edit` | Open config file in editor |
| `sy --greedy <query>` | Fuzzy match session and print its path |

## Configuration

Config file at `~/.config/seshy/config.yaml` (or `$XDG_CONFIG_HOME/seshy/config.yaml`):

```yaml
# Branch naming template. Variables: {{.Session}}, {{.Repo}}, {{.User}}
branchFormat: "sy/{{.Session}}/{{.Repo}}"

# Sessions storage directory
sessionsDir: "~/.local/state/seshy/sessions"
```

Run `sy config` to see effective settings, `sy config edit` to modify.

## Branch Naming

By default, worktree branches are named `sy/<session>/<repo>`. Override per-invocation:

```bash
sy new my-feature --branch hotfix/urgent
sy add my-feature -b feature/custom-branch
```

Or set a custom template in config:

```yaml
branchFormat: "dev/{{.User}}/{{.Session}}/{{.Repo}}"
```
