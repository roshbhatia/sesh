# Agents: Writing and updating a PRD

Use docs/template/prd.md as the canonical PRD template for projects in this repository.

Instructions for agents and contributors:

- When starting a new project, create a PRD in `docs/` using this template and name it clearly (eg. `docs/my-project-prd.md`).
- If you are an agent drafting the PRD, use this template verbatim and fill in the sections with researched content. Link to interview notes and research for every assumption.
- If the PRD needs revisions, update the PRD file directly and keep a short changelog at the top of the PRD describing what changed and why. Always keep the canonical PRD up to date.
- When revising project environments (for example updating `shell.nix`), note the change in the PRD's changelog if the environment change affects development or deployment.

Recommended workflow for agents:

1. Create `docs/<project>-prd.md` from `docs/template/prd.md`.
2. Fill research and goals sections with links to sources and interview notes.
3. Commit the PRD in a branch and open a PR for human review.
4. Iterate using PR comments; when changes are requested, update the PRD and note revisions in the top-of-file changelog.

Always treat the PRD in `docs/` as the single source of truth for product scope and decisions.
