# Harbour

Harbour runs Claude or Codex in a Colima VM with controlled repo mounts and a versioned shared context

- Run the agent outside your host machine
- Work across multiple repos in one session
- Keep prompts, skills, and policy in a versioned context repo
- Share a common agent setup across a team
- Use Docker normally through Colima
- Switch between Claude and Codex at provision time

## Quick start

1. Fork `harbour-context-skeleton`

   Clone your fork

   ```sh
   git clone --depth 1 git@github.com:agent-harbour/harbour-context-skeleton.git my-harbour-context
   ```

2. Add your repo mounts to `repos.yaml`

   Add the `host_path` entries you want mounted into the VM

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
- Install or update only the selected agent plus shared tooling in the VM
- Remove the inactive agent from the VM
- Link `AGENTS.md` or `CLAUDE.md` at `HARBOUR_WORKSPACE_ROOT`
- Sync custom skills from `harbour-context/skills` when Codex is selected

## Why Harbour

- The agent runs in a VM instead of on your host machine
- Repo access is explicit through `repos.yaml`
- The same context repo can carry prompts, policy, and skills
- Colima keeps Docker integration simple on the host
- Agent choice stays reversible between Claude and Codex

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

## Notes

The intended runtime is a Colima VM. Codex runs directly in the VM, alongside
repo containers, rather than inside its own container. Claude Code uses the
same VM model when selected during `make provision`.

Mount only the repo paths declared in `harbour-context/repos.yaml`. Anything
mounted into the VM is intentionally shared with the agent and repo containers.

Each entry in `repos.yaml` is a `host_path` and is mounted read-write.
