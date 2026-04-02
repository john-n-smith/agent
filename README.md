# Harbour

Run agents across all your repos in an isolated, shareable sandbox.

- Share a simple harness (`repos.yaml`, `AGENTS.md`, `skills/`)
- Run agents in an isolated Colima VM
- Run across multiple repositories in a single run
- Keep your existing Docker workflow unchanged
- Choose your agent (Claude or Codex) at provisioning time

## Install

```sh
brew tap agent-harbour/harbour
brew install harbour
```

## Quick start

1. Create your harness
   
   - `repos.yaml` - paths to your repos
   - `AGENTS.md` - your agent instructions. Will be symlinked to `CLAUDE.md` on provision, if you're using Claude
   - `skills/` - agent skills

   See https://github.com/agent-harbour/harbour-harness-template for an example.

   Relatvie paths in `repos.yaml` are resolved from `HARBOUR_WORKSPACE_ROOT`

2. Run provision

   ```sh
   harbour provision
   ```

   `harbour provision` will create `~/.config/harbour/env` from `config/harbour.env.example` if it does not exist, then prompt for

   - `HARBOUR_HARNESS_PATH` - the path to your harness
   - `HARBOUR_WORKSPACE_ROOT` = where you keep your repos. Accepts `~`
   - Claude or Codex

   Run `harbour provision` again after changing `repos.yaml`, `AGENTS.md` or `skills/`

3. Choose the default `harbour` command during provision

   `harbour provision` will prompt for `HARBOUR_DEFAULT_COMMAND`

   - `agent` - launch the provisioned agent
   - `yolo` - launch the provisioned agent with relaxed permissions
   - `shell` - open a shell in the Harbour VM

4. Start with your default command

   ```sh
   harbour
   ```

   Start the agent explicitly with

   ```sh
   harbour agent
   ```

   Or start it with relaxed permissions using

   ```sh
   harbour yolo
   ```

## Provisioning

`harbour provision` will:

- Create `~/.config/harbour/env` from `config/harbour.env.example` if needed
- Start the configured Harbour profile in Colima
- Mount `harbour-harness`
- Mount the work repos from `harbour-harness/repos.yaml`
- Warn and skip any repo mount whose host directory does not exist
- Install or update only the selected agent plus shared tooling in the VM
- Remove the inactive agent from the VM
- Link `AGENTS.md` or `CLAUDE.md` at `HARBOUR_WORKSPACE_ROOT`
- Sync custom skills from `harbour-harness/skills` to the selected agent's skills directory
- Save `HARBOUR_DEFAULT_COMMAND` for plain `harbour` invocations

## Layout

- `harbour`: Stable entry point
- `config/harbour.env.example`: Example local Harbour config
- `scripts/`: Provisioning and launch scripts
- `docs/architecture.md`: Runtime model
- `docs/adr/`: Design decisions

## Usage

```sh
harbour help
harbour provision
harbour shell
harbour agent
harbour yolo
harbour
```
