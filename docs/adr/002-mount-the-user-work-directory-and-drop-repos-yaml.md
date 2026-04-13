# ADR-002 - Mount The User Work Directory And Drop `repos.yaml`

## Status
Approved

## Date
2026-04-12

## Context

Harbour currently uses `repos.yaml` in the harness to decide which repo paths to mount into the VM.

The mechanism adds a second curation layer on top of the user's filesystem layout.

`repos.yaml` creates avoidable friction:

- Mount config must be maintained and reflect the workspace directory
- An organisation-managed `repos.yaml` may not match the user's chosen directory layout
- Additional checkouts and worktrees do not work without extra config
- Provisioning and mount diffing are more complex than they need to be

## Decision

Mount the single, user-selected work directory instead of mounting multiple paths from `repos.yaml`.

Provisioning will prompt for:

- `workspace_path`
- `harness_path`

`workspace_path` is the host directory Harbour mounts into the VM.

`harness_path` must be inside `workspace_path`.

`harness_path` must not equal `workspace_path`.

Harbour will stop using `repos.yaml` as a source of truth for mounts.

Harbour will link the selected agent instruction file into the VM user's home directory and link the harness skills directory into the selected agent's skills directory in the VM user's home.

## Consequences

### Benefits

- The user curates scope by choosing a work directory instead of editing a repo list
- Different repo names, parallel checkouts, and worktrees work without extra config
- Provisioning becomes simpler

### Costs

- The mounted scope is broader than an explicit repo allowlist
- The user must choose a sensible `workspace_path`
- Harbour no longer supports an organisation-managed allowlist of mounted repos

## Rejected alternatives

### Keep `repos.yaml`
This keeps an explicit allowlist. It also duplicates information the filesystem already contains and
does not fit local layouts with renamed directories, multiple checkouts, or worktrees.

### Keep `repos.yaml` and also support mounting a work directory
This preserves backwards compatibility at the cost of two competing models. The extra flexibility
does not justify the extra documentation, validation, and debugging complexity.

### Allow `harness_path` outside `workspace_path`
This would require mounting the harness separately or relaxing the single-root model. That makes the path model and
mount logic more complex than necessary.
