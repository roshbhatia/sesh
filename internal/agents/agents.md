# Agents guidance for this session

- Use Nix when possible. This session ships with a `shell.nix`; prefer adding any session-specific tools to that file so the environment is reproducible for everyone.
- After changing `shell.nix` or `.envrc`, run `direnv allow` in the session directory to reload the environment.
- When adding tools, add a short comment in `shell.nix` explaining why the tool is required.

Session layout and where things live

- `.envrc` — direnv wrapper that should contain `use nix` 
- `shell.nix` — nix environment for this session
- `docs/` — documentation for this session
- Git worktrees organized as `<repo-name>-<session-name>/`

Agent workflow

1. Update `shell.nix` only if you need new tools; document the reason in comments
2. Run `direnv allow` to activate changes locally
3. Keep `docs/` up to date with any session-specific notes
