# Harbour

Run agents across your repositories in a controlled, shareable environment.

Like Docker Compose, but for agent harnesses.

- Run agents in a sandbox (Colima VM)
- Work across multiple repositories by default
- Define and share your harness (`repos.yaml`, `AGENTS.md`, `skills/`)
- Keep your existing Docker workflow (via docker context)
- Supports Claude or Codex

## Install

```sh
brew tap agent-harbour/harbour
brew install agent-harbour/harbour/harbour
harbour help
```

## Quick Start

1. Create your harness

   - `repos.yaml` lists repo mount paths
   - `AGENTS.md` contains shared instructions
   - `skills/` contains optional custom skills

   See https://github.com/agent-harbour/harbour-harness-template for an example.

   Relative `host_path` values in `repos.yaml` are resolved from `workspace_root`.

2. Provision Harbour

   ```sh
   harbour provision
   ```

   The first run creates Harbour's local config automatically.

   Provision prompts for:

   - Path to your harness
   - Workspace root (where you repos live)
   - Agent to provision
   - The default `harbour` command

3. Run the agent

```sh
harbour
```

Or run a command explicitly:

```sh
harbour agent
harbour yolo
harbour shell
```

## Commands

```sh
harbour help
harbour version
harbour provision
harbour shell
harbour agent
harbour yolo
```

## Config

Harbour stores its config as a single JSON file.

```json
{
  "colima_profile": "harbour",
  "colima_runtime": "docker",
  "colima_vm_type": "vz",
  "colima_arch": "aarch64",
  "colima_cpu": 4,
  "colima_memory": 8,
  "colima_disk": 100,
  "colima_mount_type": "virtiofs",
  "colima_forward_ssh_agent": true,
  "colima_network_address": false,
  "codex_version": "latest",
  "claude_code_version": "latest",
  "harness_path": "",
  "workspace_root": "",
  "active_agent": "",
  "default_command": "agent"
}
```

## Development

```sh
make build
./bin/harbour help
```

`make build` builds a local macOS ARM64 binary for development use.

```sh
go test ./...
```

## Releasing

```sh
make dist VERSION=v0.1.0
```

This writes Homebrew-ready Darwin artefacts to `dist/`:

- `harbour-v0.1.0-darwin-amd64.tar.gz`
- `harbour-v0.1.0-darwin-arm64.tar.gz`
- `sha256sums.txt`

`make dist` verifies the requested tag on `origin`, clones that tag into a temporary release source checkout under `build/`, and builds the release artefacts from that remote tagged source.

Release builds inject the requested version into `harbour version`.

`VERSION` must match `vX.Y.Z`.
