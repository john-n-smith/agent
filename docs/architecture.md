# Architecture

## Goals

- One entry point for day-to-day work.
- Cross-repo visibility by default.
- VM-first execution.
- A clean split between the shareable harness and private personal state.

## Model

The `Harbour` repo acts as the shareable harness. It owns the launch scripts,
behaviour rules, and harness design records.

Personal working state should live in a separate private repo such as
`harbour-context`. That repo should hold `AGENTS.md`, `repos.yaml`, and any
other private local files.

Inside the VM, the master agent can see the mounted host repo paths declared in
`harbour-context/repos.yaml`, plus the workspace root declared in the local
Harbour env.

The master agent keeps global awareness across repos. When a task needs deeper
project-specific work, it reads that repo's local instructions and narrows focus
there, but it does not lose cross-repo context.

The agent should run directly in the VM shell, not in its own container. Repo
containers also run in the same VM, which avoids a nested runtime shape.

## Repo Split

Recommended host-side split:

- `harbour`
  Shareable harness repo
  Holds `Makefile`, `config/`, `scripts/`, `docs/`, and harness ADRs
- `harbour-context`
  Private state repo
  Holds `AGENTS.md`, `repos.yaml`, and any other private local files

Recommended VM exposure:

- Mount work repos from the host
- Mount the sibling `harbour-context` repo from the host by convention
- Link the selected root instruction file at the workspace root during provision
- Do not mount the whole harness repo into the VM by default unless a real need appears

## VM Runtime

The startup scripts are deliberately thin wrappers:

- `harbour-context/repos.yaml` defines allowed host-to-VM mounts
- Each entry in `harbour-context/repos.yaml` is mounted read-write
- Absolute repo paths are mounted as written
- Relative repo paths are resolved from `HARBOUR_WORKSPACE_ROOT`
- Missing repo mount directories are warned and skipped during provision
- `config/colima.env` defines the Colima profile and VM defaults
- `~/.config/agent-harbour/env` defines machine-local Harbour env values such as `HARBOUR_CONTEXT_HOST_PATH`, `HARBOUR_WORKSPACE_ROOT`, and `HARBOUR_ACTIVE_AGENT`
- `scripts/provision` starts the VM if needed, prompts before restarting when mount config drifts, prompts for the active agent, installs only that agent in the VM, removes the inactive agent, links the matching workspace instruction file, and syncs skills to the selected agent's skills directory
- `scripts/agent` launches the provisioned agent in the VM

This keeps shared VM defaults in config, private runtime state in
`harbour-context`, and `make` as the stable entry point.

Isolation comes from the VM boundary, but any path mounted from the host into
the VM is intentionally shared. Narrow mounts are therefore part of the safety
model, not just convenience.

Mirror mounted repo paths inside the VM rather than introducing a separate
shared root. See [ADR-001](adr/001-mirror-host-repo-paths-inside-the-vm.md).
