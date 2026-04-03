# Harbour

Run agents across your repos inside an isolated Colima VM.

- One Go CLI
- One JSON config file
- One harness repo for `repos.yaml`, `AGENTS.md`, and `skills/`
- Go source under `cmd/harbour/`

## Build

```sh
make build
./bin/harbour help
```

`make build` builds a local macOS ARM64 binary.

## Release

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

## Quick Start

1. Create your harness

   - `repos.yaml` lists repo mount paths
   - `AGENTS.md` contains shared instructions
   - `skills/` contains optional custom skills

   See https://github.com/agent-harbour/harbour-harness-template for an example.

   Relative `host_path` values in `repos.yaml` are resolved from `workspace_root`.

2. Provision Harbour

   ```sh
   ./bin/harbour provision
   ```

   The first run creates a config file at the platform config location for Harbour.
   On Linux this is typically `~/.config/harbour/config.json`.

   Provision prompts for:

   - `harness_path`
   - `workspace_root`
   - The active agent
   - The default `harbour` command

3. Run the agent

```sh
./bin/harbour
```

Or run a command explicitly:

```sh
./bin/harbour agent
./bin/harbour yolo
./bin/harbour shell
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

## Commands

```sh
./bin/harbour help
./bin/harbour version
./bin/harbour provision
./bin/harbour shell
./bin/harbour agent
./bin/harbour yolo
```
