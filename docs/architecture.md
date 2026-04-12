# Architecture

## Goals

- Keep the host-side CLI small and maintainable
- Keep the runtime model VM-first
- Keep cross-repo access explicit through the harness
- Keep the host-side config model boring

## Host-Side Shape

`harbour` is a single Go binary.

It owns:

- Command dispatch
- Config load and save
- Interactive prompting
- Mount calculation
- Colima start, stop, and ssh orchestration

The host-side config is one JSON file at the Harbour config path returned by
`os.UserConfigDir()`.

## Harness Split

Recommended host-side split:

- `harbour`
  The shareable CLI repo
- `harbour-harness`
  Private local state such as `AGENTS.md` and `skills/`

Harbour mounts the configured work directory and expects the harness to live
inside it.

## VM Runtime

Harbour keeps the VM setup model intentionally simple:

- The configured work directory is mounted once
- The harness must live inside that work directory
- The selected root instruction file is linked at the workspace path
- Custom skills are symlinked into the active agent directory
- The active agent runs directly inside the VM shell

The host-side CLI is Go. The in-VM provision step is still an embedded Bash
script because that boundary is already shell-shaped and mostly imperative
machine setup.

## Path Model

Mounted repos keep the same absolute paths inside the VM as on the host.

This avoids translation logic and keeps logs, diagnostics, and tool output
consistent across host and VM. See [ADR-001](adr/001-mirror-host-repo-paths-inside-the-vm.md).
