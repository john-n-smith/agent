# Harbour

Run agents across all your repos in an isolated, shareable sandbox.

- Share a simple harness (`repos.yaml`, `AGENTS.md`, `skills/`)
- Run agents in an isolated Colima VM
- Run across multiple repositories in a single run
- Keep your existing Docker workflow unchanged
- Choose your agent (Claude or Codex) at provisioning time

## Quick start

1. Create your harness
   
   - `repos.yaml` - paths to your repos
   - `AGENTS.md` - your agent instructions. Will be symlinked to `CLAUDE.md` on provision, if you're using Claude
   - `skills/` - agent skills

   See https://github.com/agent-harbour/harbour-harness-template for an example.

   Relatvie paths in `repos.yaml` are resolved from `HARBOUR_WORKSPACE_ROOT`

2. Run provision

   ```sh
   make provision
   ```

   `make provision` will prompt for

   - `HARBOUR_HARNESS_PATH` - the path to your harness
   - `HARBOUR_WORKSPACE_ROOT` = where you keep your repos. Accepts `~`
   - Claude or Codex

   Run `make provision` again after changing `repos`, `AGENTS.md` or `skills/`

6. Start the agent

   ```sh
   make agent
   ```
   or
   
   ```
   make yolo
   ```

## Provisioning

`make provision` will:

- Start the Colima profile
- Mount `harbour-harness`
- Mount the work repos from `harbour-harness/repos.yaml`
- Warn and skip any repo mount whose host directory does not exist
- Install or update only the selected agent plus shared tooling in the VM
- Remove the inactive agent from the VM
- Link `AGENTS.md` or `CLAUDE.md` at `HARBOUR_WORKSPACE_ROOT`
- Sync custom skills from `harbour-harness/skills` to the selected agent's skills directory

## Layout

- `Makefile`: Stable entry points
- `config/colima.env`: Colima defaults
- `scripts/`: Provisioning and launch scripts
- `docs/architecture.md`: Runtime model
- `docs/adr/`: Design decisions

## Usage

```sh
make help
make provision
make shell
make agent
make yolo
```
