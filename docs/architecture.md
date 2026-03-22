# Architecture

## Goals

- One entry point for day-to-day work.
- Cross-repo visibility by default.
- VM-first execution.
- Durable memory and prioritisation outside individual project repos.
- A clean split between the shareable harness and private personal state.

## Model

The `agent` repo acts as the shareable harness. It owns the launch scripts,
behaviour rules, and harness design records.

Personal working state should live in a separate private repo such as
`agent-context`. That repo should hold live `todo.md`, `notes.md`,
`repos.yaml`, and `runtime.env`.

Inside the VM, the master agent can see the mounted host repo paths declared in
`agent-context/repos.yaml`, plus the runtime paths declared in
`agent-context/runtime.env`.

The master agent keeps global awareness across repos. When a task needs deeper
project-specific work, it reads that repo's local instructions and narrows focus
there, but it does not lose cross-repo context.

The agent should run directly in the VM shell, not in its own container. Repo
containers also run in the same VM, which avoids a nested runtime shape.

## Repo Split

Recommended host-side split:

- `agent`
  Shareable harness repo
  Holds `Makefile`, `config/`, `scripts/`, `docs/`, and harness ADRs
- `agent-context`
  Private state repo
  Holds `todo.md`, `notes.md`, `repos.yaml`, `runtime.env`, and other personal working state

Recommended VM exposure:

- Mount work repos from the host
- Mount the sibling `agent-context` repo from the host by convention
- Link `AGENTS.md` at the workspace root to `agent-context/AGENTS.md` during provision
- Do not mount the whole harness repo into the VM by default unless a real need appears

## Handoff

Use this lightweight handoff shape when narrowing into a specific repo:

```text
repo: org-infra
goal: investigate IAM Identity Center changes
global context:
- app rollout depends on account bootstrap timing
- current priority is infra unblockers before product work
repo-local files to read first:
- AGENTS.md
- README.md
- TODO.md
done when:
- root cause identified
- change made or blocker recorded
```

## VM Runtime

The startup scripts are deliberately thin wrappers:

- `agent-context/repos.yaml` defines allowed host-to-VM mounts
- `config/colima.env` defines the Colima profile and VM defaults
- `agent-context/runtime.env` defines local runtime paths such as `WORKSPACE_ROOT`
- `scripts/destroy` stops the VM profile
- `scripts/provision` starts the VM if needed, prompts before restarting when mount config drifts, updates Codex in the VM to the configured version, links the workspace instruction file, and syncs Codex skills
- `scripts/agent` launches Codex in the VM with `workspace-write`

This keeps shared VM defaults in config, private runtime state in
`agent-context`, and `make` as the stable entry point.

Isolation comes from the VM boundary, but any path mounted from the host into
the VM is intentionally shared. Narrow mounts are therefore part of the safety
model, not just convenience.

Mirror mounted repo paths inside the VM rather than introducing a separate
shared root. See [ADR-001](adr/001-mirror-host-repo-paths-inside-the-vm.md).
