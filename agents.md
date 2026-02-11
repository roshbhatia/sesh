# Agents guidelines for sesh

Use Nix when possible. When a session needs additional tools, update the generated `shell.nix` inside the session directory so the environment is reproducible and documented.

Key points for agents working in this repository:

- Use `shell.nix` and direnv (`.envrc`) for reproducible shells; prefer adding dependencies to `shell.nix` instead of instructing developers to install global packages.
- If you add or change tools, update `shell.nix` and include a short comment explaining why the tool is required.
- After modifying `shell.nix`, remind maintainers/consumers to run `direnv allow` in the session directory.

About the repos that live in a session

This project creates sessions which are directories containing symlinks to one or more git repositories organized by `org/repo-name`. A session typically contains:

- `.sesh-desc` - human-readable session description
- `.envrc` - direnv wrapper that loads the `shell.nix`
- `shell.nix` - nix environment for the session with project tools
- one or more symlinked repositories under `org-name/repo-name/`

When writing code that manipulates session contents (creating sessions, adding symlinks, generating `shell.nix`), prefer safe, idempotent changes and document expected file locations in `README.md` or this `agents.md`.

If you are an agent (or automation) modifying session contents, leave a short commit message describing the environment change (eg. "update shell.nix: add jq, git-delta").
