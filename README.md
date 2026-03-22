# agent

Shareable harness for a Colima-backed, cross-repo coding agent.

This repo is the entry point for day-to-day work. It keeps:

- The harness configuration and launch scripts
- The harness ADRs and architecture
- Scripts and `make` targets to launch the VM workspace

Over time, durable personal working state should move to a separate private repo
such as `agent-context`.

That private repo now holds:

- `todo.md`
- `notes.md`
- `repos.yaml`
- `runtime.env`

## Layout

- `Makefile`: stable operator commands
- `config/colima.env`: Colima runtime defaults
- `scripts/`: implementation for `make` targets
- `docs/architecture.md`: operating model and handoff design
- `docs/adr/`: harness decisions

## Usage

```sh
make help
make provision
make agent
```

## Bootstrap

Create `agent-context/AGENTS.md`, `agent-context/repos.yaml`, and
`agent-context/runtime.env` from your skeleton repo structure.

`make provision` requires `agent-context/AGENTS.md`,
`agent-context/repos.yaml`, and `agent-context/runtime.env`. It starts the
Colima VM if needed, installs Codex and GitHub CLI in the VM, and links
`AGENTS.md` at `WORKSPACE_ROOT` to the private instruction file. If the
configured mount set differs from the running Colima profile, it prompts before
restarting Colima to apply the change. It also syncs custom skills into
`~/.codex/skills/`. Set `CODEX_VERSION=latest` in
`config/colima.env` to update to the newest Codex release on each run, or pin a
specific version such as `0.114.0` to hold it steady.

## Notes

The intended runtime is a Colima VM. The agent process should run directly in
the VM, alongside repo containers, rather than inside its own container.

Keep host isolation meaningful by mounting only the repo paths declared in
`agent-context/repos.yaml`. Anything mounted into the VM is intentionally shared with
the agent and repo containers.

`make agent` launches Codex with `workspace-write` so the mounted repos and
`agent-context` are writable.
