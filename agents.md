# Agents guidance for this session

- Use Nix when possible. This session ships with a `shell.nix`; prefer adding any session-specific tools to that file so the environment is reproducible for everyone.
- After changing `shell.nix` or `.envrc`, run `direnv allow` in the session directory to reload the environment.
- When adding tools, add a short comment in `shell.nix` explaining why the tool is required.

Session layout and where things live

- `.sesh-desc` — human-readable session description
- `.envrc` — direnv wrapper that should contain `use nix` (do not remove without updating the PRD)
- `shell.nix` — nix environment for this session
- `docs/` — documentation for this session (canonical PRD lives here)
- symlinked repos are organized under `org-name/repo-name/`

PRD guidance (session-specific)

- The session PRD is: `docs/provider-nrf-rate-limiting-gc-prd.md` — treat this file as the single source of truth for product scope and decisions for work done in this session.
- If you create or revise a PRD, add a top-of-file changelog entry (date, author/agent, short note) and ensure research links are included for assumptions.
- If you update development or environment files (for example `shell.nix`), add an entry to the PRD changelog describing the environment change and why it matters.

Agent workflow (quick)

1. Read `docs/provider-nrf-rate-limiting-gc-prd.md` and add your changes as a new draft or revision.
2. Update `shell.nix` only if you need new tools; document the reason in `shell.nix` and in the PRD changelog.
3. Run `direnv allow` to activate changes locally.

Notes

- This `agents.md` is local to the session and is not part of the repository. If you want the guidance copied back into the repo, copy it into `docs/` at the repo root and commit.
- Keep the PRD up to date; agents should assume the PRD in `docs/` here is authoritative for session work.
