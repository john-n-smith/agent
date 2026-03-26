# Harbour

Harbour runs Claude or Codex in a Colima VM with controlled repo mounts and a versioned shared context

- Run the agent isolated from your host machine
- Work across multiple repos
- Keep skills, and policy in a versioned context repo
- Share a common agent setup across a team
- Use Docker normally through Colima
- Switch between Claude and Codex at provision time

## Quick start

1. Fork `harbour-context-skeleton`

   Clone your fork

   ```sh
   git clone --depth 1 git@github.com:my-user/my-harbour-context-fork.git
   ```

2. Add your repo mounts to `repos.yaml`

   Add the repo entries you want mounted into the VM
   Relative paths are resolved from `HARBOUR_WORKSPACE_ROOT`

3. Add skills and update `AGENTS.md` in your forked context

   Harbour passes both to the agent

4. Run provision

   ```sh
   make provision
   ```

   `make provision` will prompt for `HARBOUR_CONTEXT_HOST_PATH`, `HARBOUR_WORKSPACE_ROOT`, and `codex` or `claude` when needed. `HARBOUR_WORKSPACE_ROOT` accepts `~`

   Run `make provision` again after changing `AGENTS.md` or skills

5. Start the agent

   ```sh
   make yolo
   ```

## Provisioning

`make provision` will:

- Start the Colima profile
- Mount `harbour-context`
- Mount the work repos from `harbour-context/repos.yaml`
- Warn and skip any repo mount whose host directory does not exist
- Install or update only the selected agent plus shared tooling in the VM
- Remove the inactive agent from the VM
- Link `AGENTS.md` or `CLAUDE.md` at `HARBOUR_WORKSPACE_ROOT`
- Sync custom skills from `harbour-context/skills` to the selected agent's skills directory

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
